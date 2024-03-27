package goreq

import (
	"bytes"
	"context"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aiscrm/goreq/codec"
)

// content type
const (
	ContentType     = "Content-Type"
	ContentTypeJSON = "application/json; charset=UTF-8"
	ContentTypeXML  = "application/xml; charset=UTF-8"
	ContentTypeForm = "application/x-www-form-urlencoded; charset=UTF-8"

	Accept                 = "Accept"
	UserAgent              = "User-Agent"
	Referer                = "Referer"
	Origin                 = "Origin"
	ContentEncoding        = "Content-Encoding"
	ContentEncodingGzip    = "gzip"
	ContentEncodingDeflate = "deflate"
)

// Req represents a http request
type Req struct {
	client      Client
	name        string
	rawURL      string
	method      string
	queryParams url.Values
	formParams  url.Values
	err         error
	uploads     []FileUpload
	header      http.Header
	cookies     []*http.Cookie
	ctx         context.Context
	body        []byte
	lazyBody    interface{} // 仅将内容原封不动的保存在Req中，交由Handler对lazyBody处理后在转换为实际的Request中的body
}

// FileUpload represents a file to upload
type FileUpload struct {
	FieldName string        // form field name
	FileName  string        // filename in multipart form.
	File      io.ReadCloser // file to upload, required
}

// New return an empty request
func New() *Req {
	return &Req{
		client:      DefaultClient,
		method:      http.MethodGet,
		queryParams: make(url.Values),
		formParams:  make(url.Values),
		uploads:     []FileUpload{},
		header:      make(http.Header),
		cookies:     []*http.Cookie{},
		ctx:         context.TODO(),
		body:        []byte{},
		err:         nil,
	}
}

// Use 仅在当前请求范围生效的中间件，不会改变到DefaultClient
func (r *Req) Use(handlers ...HandlerFunc) *Req {
	c := r.client.Clone()
	c.Use(handlers...)
	r.WithClient(c)
	return r
}

// WithName to identify this request, for trace mostly.
func (r *Req) WithName(name string) *Req {
	r.name = name
	return r
}

// WithURL set request raw url
func (r *Req) WithURL(rawURL string) *Req {
	r.rawURL = rawURL
	return r
}

// WithMethod set request method
func (r *Req) WithMethod(method string) *Req {
	r.method = method
	return r
}

// WithHeader set request method
func (r *Req) WithHeader(key, value string) *Req {
	r.header.Set(key, value)
	return r
}

// WithBasicAuth sets the request's Authorization header to use HTTP
func (r *Req) WithBasicAuth(username, password string) *Req {
	auth := username + ":" + password
	return r.WithHeader("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
}

// WithContext with context
func (r *Req) WithContext(ctx context.Context) *Req {
	r.ctx = ctx
	return r
}

// WithClient with client
func (r *Req) WithClient(c Client) *Req {
	r.client = c
	return r
}

// WithAccept is to set Accept header
func (r *Req) WithAccept(contentType string) *Req {
	return r.WithHeader(Accept, contentType)
}

// WithContentType is to set Content-Type header
func (r *Req) WithContentType(contentType string) *Req {
	return r.WithHeader(ContentType, contentType)
}

// WithUserAgent is to set User-Agent header
func (r *Req) WithUserAgent(userAgent string) *Req {
	if userAgent == "" {
		// r.request.header.Del(UserAgent)
		r.header.Del(UserAgent)
		return r
	}
	return r.WithHeader(UserAgent, userAgent)
}

// WithReferer is to set Referer header
func (r *Req) WithReferer(referer string) *Req {
	if referer == "" {
		r.header.Del(Referer)
		return r
	}
	return r.WithHeader(Referer, referer)
}

// WithOrigin is to set Origin header
func (r *Req) WithOrigin(origin string) *Req {
	if origin == "" {
		r.header.Del(Origin)
		return r
	}
	return r.WithHeader(Origin, origin)
}

func (r *Req) WithStreamHeaders() *Req {
	return r.WithHeader("Accept", "text/event-stream; charset=utf-8").
		WithHeader("Cache-Control", "no-cache").
		WithHeader("Connection", "keep-alive")
}

// WithLazyBody 仅将内容原封不动的保存在Req中，交由Handler对lazyBody处理后在转换为实际的Request中的body
func (r *Req) WithLazyBody(lazyBody interface{}) *Req {
	r.lazyBody = lazyBody
	return r
}

// WithBody add body
func (r *Req) WithBody(body interface{}) *Req {
	switch b := body.(type) {
	case json.Marshaler:
		data, err := b.MarshalJSON()
		if err != nil {
			r.err = err
			return r
		}
		return r.WithBinaryBody(data)
	case encoding.BinaryMarshaler:
		data, err := b.MarshalBinary()
		if err != nil {
			r.err = err
			return r
		}
		return r.WithBinaryBody(data)
	case io.ReadCloser:
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(b); err != nil {
			r.err = err
			return r
		}
		if err := b.Close(); err != nil {
			r.err = err
			return r
		}
		return r.WithBinaryBody(buf.Bytes())
	case io.Reader:
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(b); err != nil {
			r.err = err
			return r
		}
		return r.WithBinaryBody(buf.Bytes())
	case bytes.Buffer:
		return r.WithBinaryBody(b.Bytes())
	case string:
		return r.WithBinaryBody([]byte(b))
	case []byte:
		return r.WithBinaryBody(b)
	case func() ([]byte, error):
		data, err := b()
		if err != nil {
			r.err = err
			return r
		}
		return r.WithBinaryBody(data)
	default:
		r.err = ErrNotSupportedBody
		return r
	}
}

// WithBinaryBody add binary body
func (r *Req) WithBinaryBody(body []byte) *Req {
	if len(body) == 0 {
		return r
	}
	r.body = body
	return r
}

// WithJSONBody convert body to json data
func (r *Req) WithJSONBody(body interface{}) *Req {
	r.WithContentType(ContentTypeJSON)
	// data, err := json.Marshal(body)
	data, err := r.client.Options().Codecs.Get(codec.JSONCodec).Marshal(body)
	if err != nil {
		r.err = err
		return r
	}
	return r.WithBody(data)
}

// WithXMLBody convert body to xml data
func (r *Req) WithXMLBody(body interface{}) *Req {
	r.WithContentType(ContentTypeXML)
	// data, err := xml.Marshal(body)
	data, err := r.client.Options().Codecs.Get(codec.XMLCodec).Marshal(body)
	if err != nil {
		r.err = err
		return r
	}
	return r.WithBody(data)
}

// WithQueryParam with query parameter
func (r *Req) WithQueryParam(key string, value interface{}) *Req {
	r.queryParams.Set(key, toString(value))
	return r
}

// AddQueryParam add query parameter
func (r *Req) AddQueryParam(key string, value interface{}) *Req {
	r.queryParams.Add(key, toString(value))
	return r
}

// WithQueryParams with multi query parameters
func (r *Req) WithQueryParams(params map[string]interface{}) *Req {
	for k, v := range params {
		r.queryParams.Set(k, toString(v))
	}
	return r
}

// AddQueryParams add multi query parameters
func (r *Req) AddQueryParams(params map[string]interface{}) *Req {
	for k, v := range params {
		r.queryParams.Add(k, toString(v))
	}
	return r
}

// WithFormParam with form parameter
func (r *Req) WithFormParam(key string, value interface{}) *Req {
	r.formParams.Set(key, toString(value))
	return r
}

// AddFormParam add form parameter
func (r *Req) AddFormParam(key string, value interface{}) *Req {
	r.formParams.Add(key, toString(value))
	return r
}

// WithFormParams with multi form parameters
func (r *Req) WithFormParams(params map[string]interface{}) *Req {
	for k, v := range params {
		r.formParams.Set(k, toString(v))
	}
	return r
}

// AddFormParams add multi form parameters
func (r *Req) AddFormParams(params map[string]interface{}) *Req {
	for k, v := range params {
		r.formParams.Add(k, toString(v))
	}
	return r
}

// AddHeader add header
func (r *Req) AddHeader(key, value string) *Req {
	r.header.Add(key, value)
	return r
}

// AddHeaders add multi headers
func (r *Req) AddHeaders(headers map[string]string) *Req {
	for key, value := range headers {
		r.header.Add(key, value)
	}
	return r
}

// AddCookie adds a cookie to the request.
func (r *Req) AddCookie(c *http.Cookie) *Req {
	r.cookies = append(r.cookies, c)
	return r
}

// AddFiles upload file
func (r *Req) AddFiles(patterns ...string) *Req {
	var matches []string
	for _, pattern := range patterns {
		m, err := filepath.Glob(pattern)
		if err != nil {
			r.err = err
			return r
		}
		matches = append(matches, m...)
	}
	if len(matches) == 0 {
		r.err = ErrNoFileMatch
		return r
	}
	for _, match := range matches {
		if s, e := os.Stat(match); e != nil || s.IsDir() {
			continue
		}
		file, _ := os.Open(match)
		r.uploads = append(r.uploads, FileUpload{
			File:      file,
			FileName:  filepath.Base(match),
			FieldName: "media",
		})
	}
	return r
}

// AddFile upload file with custom field name and file name
func (r *Req) AddFile(fieldName, fileName string, file io.ReadCloser) *Req {
	r.uploads = append(r.uploads, FileUpload{
		FieldName: fieldName,
		FileName:  fileName,
		File:      file,
	})
	return r
}

// AddFileContent upload file with custom field name and file name
func (r *Req) AddFileContent(fieldName, fileName string, content []byte) *Req {
	r.uploads = append(r.uploads, FileUpload{
		FieldName: fieldName,
		FileName:  fileName,
		File:      io.NopCloser(bytes.NewReader(content)),
	})
	return r
}

// GetName return name for this request
func (r *Req) GetName() string {
	return r.name
}

// GetURL return request raw url
func (r *Req) GetURL() string {
	return r.rawURL
}

// GetHost return request host
func (r *Req) GetHost() string {
	u, err := url.Parse(r.GetURL())
	if err != nil {
		return ""
	}
	return u.Host
}

// GetPath return request path
func (r *Req) GetPath() string {
	u, err := url.Parse(r.GetURL())
	if err != nil {
		return ""
	}
	return u.Path
}

// GetMethod return request method
func (r *Req) GetMethod() string {
	return r.method
}

// GetQueryParams return request query params
func (r *Req) GetQueryParams() url.Values {
	return r.queryParams
}

// GetFormParams return request form params
func (r *Req) GetFormParams() url.Values {
	return r.formParams
}

// GetHeader return request headers
func (r *Req) GetHeader() http.Header {
	return r.header
}

// GetLazyBody return lazy body
func (r *Req) GetLazyBody() interface{} {
	return r.lazyBody
}

// GetBody return request body
// warn: file content can't be got when has upload files.
func (r *Req) GetBody() []byte {
	return r.body
}

// Context return request context
func (r *Req) Context() context.Context {
	if r.ctx == nil {
		return context.TODO()
	}
	return r.ctx
}

// GetClient return client
func (r *Req) GetClient() Client {
	return r.client
}

func (r *Req) Error() error {
	return r.err
}

// Build request
func (r *Req) Build() (*http.Request, error) {
	request := &http.Request{
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Method:     http.MethodGet,
	}
	if r.err != nil {
		return request, r.err
	}
	rawURL := r.rawURL
	if r.client.Options().PrefixPath != "" {
		rawURL = r.client.Options().PrefixPath + r.rawURL
	}
	if rawURL == "" {
		return request, ErrNoURL
	}
	request.Method = r.method
	if r.ctx != nil {
		request = request.WithContext(r.ctx)
	}
	if len(r.cookies) > 0 {
		for _, c := range r.cookies {
			request.AddCookie(c)
		}
	}
	if len(r.queryParams) > 0 {
		paramStr := r.queryParams.Encode()
		if strings.IndexByte(rawURL, '?') == -1 {
			rawURL = rawURL + "?" + paramStr
		} else {
			rawURL = rawURL + "&" + paramStr
		}
	}
	if len(r.uploads) > 0 && (request.Method == "POST" || request.Method == "PUT") {
		body := new(bytes.Buffer)
		bodyWriter := multipart.NewWriter(body)
		for key, values := range r.formParams {
			for _, val := range values {
				_ = bodyWriter.WriteField(key, val)
			}
		}
		for i, upload := range r.uploads {
			if upload.FieldName == "" {
				upload.FieldName = "file" + strconv.Itoa(i)
			}
			fileWriter, err := bodyWriter.CreateFormFile(upload.FieldName, upload.FileName)
			if err != nil {
				return request, err
			}
			_, err = io.Copy(fileWriter, upload.File)
			if err != nil {
				return request, err
			}
		}
		_ = bodyWriter.Close()

		r.WithBinaryBody(body.Bytes())
		r.WithContentType(bodyWriter.FormDataContentType())
	} else {
		if len(r.formParams) > 0 {
			r.WithBinaryBody([]byte(r.formParams.Encode()))
			r.WithContentType(ContentTypeForm)
		}
	}
	if len(r.body) > 0 {
		request.Body = io.NopCloser(bytes.NewReader(r.body))
		request.ContentLength = int64(len(r.body))
	}
	if r.header != nil {
		request.Header = r.header
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return request, err
	}
	request.URL = u

	if host := request.Header.Get("Host"); host != "" {
		request.Host = host
	}
	return request, nil
}

// Do is to call the request
func (r *Req) Do() *Resp {
	if r.client == nil {
		r.client = DefaultClient
	}
	return r.client.Do(r)
}

func toString(v interface{}) string {
	switch vv := v.(type) {
	case nil:
		return ""
	case string:
		return vv
	case int:
		return strconv.Itoa(vv)
	case int32:
		return strconv.Itoa(int(vv))
	case int64:
		return strconv.FormatInt(vv, 10)
	case int8:
		return strconv.FormatInt(int64(vv), 10)
	case int16:
		return strconv.FormatInt(int64(vv), 10)
	case uint:
		return strconv.FormatUint(uint64(vv), 10)
	case uint32:
		return strconv.FormatUint(uint64(vv), 10)
	case uint64:
		return strconv.FormatUint(vv, 10)
	case uint8:
		return strconv.FormatUint(uint64(vv), 10)
	case uint16:
		return strconv.FormatUint(uint64(vv), 10)
	case float32:
		return strconv.FormatFloat(float64(vv), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(vv, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(vv)
	case []byte:
		return string(vv)
	case fmt.Stringer:
		return vv.String()
	case error:
		return vv.Error()
	}
	return ""
}
