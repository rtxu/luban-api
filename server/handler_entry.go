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

func (s *server) getCurrentUserAndRootDir(r *http.Request) (db.User, DirectoryT) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	var userID = claims["user_id"].(string)
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

func (s *server) syncUpdatedRootDirToDB(username string, rootDir DirectoryT) {
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
			panic(err)
		}

		user, rootDir := s.getCurrentUserAndRootDir(r)
		pTargetDir := findTargetDir(param.Dir, &rootDir)

		for i := 0; i < len(*pTargetDir); i++ {
			if (*pTargetDir)[i].Name == param.Entry.Name {
				s.respond(w, r, defaultResponse{
					Code: 1,
					Msg: fmt.Sprintf("Entry(%s) already exists under directory(%s)",
						param.Entry.Name,
						param.Dir),
				}, http.StatusOK)
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
		s.syncUpdatedRootDirToDB(user.UserName, rootDir)
		s.respond(w, r, defaultResponse{}, http.StatusOK)
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
			panic(err)
		}

		user, rootDir := s.getCurrentUserAndRootDir(r)
		pTargetDir := findTargetDir(param.Dir, &rootDir)

		newTargetDir := make(DirectoryT, 0, len((*pTargetDir))-1)
		for i := 0; i < len((*pTargetDir)); i++ {
			entry := (*pTargetDir)[i]
			if entry.Name == param.EntryName {
				if entry.Type == Directory {
					if len(entry.Children) > 0 {
						s.respond(w, r, defaultResponse{
							Code: 1,
							Msg:  "non-empty directory",
						}, http.StatusOK)
						return
					}
				} else {
				}
			} else {
				newTargetDir = append(newTargetDir, entry)
			}
		}
		(*pTargetDir) = newTargetDir
		s.syncUpdatedRootDirToDB(user.UserName, rootDir)
		s.respond(w, r, defaultResponse{}, http.StatusOK)
	}
}
