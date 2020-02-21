package server

import (
	"errors"
)

var (
	errBadRequest    = errors.New("bad request")
	errInvalidParam  = errors.New("invalid param")
	errEntryNotFound = errors.New("entry not found")
	errEntryAlreadyExist = errors.New("entry already exist")
	errDirNotEmpty = errors.New("dir not empty")
	errJsonDecode = errors.New("json decode failed")
)

var errCodeMap = map[error]int{
	errBadRequest:    1,
	errInvalidParam:  2,
	errEntryNotFound: 3,
	errEntryAlreadyExist: 4, 
	errJsonDecode: 5,
}
