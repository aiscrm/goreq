package xml

import (
	"bytes"
	"encoding/xml"

	"github.com/aiscrm/goreq/codec"
)

type stdCodec struct {
	options codec.Options
}

func NewCodec(opts ...codec.Option) codec.Codec {
	options := codec.Options{}
	for _, o := range opts {
		o(&options)
	}
	return &stdCodec{options: options}
}

func (j stdCodec) Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	enc.Indent(j.options.IndentPrefix, j.options.IndentValue)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
	//return xml.Marshal(v)
}

func (j stdCodec) Unmarshal(data []byte, v interface{}) error {
	err := xml.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return nil
}

func (j stdCodec) Name() string {
	return "xml"
}
