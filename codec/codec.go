package codec

import (
	"errors"
)

const (
	JSONCodec = "json"
	XMLCodec  = "xml"
)

var (
	ErrNoMarshal   = errors.New("no code")
	ErrNoUnmarshal = errors.New("no unmarshal")
)

type Codecs map[string]Codec

func (cs Codecs) Set(name string, codec Codec) {
	cs[name] = codec
}
func (cs Codecs) Get(name string) Codec {
	if c, ok := cs[name]; ok {
		return c
	}
	return nil
}

type Codec interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
	Name() string
}
