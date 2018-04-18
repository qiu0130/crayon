package crayon

import "errors"

var (
	ErrInvalidPort = errors.New("port number is invali (should be between 0-66535)")
	ErrInvalidFilePath = errors.New("file path not exist or is invalid")

	errMultiHeaderWrite = `http: multiple response.WriteHeader calls`
	errMultiWrite       = `http: multiple response.Write calls`
	errDuplicateKey     = `error: Duplicate URI keys found`
)
