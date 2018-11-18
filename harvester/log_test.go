/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/18 20:54
  */

package harvester

import (
	"fmt"
	"github.com/Juntaran/EZLogCollector/harvester/lcFile"
	"github.com/Juntaran/EZLogCollector/harvester/reader"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestReadLine(t *testing.T) {

	absPath, err := filepath.Abs("../tests/files/logs/")
	// All files starting with tmp are ignored
	logFile := absPath + "/tmp" + strconv.Itoa(rand.Int()) + ".log"

	assert.NotNil(t, absPath)
	assert.Nil(t, err)

	if err != nil {
		t.Fatalf("Error creating the absolute path: %s", absPath)
	}

	file, err := os.Create(logFile)
	defer file.Close()
	defer os.Remove(logFile)

	assert.Nil(t, err)
	assert.NotNil(t, file)

	firstLineString := "9Characte\n"
	secondLineString := "This is line 2\n"

	length, err := file.WriteString(firstLineString)
	assert.Nil(t, err)
	assert.NotNil(t, length)

	length, err = file.WriteString(secondLineString)
	assert.Nil(t, err)
	assert.NotNil(t, length)

	file.Sync()

	// Open lcFile for reading
	readFile, err := os.Open(logFile)
	defer readFile.Close()
	assert.Nil(t, err)

	f := lcFile.File{readFile}

	h := Harvester{
		//config: harvesterConfig{
		//	CloseInactive: 500 * time.Millisecond,
		//	Backoff:       100 * time.Millisecond,
		//	MaxBackoff:    1 * time.Second,
		//	BackoffFactor: 2,
		//	BufferSize:    100,
		//	MaxBytes:      1000,
		//},
		file: f,
	}
	assert.NotNil(t, h)

	//var ok bool
	//h.encodingFactory, ok = encoding.FindEncoding(h.config.Encoding)
	//assert.True(t, ok)
	//
	//h.encoding, err = h.encodingFactory(readFile)
	//assert.NoError(t, err)

	r, err := h.newLogFileReader()
	assert.NoError(t, err)

	// Read third line
	_, text, bytesread, err := readLine(r)
	fmt.Printf("received line: '%s'\n", text)
	assert.Nil(t, err)
	assert.Equal(t, text, firstLineString[0:len(firstLineString)-1])
	assert.Equal(t, bytesread, len(firstLineString))

	// read second line
	_, text, bytesread, err = readLine(r)
	fmt.Printf("received line: '%s'\n", text)
	assert.Equal(t, text, secondLineString[0:len(secondLineString)-1])
	assert.Equal(t, bytesread, len(secondLineString))
	assert.Nil(t, err)

	// Read third line, which doesn't exist
	_, text, bytesread, err = readLine(r)
	fmt.Printf("received line: '%s'\n", text)
	assert.Equal(t, "", text)
	assert.Equal(t, bytesread, 0)
	assert.Equal(t, err, ErrInactive)
	//log.Println(err, ErrInactive)
}

func readLine(reader reader.Reader) (time.Time, string, int, error) {
	message, err := reader.LineToMessage()

	// Full line read to be returned
	if message.Bytes != 0 && err == nil {
		return message.Ts, string(message.Content), message.Bytes, err
	}

	return time.Time{}, "", 0, err
}
