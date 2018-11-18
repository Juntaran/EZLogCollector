/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/18 17:53
  */

package lcFile

import (
	"log"
	"os"
	"sync"
	"time"
)

// 文件状态
type State struct {
	Source      string      	`json:"source"`
	Offset      int64       	`json:"offset"`
	Finished    bool        	`json:"-"` // harvester state
	Fileinfo    os.FileInfo 	`json:"-"` // the lcFile info
	//FileStateOS StateOS
	Timestamp   time.Time     	`json:"timestamp"`
	TTL         time.Duration 	`json:"ttl"`
}

type StateOS struct {
	Inode  uint64 `json:"inode,"`
	Device uint64 `json:"device,"`
}

// 只读方式打开一个文件
func ReadOpen(path string) (*os.File, error) {

	flag := os.O_RDONLY
	var perm os.FileMode = 0

	return os.OpenFile(path, flag, perm)
}

type File struct {
	*os.File
}

func (File) Continuable() bool { return true }

// 检查路径与给出的文件信息是否对应
func IsSameFile(path string, info os.FileInfo) bool {
	fileInfo, err := os.Stat(path)

	if err != nil {
		log.Printf("Error during file comparison: %s with %s - Error: %s\n", path, info.Name(), err)
		return false
	}

	return os.SameFile(fileInfo, info)
}

// 一批文件状态
type States struct {
	states []State
	mutex  sync.Mutex
}

func NewStates() *States {
	return &States{
		states: []State{},
	}
}

// 返回文件状态
func (s *States) GetStates() []State {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	newStates := make([]State, len(s.states))
	copy(newStates, s.states)

	return newStates
}

// 替换原有的文件状态
func (s *States) SetStates(states []State) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.states = states
}

// 创建一个新状态对象
func (s *States) Copy() *States {
	states := NewStates()
	states.states = s.GetStates()
	return states
}