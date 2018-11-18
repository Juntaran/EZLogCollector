/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/18 00:09
  */

package reader

import (
	"io"
	"log"

	"github.com/Juntaran/EZLogCollector/streambuf"
)

type Line struct {
	reader     		io.Reader
	bufferSize 		int
	newline    		[]byte
	inBuffer   		*streambuf.Buffer
	outBuffer  		*streambuf.Buffer
	inOffset   		int 			// input buffer read offset
	byteCount  		int 			// number of bytes decoded from input buffer into output buffer
	//codec      		encoding.Encoding
	//decoder    		transform.Transformer
}

// 创建一个 new Line reader
func NewLine(input io.Reader, bufferSize int) *Line {
	return &Line{
		reader: 	input,
		bufferSize: bufferSize,
		newline: 	[]byte{'\n'},
		inBuffer:   streambuf.New(nil),
		outBuffer:  streambuf.New(nil),
		//codec:  	codec,
		//decoder:  codec.NewDecoder(),
	}
}

// 读取下一行直到 \n
func (l *Line) Next() ([]byte, int, error) {
	for {
		if err := l.advance(); err != nil {
			return nil, 0, err
		}

		buf := l.outBuffer.Bytes()
		if len(buf) == 0 {
			log.Println("return an empty buffer")
			continue
		}

		if buf[len(buf) - 1] == '\n' {
			break
		} else {
			log.Println("Line end character is", buf[len(buf) - 1])
		}
	}

	// output buffer 包含了包括 \n 在内的完整的一行的信息
	// 从 buffer 中提取 byte slice 并重置 output buffer
	bytes, err := l.outBuffer.Collect(l.outBuffer.Len())
	l.outBuffer.Reset()
	if err != nil {
		panic(err)
	}
	sz := l.byteCount
	l.byteCount = 0
	return bytes, sz, nil
}

// 从 buffer 读取直到 \n 出现
func (l *Line) advance() error {
	// 检查 buffer 是否存在一个 \n
	idx := l.inBuffer.IndexFrom(l.inOffset, l.newline)

	// 填充 inBuffer 直到 \n 在 input buffer 中出现
	for idx == -1 {
		// 不断增加搜索 offset 并减少迭代的 buffer
		newOffset := l.inBuffer.Len() - len(l.newline)
		if newOffset > l.inOffset {
			l.inOffset = newOffset
		}

		buf := make([]byte, l.bufferSize)

		// 读取更多的 bytes 写入 buffer
		n, err := l.reader.Read(buf)

		l.inBuffer.Append(buf[:n])
		if err != nil {
			return err
		}

		if n == 0 {
			return streambuf.ErrNoMoreBytes
		}

		// 检查 buffer 是否包含 \n
		idx = l.inBuffer.IndexFrom(l.inOffset, l.newline)
	}

	// 在 buffer 中已经出现了 \n
	// 在 filebeat 5.0 版本基础上消除了解码
	sz 	:= idx + len(l.newline)
	inBytes := l.inBuffer.Bytes()
	for i := 0; i < len(inBytes); i++ {
		if inBytes[i] == '\n' {
			l.outBuffer.Write(inBytes[:i+1])
			l.byteCount += i+1
			break
		}
	}

	//l.outBuffer.Write(l.inBuffer.Bytes())

	err := l.inBuffer.Advance(sz)
	l.inBuffer.Reset()

	// 从最后的位置 +1 继续扫描 input buffer
	l.inOffset = idx + 1 - sz
	if l.inOffset < 0 {
		l.inOffset = 0
	}
	return err
}


