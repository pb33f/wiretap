// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package broadcast

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroadcasterSendsPreMarshaledTransactionPayload(t *testing.T) {
	eventBus := bus.NewEventBus()
	channelName := "broadcast-test-" + uuid.NewString()
	channel := eventBus.GetChannelManager().CreateChannel(channelName)
	messages := make(chan *model.Message, 1)

	_, err := eventBus.GetChannelManager().SubscribeChannelHandler(channelName, func(message *model.Message) {
		messages <- message
	}, true)
	require.NoError(t, err)

	requestID := uuid.New()
	txn := &transaction.HttpTransaction{
		Id: "txn-1",
		Request: &transaction.HttpRequest{
			Path: "/pets",
		},
	}

	NewBroadcaster(channel, channelName).Request(&model.Request{Id: &requestID}, txn)
	require.NoError(t, eventBus.GetChannelManager().WaitForChannel(channelName))

	var message *model.Message
	select {
	case message = <-messages:
	default:
		require.Fail(t, "expected broadcast message")
	}

	payload, ok := message.Payload.([]byte)
	require.True(t, ok)
	var decoded transaction.HttpTransaction
	require.NoError(t, json.Unmarshal(payload, &decoded))
	assert.Equal(t, "txn-1", decoded.Id)
	require.NotNil(t, decoded.Request)
	assert.Equal(t, "/pets", decoded.Request.Path)
	assert.Equal(t, requestID.String(), message.DestinationId.String())
	assert.Equal(t, channelName, message.Channel)
	assert.Equal(t, channelName, message.Destination)
	assert.Equal(t, model.ResponseDir, message.Direction)
}

func TestBroadcasterDropsUnmarshalableTransactionPayload(t *testing.T) {
	broadcaster, messages := newTestBroadcaster(t)
	requestID := uuid.New()
	txn := &transaction.HttpTransaction{
		Id: "txn-1",
		Request: &transaction.HttpRequest{
			Headers: map[string]any{
				"X-Bad": func() {},
			},
		},
	}

	assert.NotPanics(t, func() {
		broadcaster.Request(&model.Request{Id: &requestID}, txn)
	})
	assertNoBroadcast(t, messages)
}

func TestBroadcasterDropsMissingRequestID(t *testing.T) {
	broadcaster, messages := newTestBroadcaster(t)
	txn := &transaction.HttpTransaction{Id: "txn-1"}

	assert.NotPanics(t, func() {
		broadcaster.Response(&model.Request{}, txn)
	})
	assertNoBroadcast(t, messages)
}

func newTestBroadcaster(t *testing.T) (Broadcaster, chan *model.Message) {
	t.Helper()

	eventBus := bus.NewEventBus()
	channelName := "broadcast-test-" + uuid.NewString()
	channel := eventBus.GetChannelManager().CreateChannel(channelName)
	messages := make(chan *model.Message, 1)

	_, err := eventBus.GetChannelManager().SubscribeChannelHandler(channelName, func(message *model.Message) {
		messages <- message
	}, true)
	require.NoError(t, err)

	return NewBroadcaster(channel, channelName), messages
}

func assertNoBroadcast(t *testing.T, messages chan *model.Message) {
	t.Helper()

	select {
	case <-messages:
		require.Fail(t, "expected no broadcast message")
	default:
	}
}
