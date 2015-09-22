package publisher

// TODO: Add test cases that wait for the Completed() signal. Currently
// this functionality is not implemented.

import (
	"testing"

	"github.com/elastic/libbeat/common"
	"github.com/stretchr/testify/assert"
)

// Fails with nil pointer dereference in:
// github.com/elastic/libbeat/outputs.(*CompositeSignal).Completed(0x0)
//    github.com/elastic/libbeat/outputs/signal.go:88
func TestAsyncPublishEvent(t *testing.T) {
	testPub := newTestPublisherNoBulk(CompletedResponse)
	event := testEvent()
	// Async PublishEvent always immediately returns true.
	assert.True(t, testPub.pub.asyncPublisher.client().PublishEvent(event))
	msgs, err := testPub.outputMsgHandler.waitForMessages(1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, event, msgs[0].event)
}

func TestAsyncPublishEvents(t *testing.T) {
	testPub := newTestPublisherNoBulk(CompletedResponse)
	events := []common.MapStr{testEvent(), testEvent()}
	testPub.pub.asyncPublisher.client().PublishEvents(events)
	msgs, err := testPub.outputMsgHandler.waitForMessages(1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, events[0], msgs[0].events[0])
	assert.Equal(t, events[1], msgs[0].events[1])
}

// Fails with nil pointer dereference in:
// github.com/elastic/libbeat/outputs.(*CompositeSignal).Completed(0x0)
//    github.com/elastic/libbeat/outputs/signal.go:88
func TestBulkAsyncPublishEvent(t *testing.T) {
	t.SkipNow()
	testPub := newTestPublisherWithBulk(CompletedResponse)
	event := testEvent()
	// Async PublishEvent always immediately returns true.
	assert.True(t, testPub.pub.asyncPublisher.client().PublishEvent(event))
	msgs, err := testPub.outputMsgHandler.waitForMessages(1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, event, msgs[0].event)
}

func TestBulkAsyncPublishEvents(t *testing.T) {
	t.SkipNow()
	testPub := newTestPublisherWithBulk(CompletedResponse)
	events := []common.MapStr{testEvent(), testEvent()}
	testPub.pub.asyncPublisher.client().PublishEvents(events)
	msgs, err := testPub.outputMsgHandler.waitForMessages(1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, events[0], msgs[0].events[0])
	assert.Equal(t, events[1], msgs[0].events[1])
}
