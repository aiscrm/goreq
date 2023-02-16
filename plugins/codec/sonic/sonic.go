package sonic

import (
	"encoding"
	"encoding/json"

	"github.com/aiscrm/goreq/codec"
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/encoder"
)

type sonicCodec struct {
	options codec.Options
	enc     encoder.Encoder
}

func NewCodec(opts ...codec.Option) codec.Codec {
	options := codec.Options{}
	for _, o := range opts {
		o(&options)
	}
	enc := encoder.Encoder{}
	enc.SetIndent(options.IndentPrefix, options.IndentValue)
	enc.SetEscapeHTML(options.EscapeHTML)
	return &sonicCodec{
		options: options,
		enc:     enc,
	}
}

func (j sonicCodec) Marshal(v interface{}) ([]byte, error) {
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
	return j.enc.Encode(v)
}

func (j sonicCodec) Unmarshal(data []byte, v interface{}) error {
	var err error
	switch vv := v.(type) {
	case json.Unmarshaler:
		err = vv.UnmarshalJSON(data)
	case encoding.BinaryUnmarshaler:
		err = vv.UnmarshalBinary(data)
	default:
		err = sonic.Unmarshal(data, v)
	}
	if err != nil {
		return err
	}
	return nil
}

func (j sonicCodec) Name() string {
	return "json"
}
