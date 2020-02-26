package server

import (
	"errors"
)

var (
	// client-side error, maybe triggered by frontend engineer under development
	// 在程序无 bug 的情况下，大概率不会触发
	errJsonDecode    = errors.New("json decode failed")
	errBadRequest    = errors.New("bad request")
	errInvalidParam  = errors.New("invalid param")
	errEntryNotFound = errors.New("entry not found")

	// user-side error, maybe triggered by end user
	errEntryAlreadyExist = errors.New("entry already exist")
	errDirNotEmpty       = errors.New("dir not empty")

	// server-side error, just panic
)

var errCodeMap = map[error]int{
	errJsonDecode:    100,
	errBadRequest:    101,
	errInvalidParam:  102,
	errEntryNotFound: 103,

	errEntryAlreadyExist: 200,
	errDirNotEmpty:       201,
}
