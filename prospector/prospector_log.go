/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/19 14:29
  */

package prospector

import (
	"expvar"
	"github.com/Juntaran/EZLogCollector/harvester/lcFile"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	filesRenamed  = expvar.NewInt("filebeat.prospector.log.files.renamed")
	filesTrucated = expvar.NewInt("filebeat.prospector.log.files.truncated")
)

type ProspectorLog struct {
	Prospector		*Prospector
	//config 			prospectorConfig
	lastClean 		time.Time
}

func NewProspectorLog(p *Prospector) (*ProspectorLog, error) {
	prospectorer := &ProspectorLog{
		Prospector: 	p,
		//config: 		p.config,
	}
	return prospectorer, nil
}

func (p *ProspectorLog) Init() {
	log.Println("Load previous states from registry into memory")
	fileStates := p.Prospector.states.GetStates()

	// 确保所有 states 已经设置为 finished
	for _, state := range fileStates {
		// 检查如果 state.Source 属于这个 prospector，更新 state
		if p.hasFile(state.Source) {
			// 把所有的 TTL 都设置成无穷以确保除非删除否则配置相同
			state.TTL = -1

			// 更新 prospector states 并发送新的 states 给 registry
			if err := p.Prospector.updateState(NewEvent(state)); err != nil {
				log.Printf("Problem putting initial state: %+v\n", err)
			}
		} else {
			p.Prospector.states.Update(state)
		}
	}
}

// hasFile returns true in case the given filePath is part of this prospector
func (p *ProspectorLog) hasFile(filePath string) bool {
	//for _, glob := range p.config.Paths {
	path := []string{"/Users/juntaran/Workspace/goWorkspace/src/github.com/Juntaran/EZLogCollector/tests/files/logs/*.log"}
	for _, glob := range path {
		// Evaluate the path as a wildcards/shell glob
		matches, err := filepath.Glob(glob)
		if err != nil {
			continue
		}
		log.Println("matches:", matches)
		// Check any matched files to see if we need to start a harvester
		for _, file := range matches {
			if filePath == file {
				return true
			}
		}
	}
	return false
}

func (p *ProspectorLog) Run() {
	log.Println("prospector", "Start next scan")

	p.scan()

	//// It is important that a first scan is run before cleanup to make sure all new states are read first
	//if p.config.CleanInactive > 0 || p.config.CleanRemoved {
	//	beforeCount := p.Prospector.states.Count()
	//	cleanedStates := p.Prospector.states.Cleanup()
	//	logp.Debug("prospector", "Prospector states cleaned up. Before: %d, After: %d", beforeCount, beforeCount-cleanedStates)
	//}

	// Marking removed files to be cleaned up. Cleanup happens after next scan to make sure all states are updated first
	//if p.config.CleanRemoved {
	//	for _, state := range p.Prospector.states.GetStates() {
	//		// os.Stat will return an error in case the file does not exist
	//		_, err := os.Stat(state.Source)
	//		if err != nil {
	//			// Only clean up files where state is Finished
	//			if state.Finished {
	//				state.TTL = 0
	//				err := p.Prospector.updateState(input.NewEvent(state))
	//				if err != nil {
	//					logp.Err("File cleanup state update error: %s", err)
	//				}
	//				logp.Debug("prospector", "Remove state for file as file removed: %s", state.Source)
	//			} else {
	//				logp.Debug("prospector", "State for file not removed because not finished: %s", state.Source)
	//			}
	//		}
	//	}
	//}
}

// 对每一个 path/glob 开启 scanGlob
func (p *ProspectorLog) scan() {
	for path, info := range p.getFiles() {
		log.Printf("Check file for harvesting: %s\n", path)

		// Create new state for comparison
		newState := lcFile.NewState(info, path)

		// Load last state
		lastState := p.Prospector.states.FindPrevious(newState)

		//// Ignores all files which fall under ignore_older
		//if p.isIgnoreOlder(newState) {
		//	logp.Debug("prospector", "Ignore file because ignore_older reached: %s", newState.Source)
		//
		//	// If last state is empty, it means state was removed or never created -> can be ignored
		//	if !lastState.IsEmpty() && !lastState.Finished {
		//		logp.Err("File is falling under ignore_older before harvesting is finished. Adjust your close_* settings: %s", newState.Source)
		//	}
		//	continue
		//}

		// 根据之前状态是否存在决定如何 harvest
		if lastState.IsEmpty() {
			log.Printf("Start harvester for new file: %s\n", newState.Source)
			err := p.Prospector.startHarvester(newState, 0)
			if err != nil {
				log.Printf("Harvester could not be started on new file: %s, Err: %s\n", newState.Source, err)
			}
		} else {
			p.harvestExistingFile(newState, lastState)
		}
	}
}

// 返回所有需要被 harvest 的文件
func (p *ProspectorLog) getFiles() map[string]os.FileInfo {
	paths := map[string]os.FileInfo{}

	// range p.config.Paths
	path := []string{"/Users/juntaran/Workspace/goWorkspace/src/github.com/Juntaran/EZLogCollector/tests/files/logs/*.log"}
	for _, glob := range path {
		matches, err := filepath.Glob(glob)
		if err != nil {
			log.Printf("glob(%s) failed: %v\n", glob, err)
			continue
		}

	//OUTER:
		for _, file := range matches {
			// Lstat 返回一个描述 nam e指定的文件对象的 FileInfo。 如果指定的文件对象是一个符号链接，返回的 FileInfo 描述该符号链接的信息
			fileInfo, err := os.Lstat(file)
			if err != nil {
				log.Printf("lstat(%s) failed: %s\n", file, err)
				continue
			}
			if fileInfo.IsDir() {
				log.Printf("Skipping directory: %s\n", file)
				continue
			}
			//isSymlink := fileInfo.Mode()&os.ModeSymlink > 0
			//if isSymlink && !p.config.Symlinks {
			//	logp.Debug("prospector", "File %s skipped as it is a symlink.", file)
			//	continue
			//}
			fileInfo, err = os.Stat(file)
			if err != nil {
				log.Printf("stat(%s) failed: %s\n", file, err)
				continue
			}

			paths[file] = fileInfo
		}
	}
	return paths
}

// 根据已知的 state 继续 harvest 一个文件
func (p *ProspectorLog) harvestExistingFile(newState lcFile.State, oldState lcFile.State) {
	log.Printf("Update existing file for harvesting: %s, offset: %v\n", newState.Source, oldState.Offset)
	// 没有 harvester 正在读取这个文件，启动一个新的 harvester
	if oldState.Finished && newState.Fileinfo.Size() < oldState.Offset {
		log.Printf("Old file was truncated. Starting from the beginning: %s\n", newState.Source)
		err := p.Prospector.startHarvester(newState, 0)
		if err != nil {
			log.Printf("Harvester could not be started on truncated file: %s, Err: %s", newState.Source, err)
		}
		filesTrucated.Add(1)
		return
	}

	// 检查文件名是否被修改了
	if oldState.Source != "" && oldState.Source != newState.Source {
		// 如果 old harvester 仍然在运行，不开启新的 harvester
		log.Printf("File rename was detected: %s -> %s, Current offset: %v\n", oldState.Source, newState.Source, oldState.Offset)

		if oldState.Finished {
			log.Printf("Updating state for renamed file: %s -> %s, Current offset: %v\n", oldState.Source, newState.Source, oldState.Offset)
			oldState.Source = newState.Source
			err := p.Prospector.updateState(NewEvent(oldState))
			if err != nil {
				log.Printf("File rotation state update error: %s\n", err)
			}
			filesRenamed.Add(1)
		} else {
			log.Printf("File rename detected but harvester not finished yet.")
		}
	}
	if !oldState.Finished {
		// 什么都不做
		log.Printf("Harvester for file is still running: %s\n", newState.Source)
	} else {
		log.Printf("File didn't change: %s\n", newState.Source)
	}
}