package publisher

import (
	"fmt"
	"time"

	"github.com/elastic/libbeat/common"
	"github.com/elastic/libbeat/outputs"
)

type testMessageHandler struct {
	msgs     chan message
	response OutputResponse
	stopped  bool
}

var _ messageHandler = &testMessageHandler{}
var _ worker = &testMessageHandler{}

func (mh *testMessageHandler) onMessage(m message) {
	mh.msgs <- mh.acknowledgeMessage(m)
}

func (mh *testMessageHandler) onStop() {
	mh.stopped = true
}

func (mh *testMessageHandler) send(m message) {
	mh.msgs <- mh.acknowledgeMessage(m)
}

func (mh *testMessageHandler) acknowledgeMessage(m message) message {
	fmt.Println("testMessageHandler acknowledgeMessage", m)
	if mh.response == CompletedResponse {
		fmt.Println("Sending Completed signal for", m)
		outputs.SignalCompleted(m.signal)
	} else {
		fmt.Println("Sending Failed signal for", m)
		outputs.SignalFailed(m.signal)
	}
	return m
}

// Waits for n messages to be received and then returns. If n messages are not
// received within one second the method returns an error.
func (mh *testMessageHandler) waitForMessages(n int) ([]message, error) {
	var msgs []message
	for {
		select {
		case m := <-mh.msgs:
			msgs = append(msgs, m)
			if len(msgs) >= n {
				return msgs, nil
			}
		case <-time.After(10 * time.Second):
			return nil, fmt.Errorf("Expected %d messages but received %d.", n, len(msgs))
		}
	}
}

type testSignaler struct {
	completed bool
	failed    bool
}

var _ outputs.Signaler = &testSignaler{}

func (s *testSignaler) Completed() {
	fmt.Println("testSignaler completed")
	s.completed = true
}

func (s *testSignaler) Failed() {
	fmt.Println("testSignaler failed")
	s.failed = true
}

func testEvent() common.MapStr {
	event := common.MapStr{}
	event["timestamp"] = common.Time(time.Now())
	event["type"] = "test"
	event["src"] = &common.Endpoint{}
	event["dst"] = &common.Endpoint{}
	return event
}

type testPublisher struct {
	pub              *PublisherType
	outputMsgHandler *testMessageHandler
}

const (
	BulkOn  = true
	BulkOff = false
)

type OutputResponse bool

const (
	CompletedResponse OutputResponse = true
	FailedResponse    OutputResponse = false
)

func newTestPublisher(bulkSize int, response OutputResponse) *testPublisher {
	mh := &testMessageHandler{
		msgs:     make(chan message, 10),
		response: response,
	}

	ow := &outputWorker{}
	ow.config.Bulk_size = &bulkSize
	ow.handler = mh
	ws := workerSignal{}
	ow.messageWorker.init(&ws, 1000, mh)

	pub := &PublisherType{
		Output:   []*outputWorker{ow},
		wsOutput: ws,
	}
	pub.wsOutput.Init()
	pub.wsPublisher.Init()
	pub.syncPublisher = newSyncPublisher(pub)
	pub.asyncPublisher = newAsyncPublisher(pub)
	return &testPublisher{
		pub:              pub,
		outputMsgHandler: mh,
	}
}

func newTestPublisherWithBulk(response OutputResponse) *testPublisher {
	return newTestPublisher(defaultBulkSize, response)
}

func newTestPublisherNoBulk(response OutputResponse) *testPublisher {
	return newTestPublisher(-1, response)
}
