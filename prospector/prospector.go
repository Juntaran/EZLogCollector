/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/19 02:29
  */

package prospector

import (
	"expvar"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Juntaran/EZLogCollector/harvester"
	"github.com/Juntaran/EZLogCollector/harvester/lcFile"
)

var (
	harvesterSkipped = expvar.NewInt("filebeat.harvester.skipped")
)

/*
	prospector 功能:
	1. 解析配置，默认读取的日志是 nginx log
		conf: 	NginxPath		nginx.conf 路径
				NginxFormat 	nginx 格式，默认 main
				close_inactive	不活跃关闭时间，默认 time.Hour * 24 * 7
				Backoff 		在达到 EOF 之后再次检查文件之间等待的时间
				MaxBackoff		在达到 EOF 之后再次检查文件之前等待的最长时间
				backoff_factor	指定 backoff 尝试等待时间几次，默认是 2
				buffer_size		每次 harvester 读取文件缓冲字节数，默认是 16384
				scan_frequency	prospector 检查指定用于 harvest 的路径的新文件的频率，默认 10s
				//ignore_older	忽略时间段以外的日志内容	ignore_older > close_inactive
		还要支持正则 * 的使用
	2. 获取日志信息
		对每一个文件生成一个 harvester 读取日志
		构造一个 chan，保存每个 harvester 返回的信息
*/
type Prospector struct {
	prospectorer      Prospectorer
	outlet            Outlet
	harvesterChan     chan *Event
	done              chan struct{}
	states            *lcFile.States
	wg                sync.WaitGroup
	harversterCounter uint64
}

// Prospectorer 接口由 ProspectorLog 和 ProspectorStdin 实现
// 此处已经忽略 ProspectorStdin
type Prospectorer interface {
	Init()
	Run()
}

type Outlet interface {
	OnEvent(event *Event) bool
}

func NewProspector(states lcFile.States, outlet Outlet) (*Prospector, error) {
	prospector := &Prospector{
		outlet:        outlet,
		harvesterChan: make(chan *Event),
		done:          make(chan struct{}),
		states:        states.Copy(),
		wg:            sync.WaitGroup{},
	}

	if err := prospector.Init(); err != nil {
		return nil, err
	}
	return prospector, nil
}

func (p *Prospector) Init() error {
	var prospectorer Prospectorer
	var err error
	prospectorer, err = NewProspectorLog(p)
	prospectorer.Init()
	p.prospectorer = prospectorer
	_, err = p.createHarvester(lcFile.State{})
	if err != nil {
		return err
	}
	return nil
}

// 扫描所有文件路径，对每一个文件开启一个 harvester 获取内容
func (p *Prospector) Run() {
	log.Println("Starting prospector")
	p.wg.Add(2)
	defer p.wg.Done()

	// 开启一个 channel 接收 harvester 提取的 events 并把他们转发给 spooler
	go func() {
		defer p.wg.Done()
		for {
			select {
			case <-p.done:
				log.Println("Prospector channel stopped")
				return
			case event := <-p.harvesterChan:
				err := p.updateState(event)
				if err != nil {
					return
				}
			}
		}
	}()

	// 启动 prospector
	p.prospectorer.Run()

	for {
		select {
		case <-p.done:
			log.Println("Prospector ticker stopped")
			return
		case <-time.After(time.Second * 10):
			log.Println("Run prospector")
			p.prospectorer.Run()
		}
	}
}

// 通过文件状态创建一个新的 harvester 实例
func (p *Prospector) createHarvester(state lcFile.State) (*harvester.Harvester, error) {
	h := harvester.NewHarvester(
		//p.cfg,
		state,
		//p.harvesterChan,
		p.done,
	)
	return h, nil
}

// 更新 prospector state 并发送 event 给 spooler
// 同时更新所有 state
func (p *Prospector) updateState(event *Event) error {
	// default clean_inactive = 0
	if event.State.TTL != 0 && false {
		event.State.TTL = time.Second * 3600
	}
	//ok := p.outlet.OnEvent(event)
	//if !ok {
	//	log.Println("Prospector outlet closed")
	//	return errors.New("Prospector outlet closed")
	//}
	p.states.Update(event.State)
	return nil
}

// 根据给定的 offset 开启一个新的 harvester，并以防打到 HarvesterLimit
func (p *Prospector) startHarvester(state lcFile.State, offset int64) error {
	// harvester_limit 项限制一个 prospector 并行启动的 harvester 数量，直接影响文件打开数
	if (atomic.LoadUint64(&p.harversterCounter) > 10) {
		harvesterSkipped.Add(1)
		return fmt.Errorf("Harvester limit reached.")
	}
	state.Offset = offset
	// state.Finished 设置为 false 表明一个 harvester running
	state.Finished = false

	// 根据 state 创建一个 harvester
	h, err := p.createHarvester(state)
	if err != nil {
		return err
	}
	reader, err := h.Setup()
	if err != nil {
		return fmt.Errorf("Error setting up harvester: %s", err)
	}
	// State 直接更新，不通过 channel 进行状态更新
	// State 仅在 setup 成功完成后更新
	err = p.updateState(NewEvent(state))
	if err != nil {
		return err
	}
	p.wg.Add(1)
	// startHarvester 不是并行的，但是需要源自操作来减少接下来的 goroutine 中的 counter
	atomic.AddUint64(&p.harversterCounter, 1)
	go func() {
		defer func() {
			// ^ 异或
			atomic.AddUint64(&p.harversterCounter, ^uint64(0))
			p.wg.Done()
		}()

		// 启动 harvester 并且选择正确的 type
		h.Harvest(reader)
	}()
	return nil
}
