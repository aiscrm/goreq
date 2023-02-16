package json

import (
	"bytes"
	"encoding"
	"encoding/json"

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
	switch vv := v.(type) {
	case []byte:
		return vv, nil
	case string:
		return []byte(vv), nil
	case json.Marshaler:
		return vv.MarshalJSON()
	case encoding.BinaryMarshaler:
		return vv.MarshalBinary()
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent(j.options.IndentPrefix, j.options.IndentValue)
	enc.SetEscapeHTML(j.options.EscapeHTML)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (j stdCodec) Unmarshal(data []byte, v interface{}) error {
	var err error
	switch vv := v.(type) {
	case json.Unmarshaler:
		err = vv.UnmarshalJSON(data)
	case encoding.BinaryUnmarshaler:
		err = vv.UnmarshalBinary(data)
	default:
		err = json.Unmarshal(data, v)
	}
	if err != nil {
		return err
	}
	return nil
}

func (j stdCodec) Name() string {
	return "json"
}
