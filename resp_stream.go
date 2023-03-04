package goreq

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/http"
	"sync"
)

type RespStream struct {
	err                error
	reader             *bufio.Reader
	response           *http.Response
	state              string
	curEventName       []byte
	emptyMessagesCount int
	closeOnce          sync.Once
}

const (
	StreamStateOpen    = "open"
	StreamStateClosed  = "closed"
	StreamEventMessage = "message"
)

var (
	streamFieldID                 = []byte("id: ")
	streamFieldEvent              = []byte("event: ")
	streamFieldData               = []byte("data: ")
	streamFieldRetry              = []byte("retry: ")
	returnDelim                   = []byte("\n")
	ErrStreamClosed               = errors.New("stream closed")
	ErrTooManyEmptyStreamMessages = errors.New("stream has sent too many empty messages")
)

func (rs *RespStream) State() string {
	return rs.state
}

func (rs *RespStream) Close() {
	rs.closeOnce.Do(func() {
		rs.state = StreamStateClosed
		rs.response.Body.Close()
	})
}

func (rs *RespStream) Read() (eventName string, data string, err error) {
	if rs.err != nil {
		return "", "", rs.err
	}
	if rs.state == StreamStateClosed {
		return "", "", ErrStreamClosed
	}
	var line []byte
	line, err = rs.readData()
	if len(rs.curEventName) == 0 {
		return StreamEventMessage, string(line), nil
	} else {
		return string(rs.curEventName), string(line), nil
	}
}

func (rs *RespStream) readData() ([]byte, error) {
	line, err := rs.reader.ReadBytes('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			rs.Close()
			return []byte{}, err
		}
	}
	line = bytes.TrimSpace(line)
	if bytes.Equal(line, returnDelim) {
		rs.curEventName = []byte{}
		return rs.readData()
	} else if bytes.HasPrefix(line, streamFieldEvent) {
		rs.curEventName = streamFieldEvent
		return rs.readData()
	} else if bytes.HasPrefix(line, streamFieldData) {
		line = bytes.TrimPrefix(line, streamFieldData)
		return line, nil
	} else {
		return rs.readData()
	}
}
