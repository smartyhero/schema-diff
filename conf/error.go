package conf

import "errors"

var (
	ErrConfigSrcMiss       = errors.New("ErrConfigSrcMiss")
	ErrConfigDstMiss       = errors.New("ErrConfigDstMiss")
	ErrUnknownSchemaSource = errors.New("ErrUnknownSchemaSource")
)
