package publisher

// TODO: Add test cases that wait for the Completed() signal. Currently
// this functionality is not implemented.

import (
	"testing"

	"github.com/elastic/libbeat/common"
	"github.com/stretchr/testify/assert"
)

func TestAsyncPublishEvent(t *testing.T) {
	// init
	testPub := newTestPublisherNoBulk(CompletedResponse)
	event := testEvent()

	// execute
	// Async PublishEvent always immediately returns true.
	ok := testPub.pub.asyncPublisher.client().PublishEvent(event)
	assert.True(t, ok)

	// validate
	msgs, err := testPub.outputMsgHandler.waitForMessages(1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, event, msgs[0].event)
}

func TestAsyncPublishEvents(t *testing.T) {
	// init
	testPub := newTestPublisherNoBulk(CompletedResponse)
	events := []common.MapStr{testEvent(), testEvent()}

	// execute
	// Async PublishEvent always immediately returns true.
	ok := testPub.pub.asyncPublisher.client().PublishEvents(events)
	assert.True(t, ok)

	// validate
	msgs, err := testPub.outputMsgHandler.waitForMessages(1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, events[0], msgs[0].events[0])
	assert.Equal(t, events[1], msgs[0].events[1])
}

func TestBulkAsyncPublishEvent(t *testing.T) {
	t.SkipNow()

	// init
	testPub := newTestPublisherWithBulk(CompletedResponse)
	event := testEvent()

	// Async PublishEvent always immediately returns true.
	ok := testPub.pub.asyncPublisher.client().PublishEvent(event)
	assert.True(t, ok)

	// validate
	msgs, err := testPub.outputMsgHandler.waitForMessages(1)
	if err != nil {
		t.Fatal(err)
	}
	// Bulk outputer always sends bulk messages (even if only one event is
	// present)
	assert.Equal(t, event, msgs[0].events[0])
}

func TestBulkAsyncPublishEvents(t *testing.T) {
	t.SkipNow()

	// init
	testPub := newTestPublisherWithBulk(CompletedResponse)
	events := []common.MapStr{testEvent(), testEvent()}

	// Async PublishEvent always immediately returns true.
	ok := testPub.pub.asyncPublisher.client().PublishEvents(events)
	assert.True(t, ok)

	msgs, err := testPub.outputMsgHandler.waitForMessages(1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, events[0], msgs[0].events[0])
	assert.Equal(t, events[1], msgs[0].events[1])
}
