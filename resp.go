package goreq

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/aiscrm/goreq/codec"
)

// Resp represents a http response
type Resp struct {
	request  *http.Request
	response *http.Response
	body     []byte
	err      error
	cost     time.Duration
	timeout  bool
	codecs   codec.Codecs
}

// Request returns *http.Request
func (r *Resp) Request() *http.Request {
	return r.request
}

// Response returns *http.Response
func (r *Resp) Response() *http.Response {
	return r.response
}

// StatusCode returns status code
func (r *Resp) StatusCode() int {
	return r.Response().StatusCode
}

// ContentLength returns content length
func (r *Resp) ContentLength() int64 {
	return r.Response().ContentLength
}

// ContentType returns content type
func (r *Resp) ContentType() string {
	return r.Response().Header.Get(ContentType)
}

// Timeout returns true if timeout
func (r *Resp) Timeout() bool {
	return r.timeout
}

// Cost returns cost time
func (r *Resp) Cost() time.Duration {
	return r.cost
}

// Error get error
func (r *Resp) Error() error {
	return r.err
}

func (r *Resp) SetError(err error) {
	r.err = err
}

// Consume close response body
func (r *Resp) Consume(read bool) {
	if read {
		_, _ = r.AsBytes()
	} else if r.body == nil {
		r.response.Body.Close()
		r.body = []byte{}
	}
}

// Bytes returns response body as []byte
func (r *Resp) Bytes() []byte {
	data, _ := r.AsBytes()
	return data
}

// AsBytes returns response body as []byte,
// return error if error happend when reading
// the response body
func (r *Resp) AsBytes() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.body != nil {
		return r.body, nil
	}
	defer r.response.Body.Close()
	r.body, r.err = ioutil.ReadAll(r.response.Body)
	return r.body, r.err
}

// AsReader returns response body as reader
func (r *Resp) AsReader() (io.Reader, error) {
	data, err := r.AsBytes()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

// String returns response body as string
func (r *Resp) String() string {
	data, _ := r.AsBytes()
	return string(data)
}

// AsString returns response body as string,
// return error if error happend when reading
// the response body
func (r *Resp) AsString() (string, error) {
	data, err := r.AsBytes()
	return string(data), err
}

// AsStruct convert to struct. default to use json format
func (r *Resp) AsStruct(v interface{}, unmarshal func([]byte, interface{}) error) error {
	data, err := r.AsBytes()
	if err != nil {
		return err
	}
	return unmarshal(data, v)
}

// AsJSONStruct convert json response body to struct or map
func (r *Resp) AsJSONStruct(v interface{}) error {
	return r.AsStruct(v, r.codecs.Get(codec.JSONCodec).Unmarshal)
}

// AsXMLStruct convert xml response body to struct or map
func (r *Resp) AsXMLStruct(v interface{}) error {
	return r.AsStruct(v, r.codecs.Get(codec.XMLCodec).Unmarshal)
}

func (r *Resp) AsJSONMap() (map[string]interface{}, error) {
	var m map[string]interface{}
	err := r.AsJSONStruct(&m)
	return m, err
}

func (r *Resp) AsXMLMap() (map[string]interface{}, error) {
	var m map[string]interface{}
	err := r.AsXMLStruct(&m)
	return m, err
}

// AsFile save to file
func (r *Resp) AsFile(dest string) error {
	data, err := r.AsBytes()
	if err != nil {
		return err
	}
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
	// if r.body != nil {
	// 	_, err = file.Write(r.body)
	// 	return err
	// }

	// defer r.response.Body.Close()
	// _, err = io.Copy(file, r.response.Body)
	// return err
}

func (r *Resp) Dump() string {
	buf := &bytes.Buffer{}
	if r.request != nil {
		reqData, _ := httputil.DumpRequest(r.request, true)
		buf.Write(reqData)
	}
	buf.WriteString("\n\n")
	if r.response != nil {
		respData, _ := httputil.DumpResponse(r.response, true)
		buf.Write(respData)
	}
	return buf.String()
}
