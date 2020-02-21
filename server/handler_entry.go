package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/jwtauth"

	"github.com/rtxu/luban-api/db"
)

type EntryTypeT int

const (
	Unknown EntryTypeT = iota
	App
	Directory
)

func (t *EntryTypeT) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	default:
		*t = Unknown
	case "app":
		*t = App
	case "directory":
		*t = Directory
	}

	return nil
}

func (a EntryTypeT) MarshalJSON() ([]byte, error) {
	var s string
	switch a {
	default:
		s = "unknown"
	case App:
		s = "app"
	case Directory:
		s = "directory"
	}

	return json.Marshal(s)
}

// EntryT 代表用户导航菜单中的一项
type EntryT struct {
	Name    string     `json:"name"`
	Type    EntryTypeT `json:"type"`
	Comment string     `json:"comment"`
	Icon    string     `json:"icon"`

	// when Type="app"
	AppId uint32 `json:"appId"`

	// when Type="directory"
	Children DirectoryT `json:"children"`
}

type DirectoryT []*EntryT

func (s *server) getCurrentUserAndRootDirFromDB(userID string) (db.User, DirectoryT) {
	user, err := s.userService.Find(userID)
	if err != nil {
		panic(err)
	}
	var rootDir DirectoryT
	if user.RootDir != nil {
		if err := json.NewDecoder(strings.NewReader(*user.RootDir)).Decode(&rootDir); err != nil {
			panic(err)
		}
	}
	return user, rootDir
}

func (s *server) getCurrentUserAndRootDir(r *http.Request) (db.User, DirectoryT) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	var userID = claims["user_id"].(string)
	return s.getCurrentUserAndRootDirFromDB(userID)
}

func findDir(targetDirName string, rootDir *DirectoryT) (*DirectoryT, error) {
	if targetDirName == "/" {
		return rootDir, nil
	}
	dirName := targetDirName[1 : len(targetDirName)-1]
	dirParts := strings.Split(dirName, "/")
	currentDirPtr := rootDir
	for i := 0; i < len(dirParts); i++ {
		targetDirName := dirParts[i]
		found := false
		for j := 0; j < len(*currentDirPtr); j++ {
			entry := (*currentDirPtr)[j]
			if entry.Type == Directory && entry.Name == targetDirName {
				found = true
				currentDirPtr = &entry.Children
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("%w: dir(%s) not found", errEntryNotFound, targetDirName)
		}
	}
	return currentDirPtr, nil
}

func validate(dir, entryName string) error {
	if entryName == "" {
		return fmt.Errorf("%w: empty entry name", errInvalidParam)
	}
	if strings.Contains(entryName, "/") {
		return fmt.Errorf("%w: '/' is illegal in entry name", errInvalidParam)
	}
	if dir == "" {
		return fmt.Errorf("%w: empty dirname", errInvalidParam)
	}
	length := len(dir)
	if dir[0] != '/' || dir[length-1] != '/' {
		return fmt.Errorf("%w: dir(%s) should begin and end with '/'",
			errInvalidParam, dir)
	}
	return nil
}

func (s *server) syncRootDirToDB(username string, rootDir DirectoryT) {
	var buffer strings.Builder
	json.NewEncoder(&buffer).Encode(rootDir)
	err := s.userService.Update(username, map[string]interface{}{
		"root_dir": buffer.String(),
	})
	if err != nil {
		panic(err)
	}
}

func (s *server) handleEntryCreate() http.HandlerFunc {
	type request struct {
		Dir   string `json:"dir"`
		Entry EntryT `json:"entry"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var param request
		if err := s.decode(w, r, &param); err != nil {
			s.respond(w, r, fmt.Errorf("%w: %v", errJsonDecode, err), http.StatusOK)
			return
		}
		if err := validate(param.Dir, param.Entry.Name); err != nil {
			s.respond(w, r, err, http.StatusOK)
			return
		}

		user, rootDir := s.getCurrentUserAndRootDir(r)
		pTargetDir, err := findDir(param.Dir, &rootDir)
		if err != nil {
			s.respond(w, r, err, http.StatusOK)
			return
		}

		for i := 0; i < len(*pTargetDir); i++ {
			if (*pTargetDir)[i].Name == param.Entry.Name {
				s.respond(w, r,
					fmt.Errorf("%w: %s%s", errEntryAlreadyExist,
						param.Dir, param.Entry.Name),
					http.StatusOK)
				return
			}
		}

		if param.Entry.Type == Directory {
			param.Entry.Children = make(DirectoryT, 0)
		} else {
			app := &db.App{
				OwnerID: user.ID,
			}
			err := s.appService.NewApp(app)
			if err != nil {
				panic(err)
			}
			param.Entry.AppId = app.ID
		}
		(*pTargetDir) = append((*pTargetDir), &param.Entry)
		s.syncRootDirToDB(user.UserName, rootDir)
		s.respond(w, r, success, http.StatusOK)
	}
}

func (s *server) handleEntryDelete() http.HandlerFunc {
	type request struct {
		Dir       string `json:"dir"`
		EntryName string `json:"entryName"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var param request
		if err := s.decode(w, r, &param); err != nil {
			s.respond(w, r, fmt.Errorf("%w: %v", errJsonDecode, err), http.StatusOK)
			return
		}
		if err := validate(param.Dir, param.EntryName); err != nil {
			s.respond(w, r, err, http.StatusOK)
			return
		}

		user, rootDir := s.getCurrentUserAndRootDir(r)
		pTargetDir, err := findDir(param.Dir, &rootDir)
		if err != nil {
			s.respond(w, r, err, http.StatusOK)
			return
		}

		newTargetDir := make(DirectoryT, 0, len((*pTargetDir))-1)
		for i := 0; i < len((*pTargetDir)); i++ {
			entry := (*pTargetDir)[i]
			if entry.Name == param.EntryName {
				if entry.Type == Directory && len(entry.Children) > 0 {
					s.respond(w, r,
						fmt.Errorf("%w: %s", errDirNotEmpty, param.Dir),
						http.StatusOK)
					return
				}
			} else {
				newTargetDir = append(newTargetDir, entry)
			}
		}
		(*pTargetDir) = newTargetDir
		s.syncRootDirToDB(user.UserName, rootDir)
		s.respond(w, r, success, http.StatusOK)
	}
}
