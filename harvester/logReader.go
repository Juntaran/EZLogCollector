/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/14 18:55
  */

package harvester

import "github.com/Juntaran/EZLogCollector/reader"

// newLogFileReader creates a new reader to read log files
//
// It creates a chain of readers which looks as following:
//
//   limit -> (multiline -> timeout) -> strip_newline -> json -> encode -> reader -> log_file
//
// Each reader on the left, contains the reader on the right and calls `Next()` to fetch more data.
// At the base of all readers the the log_file reader. That means in the data is flowing in the opposite direction:
//
//   log_file -> reader -> encode -> json -> strip_newline -> (timeout -> multiline) -> limit
//
// log_file implements io.Reader interface and encode reader is an adapter for io.Reader to
// reader.Reader also handling lcFile encodings. All other readers implement reader.Reader
func (h *Harvester) newLogFileReader() (reader.Reader, error) {

	var r reader.Reader
	var err error

	// TODO: NewLineReader uses additional buffering to deal with encoding and testing
	//       for new lines in input stream. Simple 8-bit based encodings, or plain
	//       don't require 'complicated' logic.
	h.fileReader, err = NewLogFile(h.file)
	if err != nil {
		return nil, err
	}

	r = reader.NewLine(h.fileReader, 1024)
	if err != nil {
		return nil, err
	}

	//if h.config.JSON != nil {
	//	r = reader.NewJSON(r, h.config.JSON)
	//}
	//
	//r = reader.NewStripNewline(r)
	//
	//if h.config.Multiline != nil {
	//	r, err = reader.NewMultiline(r, "\n", h.config.MaxBytes, h.config.Multiline)
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	r = reader.NewStripNewline(r)
	//return reader.NewLimit(r, h.config.MaxBytes), nil
	return reader.NewLimit(r, 1000), nil
}