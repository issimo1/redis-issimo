package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {

	var k kk
	fmt.Printf("%+v", k.getMap())
	re := strings.NewReader("qwwe\r\nweqwe\nqq\nwqewq\r\n12321\n")
	reader := bufio.NewReader(re)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			return
		}
		t.Log(line)
		line = bytes.TrimSuffix(line, []byte{'\r', '\n'})
		t.Log(line)
	}

}

func read(reader *bufio.Reader) {
	line, err := reader.ReadBytes('\n')
	if err == io.EOF {
		return
	}
	fmt.Println(line)
}

type kk struct {
	p map[string]string
}

func (k *kk) getMap() map[string]string {
	return k.p
}
