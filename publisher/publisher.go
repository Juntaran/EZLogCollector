/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/25 22:22
  */

package publisher

import (
	"errors"
	"expvar"
	"github.com/Juntaran/EZLogCollector/prospector"
	"github.com/elastic/beats/libbeat/publisher"
	"log"
)

var (
	eventsSent = expvar.NewInt("publish.events")
)

// LogPublisher provides functionality to start and stop a publisher worker.
type LogPublisher interface {
	Start()
	Stop()
}

// SuccessLogger is used to report successfully published events.
type SuccessLogger interface {

	// Published will be run after events have been acknowledged by the outputs.
	Published(events []*prospector.Event) bool
}

func New(async bool, in chan []*prospector.Event, out SuccessLogger, pub publisher.Publisher) LogPublisher {
	if async {
		log.Println("Using publish_async is experimental!")
		return newAsyncLogPublisher(in, out, pub)
	}
	return newSyncLogPublisher(in, out, pub)
}

var (
	sigPublisherStop = errors.New("publisher was stopped")
)

//// getDataEvents returns all events which contain data (not only state updates)
//func getDataEvents(events []*input.Event) []common.MapStr {
//	dataEvents := make([]common.MapStr, 0, len(events))
//	for _, event := range events {
//		if event.HasData() {
//			dataEvents = append(dataEvents, event.ToMapStr())
//		}
//	}
//	return dataEvents
//}


type Publisher interface {
	Connect() Client
}

type Client interface {
	// Close disconnects the Client from the publisher pipeline.
	Close() error

	// PublishEvent publishes one event with given options. If Sync option is set,
	// PublishEvent will block until output plugins report success or failure state
	// being returned by this method.
	PublishEvent(event mapstr, opts ...ClientOption) bool

	// PublishEvents publishes multiple events with given options. If Guaranteed
	// option is set, PublishEvent will block until output plugins report
	// success or failure state being returned by this method.
	PublishEvents(events []common.MapStr, opts ...ClientOption) bool
}