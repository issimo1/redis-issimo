package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/issimo1/redis-issimo/utils/logger"
	"io"
	"runtime/debug"
	"strconv"
)

type Payload struct {
	Err error
	//Reply
}

func ParseStream(reader io.Reader) <-chan *Payload {
	dataStream := make(chan *Payload)
	// 启动协程
	go parse(reader, dataStream)
	return dataStream
}

func parse(r io.Reader, ds chan<- *Payload) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err, string(debug.Stack()))
		}
	}()
	reader := bufio.NewReader(r)
	for {

		line, err := reader.ReadBytes('\n')
		if err != nil {
			ds <- &Payload{Err: err}
			close(ds)
			return
		}

		length := len(line)
		if length <= 2 || line[length-2] != '\r' {
			continue
		}
		line = bytes.TrimSuffix(line, []byte{'\r', '\n'})
		switch line[0] {
		case '*':
			err := parseArrays(line, reader, ds)
			if err != nil {

			}
		case '-':
		case '+':
		case ':':
		case '$':

		}
	}
}

func parseArrays(header []byte, reader *bufio.Reader, out chan<- *Payload) error {
	bodyCnt, err := strconv.ParseInt(string(header[1]), 10, 64)
	if err != nil || bodyCnt < 0 {
		out <- &Payload{Err: fmt.Errorf("illegal array header: %s", string(header[1]))}
		return nil
	}
	lines := make([][]byte, 0, bodyCnt)
	for i := int64(0); i < bodyCnt; i++ {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return err
		}
		//*5/r/n
		length := len(line)
		if length < 4 || line[length-2] != '\r' || line[0] != '$' {
			out <- &Payload{Err: fmt.Errorf("illegal string syntax: %s", string(line))}
			return nil
		}

		// 字符串长度为length[1]
		dataLen, err := strconv.ParseInt(string(line[1]), 10, 64)
		if err != nil || dataLen < -1 {
			out <- &Payload{Err: fmt.Errorf("illegal string length: %s", string(line[1]))}
			return nil
		} else if dataLen == -1 {
			lines = append(lines, nil)
		} else {
			body := make([]byte, 0, dataLen+2)
			_, err := io.ReadFull(reader, body)
			if err != nil {
				return err
			}
			lines = append(lines, body[:len(body)-2])
		}
	}
	out <- &Payload{Err: nil}
	return nil
}
