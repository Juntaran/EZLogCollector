/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/18 01:50
  */

package reader

import (
	"bytes"
	"log"
	"testing"
)

//var tests = []struct {
//	encoding string
//	strings  []string
//}{
//	{"plain", []string{"I can", "eat glass"}},
//	//{"latin1", []string{"I kå Glas frässa", "ond des macht mr nix!"}},
//	//{"utf-16be", []string{"Pot să mănânc sticlă", "și ea nu mă rănește."}},
//	//{"utf-16le", []string{"काचं शक्नोम्यत्तुम् ।", "नोपहिनस्ति माम् ॥"}},
//	//{"big5", []string{"我能吞下玻", "璃而不傷身體。"}},
//	//{"gb18030", []string{"我能吞下玻璃", "而不傷身。體"}},
//	//{"euc-kr", []string{" 나는 유리를 먹을 수 있어요.", " 그래도 아프지 않아요"}},
//	//{"euc-jp", []string{"私はガラスを食べられます。", "それは私を傷つけません。"}},
//}

var tests = []string{"I can", "eat glass"}

func TestReaderEncoding(t *testing.T)  {
	buffer := bytes.NewBuffer(nil)
	var expectedCount []int
	for _, line := range tests {
		log.Println("reader:", line)
		buffer.Write([]byte(line))
		buffer.Write([]byte{'\n'})
		expectedCount = append(expectedCount, buffer.Len())
	}

	reader := NewLine(buffer, 1024)

	var readLines []string
	var byteCounts []int
	current := 0
	for {
		bytes, sz, err := reader.Next()
		if sz > 0 {
			readLines = append(readLines, string(bytes[:len(bytes)-1]))
		}

		if err != nil {
			break
		}

		current += sz
		byteCounts = append(byteCounts, current)
	}

	// validate lines and byte offsets
	if len(tests) != len(readLines) {
		t.Errorf("number of lines mismatch (expected=%v actual=%v)", len(tests), len(readLines))
	}

	for i := range tests {
		expected := tests[i]
		actual := readLines[i]
		if expected != actual {
			t.Errorf("expect != actual (expected=%v actual=%v)", expected, actual)
			continue
		}
	}
}