package db

import "errors"

var (
	ErrNoSuchTable        = errors.New("NoSuchTable")
	ErrSrcDstSchemaIsNull = errors.New("ErrSrcDstSchemaIsNull")
)
