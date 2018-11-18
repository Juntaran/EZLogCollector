/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/18 17:55
  */

package harvester

import (
	"fmt"
	"github.com/Juntaran/EZLogCollector/lcFile"
	"io"
	"log"
	"os"
	"time"
)

type LogFile struct {
	fs           FileSource
	offset       int64
	//config       harvesterConfig
	lastTimeRead time.Time
	backoff      time.Duration
	done         chan struct{}
}

func NewLogFile(fs FileSource) (*LogFile, error) {
	var offset int64
	if seeker, ok := fs.(io.Seeker); ok {
		var err error
		offset, err = seeker.Seek(0, os.SEEK_CUR)
		if err != nil {
			return nil, err
		}
	}

	return &LogFile{
		fs:           fs,
		offset:       offset,
		//config:       config,
		lastTimeRead: time.Now(),
		backoff:      100000000,
		done:         make(chan struct{}),
	}, nil
}

// 从 reader 读取并更新 offset，返回总读取的 bytes 长度
func (r *LogFile) Read(buf []byte) (int, error) {
	totalN := 0

	for {
		select {
		case <- r.done:
			return 0, ErrClosed
		default:

		}

		n, err := r.fs.Read(buf)
		if n > 0 {
			r.offset += int64(n)
			r.lastTimeRead = time.Now()
		}
		totalN += n

		// 从源文件(source)中读取，直到出错或 buffer 满了
		if err == nil {
			// 重置 backoff 以便与下一次 read
			r.backoff = 100000000
			return totalN, nil
		}
		buf = buf[n:]

		// Checks if an error happened or buffer is full
		// If buffer is full, cannot continue reading.
		// Can happen if n == bufferSize + io.EOF error
		err = r.errorChecks(err)
		if err != nil || len(buf) == 0 {
			return totalN, err
		}

		log.Printf( "End of file reached: %s; Backoff now.\n", r.fs.Name())
		r.wait()
	}
	return 0, nil
}

func (r *LogFile) Close() {
	close(r.done)
	// Note: File reader is not closed here because that leads to race conditions
}


func (r *LogFile) wait() {
	// Wait before trying to read file again. File reached EOF.
	select {
	case <-r.done:
		return
	case <-time.After(r.backoff):
	}

	// Increment backoff up to maxBackoff
	if r.backoff < 100000000 {
		r.backoff = r.backoff * time.Duration(2)
		if r.backoff > 100000000 {
			r.backoff = 100000000
		}
	}
}

func (r *LogFile) errorChecks(err error) error {
	if err != io.EOF {
		return fmt.Errorf("Unexpected state reading from %s; error: %s\n", r.fs.Name(), err)
	}

	// Stdin is not continuable
	if !r.fs.Continuable() {
		return fmt.Errorf("Source is not continuable: %s\n", r.fs.Name())
	}

	if err == io.EOF && false {
		return err
	}

	// Refetch fileinfo to check if the file was truncated or disappeared.
	// Errors if the file was removed/rotated after reading and before
	// calling the stat function
	info, statErr := r.fs.Stat()
	if statErr != nil {
		return fmt.Errorf("Unexpected error reading from %s; error: %s\n", r.fs.Name(), statErr)
	}

	// check if file was truncated
	if info.Size() < r.offset {
		return fmt.Errorf("File was truncated as offset (%d) > size (%d): %s\n", r.offset, info.Size(), r.fs.Name())
	}

	// Check file wasn't read for longer then CloseInactive
	age := time.Since(r.lastTimeRead)
	if age > 500 * time.Millisecond {
		return ErrInactive
	}

	if !lcFile.IsSameFile(r.fs.Name(), info) {
		return ErrRenamed
	}

	//if r.config.CloseRemoved {
	//	// Check if the file name exists. See https://github.com/elastic/filebeat/issues/93
	//	_, statErr := os.Stat(r.fs.Name())
	//
	//	// Error means file does not exist.
	//	if statErr != nil {
	//		return ErrRemoved
	//	}
	//}

	return nil
}