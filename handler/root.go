package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"

	"github.com/rtxu/luban-api/config"
	"github.com/rtxu/luban-api/db"
)

var TokenAuth *jwtauth.JWTAuth = jwtauth.New("HS256", []byte(config.JWTSecret), nil)

func GithubLogin(w http.ResponseWriter, r *http.Request) {
	githubCode := r.URL.Query().Get("code")
	params := make(url.Values)
	params.Add("code", githubCode)
	params.Add("client_id", config.GithubOAuth.ClientID)
	params.Add("client_secret", config.GithubOAuth.ClientSecret)
	req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token?"+params.Encode(), nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	type accessTokenT struct {
		TokenType   string `json:"token_type"`
		AccessToken string `json:"access_token"`
	}
	var accessToken accessTokenT
	if err := json.Unmarshal(body, &accessToken); err != nil {
		panic(err)
	}

	req2, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req2.Header.Add("Authorization", fmt.Sprintf("%s %s", accessToken.TokenType, accessToken.AccessToken))
	resp2, err := client.Do(req2)
	if err != nil {
		panic(err)
	}
	defer resp2.Body.Close()
	body2, err := ioutil.ReadAll(resp2.Body)
	type userInfoT struct {
		Login     string `json:"login"`
		AvatarUrl string `json:"avatar_url"`
	}
	var userInfo userInfoT
	if err := json.Unmarshal(body2, &userInfo); err != nil {
		panic(err)
	}

	_, err = db.UserService.FindByGithubUserName(userInfo.Login)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			err := db.UserService.Insert(db.User{
				UserName:       userInfo.Login,
				GithubUserName: &userInfo.Login,
				AvatarUrl:      &userInfo.AvatarUrl,
			})
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	claims := jwt.MapClaims{
		"user_id": userInfo.Login,
	}
	jwtauth.SetExpiryIn(claims, 7*24*time.Hour)
	_, tokenString, _ := TokenAuth.Encode(claims)
	query := url.Values{
		"access_token": {fmt.Sprintf("Bearer %s", tokenString)},
	}.Encode()

	http.Redirect(w, r,
		fmt.Sprintf("%s/login-success?%s", config.AppRoot, query),
		303)
}

func readParam(r io.Reader, param interface{}) {
	bytes, _ := ioutil.ReadAll(r)
	json.Unmarshal(bytes, &param)
}

func getCurrentUserAndRootDir(r *http.Request) (db.User, DirectoryT) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	var userID = claims["user_id"].(string)
	user, err := db.UserService.Find(userID)
	if err != nil {
		panic(err)
	}
	var rootDir DirectoryT
	if user.RootDir != nil {
		if err := json.Unmarshal([]byte(*user.RootDir), &rootDir); err != nil {
			panic(err)
		}
	}
	return user, rootDir
}

func findTargetDir(targetDirName string, rootDir *DirectoryT) *DirectoryT {
	currentDirPtr := rootDir
	dirParts := strings.FieldsFunc(targetDirName, func(r rune) bool {
		return r == '/'
	})
	for i := 0; i < len(dirParts); i++ {
		targetDirName := dirParts[i]
		for j := 0; j < len(*currentDirPtr); j++ {
			entry := (*currentDirPtr)[j]
			if entry.Type == Directory && entry.Name == targetDirName {
				currentDirPtr = &entry.Children
				break
			}
		}
	}
	return currentDirPtr
}

func syncUpdatedRootDirToDB(username string, rootDir DirectoryT) {
	newRootDirBytes, _ := json.Marshal(rootDir)
	err := db.UserService.Update(username, map[string]interface{}{
		"root_dir": string(newRootDirBytes),
	})
	if err != nil {
		panic(err)
	}
}

func AddEntry(w http.ResponseWriter, r *http.Request) {
	var param struct {
		Dir   string `json:"dir"`
		Entry EntryT `json:"entry"`
	}
	readParam(r.Body, &param)

	user, rootDir := getCurrentUserAndRootDir(r)
	pTargetDir := findTargetDir(param.Dir, &rootDir)

	for i := 0; i < len(*pTargetDir); i++ {
		if (*pTargetDir)[i].Name == param.Entry.Name {
			writeJsonResponse(w, JsonResponse{
				Code: 1,
				Msg: fmt.Sprintf("Entry(%s) already exists under directory(%s)",
					param.Entry.Name,
					param.Dir),
			})
			return
		}
	}

	if param.Entry.Type == Directory {
		param.Entry.Children = make(DirectoryT, 0)
	} else {
		app := &db.App{
			OwnerID: user.ID,
		}
		err := db.AppService.NewApp(app)
		if err != nil {
			panic(err)
		}
		param.Entry.AppId = app.ID
	}
	(*pTargetDir) = append((*pTargetDir), &param.Entry)
	syncUpdatedRootDirToDB(user.UserName, rootDir)
	writeJsonData(w, nil)
}

func DeleteEntry(w http.ResponseWriter, r *http.Request) {
	var param struct {
		Dir       string `json:"dir"`
		EntryName string `json:"entryName"`
	}
	readParam(r.Body, &param)

	user, rootDir := getCurrentUserAndRootDir(r)
	pTargetDir := findTargetDir(param.Dir, &rootDir)

	newTargetDir := make(DirectoryT, 0, len((*pTargetDir))-1)
	for i := 0; i < len((*pTargetDir)); i++ {
		entry := (*pTargetDir)[i]
		if entry.Name == param.EntryName {
			if entry.Type == Directory {
				if len(entry.Children) > 0 {
					writeJsonResponse(w, JsonResponse{
						Code: 1,
						Msg:  "non-empty directory",
					})
					return
				}
			} else {
			}
		} else {
			newTargetDir = append(newTargetDir, entry)
		}
	}
	(*pTargetDir) = newTargetDir
	syncUpdatedRootDirToDB(user.UserName, rootDir)
	writeJsonData(w, nil)
}

func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	var userID = claims["user_id"].(string)
	user, err := db.UserService.Find(userID)
	if err != nil {
		panic(err)
	}
	type DataT struct {
		Username  string     `json:"username"`
		AvatarUrl string     `json:"avatarUrl"`
		RootDir   DirectoryT `json:"rootDir"`
	}
	data := DataT{
		Username:  user.UserName,
		AvatarUrl: *user.AvatarUrl,
	}
	if user.RootDir != nil {
		err := json.Unmarshal([]byte(*user.RootDir), &data.RootDir)
		if err != nil {
			panic(err)
		}
	}
	writeJsonData(w, data)
}
