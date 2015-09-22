package publisher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Start and stop the bulkWorker.
func TestBulkWorkerStartStop(t *testing.T) {
	tw := &testMessageHandler{msgs: make(chan message, 10)}
	ws := newWorkerSignal()
	defer ws.stop()
	_ = newBulkWorker(ws, 10, tw, 500*time.Millisecond, 10)
}

func TestBulkWorkerSend(t *testing.T) {
	t.Skip()

	mh := &testMessageHandler{
		response: CompletedResponse,
		msgs:     make(chan message, 10),
	}
	ws := newWorkerSignal()
	defer ws.stop()
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
