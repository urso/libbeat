package publisher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Start and stop the bulkWorker.
func TestBulkWorkerStartStop(t *testing.T) {
	tw := &testMessageHandler{msgs: make(chan message, 10)}
	ws := &workerSignal{}
	ws.Init()
	bw := newBulkWorker(ws, 10, tw, 500*time.Millisecond, 10)
	bw.ws.stop()
	bw.ws.wg.Wait()
}

// fatal error: stack overflow
// github.com/elastic/libbeat/outputs.(*CompositeSignal).Completed(0xc20801e5a0)
//     github.com/elastic/libbeat/outputs/signal.go:83
func TestBulkWorkerSend(t *testing.T) {
	t.SkipNow()
	mh := &testMessageHandler{
		response: CompletedResponse,
		msgs:     make(chan message, 10)}
	ws := &workerSignal{}
	ws.Init()
	bw := newBulkWorker(ws, 10, mh, 500*time.Millisecond, 10)

	s := &testSignaler{}
	m := message{event: testEvent(), signal: s}
	bw.send(m)
	msgs, err := mh.waitForMessages(1)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, s.completed)
	assert.Equal(t, m.event, msgs[0].events[0])
}
