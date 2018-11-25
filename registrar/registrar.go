/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/25 22:21
  */

package registrar

import (
	"github.com/Juntaran/EZLogCollector/harvester/lcFile"
	"github.com/Juntaran/EZLogCollector/prospector"
	"github.com/Juntaran/EZLogCollector/publisher"
	"github.com/pkg/errors"
	"sync"
)

type Registrar struct {
	Channel      chan []*prospector.Event
	out          publisher.SuccessLogger
	done         chan struct{}
	registryFile string       // Path to the Registry File
	states       *lcFile.States // Map with all file paths inside and the corresponding state
	wg           sync.WaitGroup
}

var (
	sigPublisherStop = errors.New("publisher was stopped")
)

//func getDataEvents(events []*prospector.Event) [] {
//
//}