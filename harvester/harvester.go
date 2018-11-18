/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/14 19:00
  */

package harvester

import (
	"bytes"
	"fmt"
	"github.com/Juntaran/EZLogCollector/harvester/reader"
	"io"
	"log"
	"os"
	"time"

	"github.com/Juntaran/EZLogCollector/harvester/lcFile"
)

type Harvester struct {
	//config          harvesterConfig
	state lcFile.State
	//prospectorChan  chan *input.Event
	file            FileSource /* the lcFile being watched */
	fileReader      *LogFile
	//encodingFactory encoding.EncodingFactory
	//encoding        encoding.Encoding
	done            chan struct{}
}

type LogSource interface {
	io.ReadCloser
	Name() string
}

type FileSource interface {
	LogSource
	Stat() (os.FileInfo, error)
	Continuable() bool // can we continue processing after EOF?
}

func NewHarvester(state lcFile.State, done chan struct{}) *Harvester {
	h := &Harvester {
		state: 			state,
		done:			done,
	}
	return h
}

// 开启 lcFile handler 并创建 harvester 的 reader
func (h *Harvester) Setup() (reader.Reader, error) {
	if err := h.openFile(); err != nil {
		return nil, fmt.Errorf("Harvester setup failed. Unexpected lcFile opening error: %s", err)
	}

	if r, err := h.newLogFileReader(); err != nil {
		if h.file != nil {
			h.file.Close()
		}
		return nil, fmt.Errorf("Harvester setup failed. Unexpected encoding line reader error: %s", err)
	} else {
		return r, nil
	}
}

// 打开一个文件
func (h *Harvester) openFile() error {
	f, err := lcFile.ReadOpen(h.state.Source)
	if err != nil {
		return fmt.Errorf("Failed opening %s: %s", h.state.Source, err)
	}
	harvesterOpenFiles.Add(1)
	h.file = lcFile.File{f}
	return nil
}

// Harvest 逐行读取文件并发送 event
func (h *Harvester) Harvest(r reader.Reader) {
	harvesterStarted.Add(1)
	harvesterRunning.Add(1)
	defer harvesterRunning.Add(-1)
	defer h.close()

	harvestDone := make(chan struct{})
	defer close(harvestDone)

	go func() {
		var closeTimeout <-chan time.Time
		closeTimeout = time.After(5)

		select {
		// Applies when timeout is reached
		case <-closeTimeout:
			log.Printf("Closing harvester because close_timeout was reached: %s\n", h.state.Source)
			// Required for shutdown when hanging inside reader
		case <-h.done:
			// Required when reader loop returns and reader finished
		case <-harvestDone:
		}
		h.fileReader.Close()
	}()

	log.Printf("Harvester started for lcFile: %s\n", h.state.Source)

	for {
		select {
		case <-h.done:
			return
		default:
		}

		message, err := r.LineToMessage()
		if err != nil {
			switch err {
			case ErrFileTruncate:
				log.Printf("File was truncated. Begin reading lcFile from offset 0: %s\n", h.state.Source)
				h.state.Offset = 0
				filesTruncated.Add(1)
			case ErrRemoved:
				log.Printf("File was removed: %s. Closing because close_removed is enabled.\n", h.state.Source)
			case ErrRenamed:
				log.Printf("File was renamed: %s. Closing because close_renamed is enabled.\n", h.state.Source)
			case ErrClosed:
				log.Printf("Reader was closed: %s. Closing.\n", h.state.Source)
			case io.EOF:
				log.Printf("End of lcFile reached: %s. Closing because close_eof is enabled.\n", h.state.Source)
			case ErrInactive:
				log.Printf("File is inactive: %s. Closing because close_inactive of %v reached.\n", h.state.Source, 5)
			default:
				log.Printf("Read line error: %s; File: %v\n", err, h.state.Source)
			}
			return
		}

		// Strip UTF-8 BOM if beginning of lcFile
		// As all BOMS are converted to UTF-8 it is enough to only remove this one
		if h.state.Offset == 0 {
			message.Content = bytes.Trim(message.Content, "\xef\xbb\xbf")
		}

		// Update offset
		h.state.Offset += int64(message.Bytes)

		// Create state event
		//event := input.NewEvent(h.getState())

		text := string(message.Content)

		log.Println("text:", text)

		//// Check if data should be added to event. Only export non empty events.
		//if !message.IsEmpty() && h.shouldExportLine(text) {
		//	event.ReadTime = message.Ts
		//	event.Bytes = message.Bytes
		//	event.Text = &text
		//	event.JSONFields = message.Fields
		//	event.EventMetadata = h.config.EventMetadata
		//	event.InputType = h.config.InputType
		//	event.DocumentType = h.config.DocumentType
		//	event.JSONConfig = h.config.JSON
		//}
		//
		//// Always send event to update state, also if lines was skipped
		//// Stop harvester in case of an error
		//if !h.sendEvent(event) {
		//	return
		//}
	}
}

func (h *Harvester) close() {
	// 标记 harvester 已完成
	h.state.Finished = true
	log.Printf("Stopping harvester for lcFile: %s\n", h.state.Source)

	if h.file != nil {
		h.file.Close()
		log.Printf("Closing lcFile: %s\n", h.state.Source)
		harvesterOpenFiles.Add(-1)

		// 更新 offset
		//h.sendStateUpdate()
	} else {
		log.Printf("Stopping harvester, NOT closing lcFile as lcFile info not available: %s\n", h.state.Source)
	}
	harvesterClosed.Add(1)
}