/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/18 00:13
  */

package streambuf

import (
	"bytes"
)

type Buffer struct {
	data  []byte
	err   error
	fixed bool

	// Internal parser state offsets.
	// Offset is the position a parse might continue to work at when called
	// again (e.g. usefull for parsing tcp streams.). The mark is used to remember
	// the position last parse operation ended at. The variable available is used
	// for faster lookup
	// Invariants:
	//    (1) 0 <= mark <= offset
	//    (2) 0 <= available <= len(data)
	//    (3) available = len(data) - mark
	// offset 是再次调用的时候，可能继续处理的为止(例如解析 tcp stream)
	// mark 标记了最后一个处理的位置，可以快速寻找
	mark, offset, available int
}

func (b *Buffer) Init(data []byte, fixed bool) {
	b.data 		= data
	b.err		= nil
	b.fixed 	= fixed
	b.mark  	= 0
	b.offset 	= 0
	b.available = len(data)
}

func New(data []byte) *Buffer {
	return &Buffer{
		data: 		data,
		fixed:		false,
		available: 	len(data),
	}
}

func NexFixed(data []byte) *Buffer {
	return &Buffer{
		data: 		data,
		fixed:		true,
		available: 	len(data),
	}
}

// 返回所有剩余未处理的 bytes
func (b *Buffer) Bytes() []byte {
	return b.data[b.mark:]
}

// 返回未读取部分的长度
func (b *Buffer) Len() int {
	return b.available
}

// 检查 available 是否大于 count
func (b *Buffer) Avail(count int) bool {
	return count <= b.available
}

// 从 buffer 中收集 count 数量的 bytes，并更新 read 指针
func (b *Buffer) Collect(count int) ([]byte, error) {
	if b.err != nil {
		return nil, b.err
	}
	if !b.Avail(count) {
		// count > b.available
		return nil, b.bufferEndError()
	}
	data := b.data[b.mark : b.mark + count]
	b.Advance(count)
	return data, nil
}

// 向前推进 buffer 读取指针 count bytes
func (b *Buffer) Advance(count int) error {
	if !b.Avail(count) {
		return b.bufferEndError()
	}
	b.mark 		+= count
	b.offset 	= b.mark
	b.available -= count

	return nil
}

// 从 buffer 移除所有已经处理的 bytes
func (b *Buffer) Reset() {
	b.data 		= b.data[b.mark:]
	b.offset 	-= b.mark
	b.mark 		= 0
	b.available = len(b.data)
	b.err 		= nil
}

// 返回 seq 从 from 开始的未处理的 buffer 的 offset
func (b *Buffer) IndexFrom(from int, seq []byte) int {
	if b.err != nil {
		return -1
	}
	// bytes.Index 返回 seq 在 s 中第一次出现的位置，不存在 -1
	idx := bytes.Index(b.data[b.mark + from:], seq)
	if idx < 0 {
		return -1
	}
	return idx + from + b.mark
}

// p 写入 buffer(buffer 不固定)
func (b *Buffer) Write(p []byte) (int, error) {
	err := b.doAppend(p, false, -1)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// 修改 Buffer 中的 err
func (b *Buffer) SetError(err error) error {
	b.err = err
	return err
}

func (b *Buffer) bufferEndError() error {
	if b.fixed {
		return b.SetError(ErrUnexpectedEOB)
	} else {
		return b.SetError(ErrNoMoreBytes)
	}
}

func (b *Buffer) doAppend(data []byte, retainable bool, newCap int) error {
	if b.fixed {
		return b.SetError(ErrOperationNotAllowed)
	}
	if b.err != nil && b.err != ErrNoMoreBytes {
		return b.err
	}

	if len(b.data) == 0 {
		retain := retainable && cap(data) > newCap
		if retain {
			b.data = data
		} else {
			if newCap < len(data) {
				b.data = make([]byte, len(data))
			} else {
				b.data = make([]byte, len(data), newCap)
			}
			copy(b.data, data)
		}
	} else {
		if newCap > 0 && cap(b.data[b.offset:]) < len(data) {
			required := cap(b.data) + len(data)
			if required < newCap {
				tmp := make([]byte, len(b.data), newCap)
				copy(tmp, b.data)
				b.data = tmp
			}
		}
		b.data = append(b.data, data...)
	}
	b.available += len(data)

	// 重置 error 状态
	if b.err == ErrNoMoreBytes {
		b.err = nil
	}
	return nil
}

// data 写入 buffer
func (b *Buffer) Append(data []byte) error {
	return b.doAppend(data, true, -1)
}