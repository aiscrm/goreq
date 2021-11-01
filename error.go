package goreq

import "errors"

// define errors
var (
	ErrNoTransport      = errors.New("req: no transport")
	ErrNoURL            = errors.New("req: url not specified")
	ErrLackParam        = errors.New("req: lack param")
	ErrNoParser         = errors.New("resp: no parser")
	ErrNoFileMatch      = errors.New("req: no file match")
	ErrNotSupportedBody = errors.New("req: not supported body")
	ErrNoUnmarshal      = errors.New("resp: no unmarshal")
	ErrNoMarshal        = errors.New("req: no marshal")
	ErrParseStruct      = errors.New("req: can not parse struct param")
)
