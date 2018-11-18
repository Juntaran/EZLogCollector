/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/19 02:29
  */

package prospector

import (
	"github.com/Juntaran/EZLogCollector/harvester"
	"sync"

	"github.com/Juntaran/EZLogCollector/harvester/lcFile"
)

/*
	prospector 功能:
	1. 解析配置，默认读取的日志是 nginx log
		conf: 	NginxPath		nginx.conf 路径
				NginxFormat 	nginx 格式，默认 main
				CloseInactive	不活跃关闭时间
				Backoff 		在达到 EOF 之后再次检查文件之间等待的时间
				MaxBackoff		在达到 EOF 之后再次检查文件之前等待的最长时间
				backoff_factor	指定 backoff 尝试等待时间几次，默认是 2
				buffer_size		每次 harvester 读取文件缓冲字节数，默认是 16384
		还要支持正则 * 的使用
	2. 获取日志信息
		对每一个文件生成一个 harvester 读取日志
		构造一个 chan，保存每个 harvester 返回的信息
*/
type Prospector struct {
	prospectorer 		Prospectorer
	outlet 				Outlet
	harvesterChan		chan *Event
	done 				chan struct{}
	states 				*lcFile.States
	wg 					sync.WaitGroup
	harversterCounter 	uint64
}

type Prospectorer interface {
	Init()
	Run()
}

type Outlet interface {
	OnEvent(event *Event) bool
}

func NewProspector(states lcFile.States, outlet Outlet) (*Prospector, error) {
	prospector := &Prospector{
		outlet: 		outlet,
		harvesterChan: 	make(chan *Event),
		done:			make(chan struct{}),
		states:			states.Copy(),
		wg:				sync.WaitGroup{},
	}

	if err := prospector.Init(); err != nil {
		return nil, err
	}
	return prospector, nil
}

func (p *Prospector) Init() error {
	var prospectorer 	Prospectorer
	prospectorer.Init()
	p.prospectorer = prospectorer
	_, err := p.createHarvester(lcFile.State{})
	if err != nil {
		return err
	}
	return nil
}

func (p *Prospector) Run() {

}

func (p *Prospector) createHarvester(state lcFile.State) (*harvester.Harvester, error) {
	h := harvester.NewHarvester(
		//p.cfg,
		state,
		//p.harvesterChan,
		p.done,
	)
	return h, nil
}