/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/23 19:16
  */

package main

import (
	"github.com/Juntaran/EZLogCollector/harvester/lcFile"
	"github.com/Juntaran/EZLogCollector/prospector"
	"log"
	"os"
)

func buildStates(path string) (lcFile.States, error) {
	fileinfo, err := os.Stat(path)
	if err != nil {
		log.Println(err)
		return *lcFile.NewStates(), err
	}

	state := lcFile.NewState(fileinfo, path)
	states := lcFile.NewStates()
	states.SetStates([]lcFile.State{state})
	return *states, nil
}

func main() {


	states, err := buildStates("/Users/juntaran/Workspace/goWorkspace/src/github.com/Juntaran/EZLogCollector/tests/files/logs/json.log")
	if err != nil {
		log.Println(err)
	}
	prospector, err := prospector.NewProspector(states, nil)
	if err != nil {
		log.Printf("Error in initing prospector: %s\n", err)
	}
	prospector.Run()
}
