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
	"syscall"
	"time"
)

// 文件状态
type State struct {
	Source      string      	`json:"source"`
	Offset      int64       	`json:"offset"`
	Finished    bool        	`json:"-"` // harvester state
	Fileinfo    os.FileInfo 	`json:"-"` // the lcFile info
	FileStateOS StateOS
	Timestamp   time.Time     	`json:"timestamp"`
	TTL         time.Duration 	`json:"ttl"`
}

// NewState creates a new file state
func NewState(fileInfo os.FileInfo, path string) State {
	return State{
		Fileinfo:    fileInfo,
		Source:      path,
		Finished:    false,
		FileStateOS: GetOSState(fileInfo),
		Timestamp:   time.Now(),
		TTL:         -1 * time.Second, // By default, state does have an infinit ttl
	}
}

type StateOS struct {
	Inode  uint64 `json:"inode,"`
	Device uint64 `json:"device,"`
}

// GetOSState returns the FileStateOS for non windows systems
func GetOSState(info os.FileInfo) StateOS {
	stat := info.Sys().(*syscall.Stat_t)
	// Convert inode and dev to uint64 to be cross platform compatible
	fileState := StateOS{
		Inode:  uint64(stat.Ino),
		Device: uint64(stat.Dev),
	}

	return fileState
}

// 查看文件是否相同
func (fs StateOS) IsSame(state StateOS) bool {
	return fs.Inode == state.Inode && fs.Device == state.Device
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

// 状态是否为空
func (s *State) IsEmpty() bool {
	return *s == State{}
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

// 更新状态
func (s *States) Update(newState State) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	index, _ := s.findPrevious(newState)
	newState.Timestamp = time.Now()
	if index > 0 {
		s.states[index] = newState
	} else {
		// 未找到旧状态，生成一个新状态
		s.states = append(s.states, newState)
		log.Printf("prospector new state added for %s\n", newState.Source)
	}
}

func (s *States) FindPrevious(newState State) State {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, state := s.findPrevious(newState)
	return state
}

// 搜索旧状态
func (s *States) findPrevious(newState State) (int, State) {
	for index, oldState := range s.states {
		if oldState.FileStateOS.IsSame(newState.FileStateOS) {
			return index, oldState
		}
	}
	return -1, State{}
}