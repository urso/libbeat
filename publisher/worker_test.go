package publisher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test sending events through the messageWorker.
func TestMessageWorkerSend(t *testing.T) {
	ws := &workerSignal{}
	ws.Init()
	mh := &testMessageHandler{msgs: make(chan message, 10), response: true}
	mw := newMessageWorker(ws, 10, mh)

	s1 := &testSignaler{}
	m1 := message{signal: s1}
	mw.send(m1)

	s2 := &testSignaler{}
	m2 := message{signal: s2}
	mw.send(m2)

	msgs, err := mh.waitForMessages(2)
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, msgs, m1)
	assert.True(t, s1.completed)
	assert.Contains(t, msgs, m2)
	assert.True(t, s2.completed)
	ws.stop()
	assert.True(t, mh.stopped)
}

// Test that stopQueue invokes the Failed callback on all events in the queue.
func TestMessageWorkerStopQueue(t *testing.T) {
	s1 := &testSignaler{}
	m1 := message{signal: s1}

	s2 := &testSignaler{}
	m2 := message{signal: s2}

	qu := make(chan message, 2)
	qu <- m1
	qu <- m2

	stopQueue(qu)
	assert.True(t, s1.failed)
	assert.True(t, s2.failed)
}
