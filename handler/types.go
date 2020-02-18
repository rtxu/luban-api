package handler

import (
	"encoding/json"
	"strings"
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
