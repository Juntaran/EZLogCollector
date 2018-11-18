/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/18 18:25
  */

package reader

import (
	"io"
	"time"
)

type Reader interface {
	LineToMessage() (Message, error)
}

// reader event
type Message struct {
	Ts      time.Time     // timestamp the content was read
	Content []byte        // actual content read
	Bytes   int           // total number of bytes read to generate the message
	//Fields  common.MapStr // optional fields that can be added by reader
}

type Encode struct {
	reader *Line
}

// NewEncode creates a new Encode reader from input reader by applying
// the given codec.
func NewEncode(reader io.Reader, bufferSize int) Encode {
	return Encode{NewLine(reader, bufferSize)}
}

func (p Line) LineToMessage() (Message, error) {
	c, sz, err := p.Next()
	return Message{
		Ts: 		time.Now(),
		Content: 	c,
		Bytes: 		sz,
	}, err
}

type Limit struct {
	reader 		Reader
	maxBytes	int
}

// 创建一个限制 line 长度的 reader
func NewLimit(r Reader, maxBytes int) *Limit {
	return &Limit{
		reader: 	r,
		maxBytes: 	maxBytes,
	}
}

// 返回下一行
func (p *Limit) LineToMessage() (Message, error) {
	message, err := p.reader.LineToMessage()
	if len(message.Content) > p.maxBytes {
		message.Content = message.Content[:p.maxBytes]
	}
	return message, err
}


// StripNewline reader removes the last trailing newline characters from
// read lines.
type StripNewline struct {
	reader Reader
}

// NewStripNewline creates a new line reader stripping the last tailing newline.
func NewStripNewline(r Reader) *StripNewline {
	return &StripNewline{r}
}

// Next returns the next line.
func (p *StripNewline) LineToMessage() (Message, error) {
	message, err := p.reader.LineToMessage()
	if err != nil {
		return message, err
	}

	L := message.Content
	message.Content = L[:len(L)-lineEndingChars(L)]

	return message, err
}

// isLine checks if the given byte array is a line, means has a line ending \n
func isLine(l []byte) bool {
	return l != nil && len(l) > 0 && l[len(l)-1] == '\n'
}

// lineEndingChars returns the number of line ending chars the given by array has
// In case of Unix/Linux files, it is -1, in case of Windows mostly -2
func lineEndingChars(l []byte) int {
	if !isLine(l) {
		return 0
	}

	if len(l) > 1 && l[len(l)-2] == '\r' {
		return 2
	}
	return 1
}
