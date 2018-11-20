/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/19 02:47
  */

package prospector

import (
	"github.com/Juntaran/EZLogCollector/harvester/lcFile"
	"time"
)

// 发送到 outlet，包含所有相关数据
type Event struct {
	ReadTime     time.Time
	InputType    string
	DocumentType string
	Bytes        int
	Text         *string
	State        lcFile.State
}

func NewEvent(state lcFile.State) *Event {
	return &Event{
		State: state,
	}
}
