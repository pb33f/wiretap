// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package broadcast

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/transaction"
)

type Broadcaster interface {
	RequestValidationErrors(
		request *model.Request,
		errors []*shared.WiretapValidationError,
		txn *transaction.HttpTransaction,
	)
	Request(request *model.Request, txn *transaction.HttpTransaction)
	Response(request *model.Request, txn *transaction.HttpTransaction)
	ResponseError(request *model.Request, txn *transaction.HttpTransaction, err error)
	ResponseValidationErrors(
		request *model.Request,
		txn *transaction.HttpTransaction,
		errors []*shared.WiretapValidationError,
	)
}

type LazyBroadcaster struct {
	lock     sync.RWMutex
	delegate Broadcaster
}

func NewLazyBroadcaster() *LazyBroadcaster {
	return &LazyBroadcaster{}
}

func (b *LazyBroadcaster) Set(delegate Broadcaster) {
	if delegate == nil {
		panic("wiretap broadcaster delegate cannot be nil")
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	b.delegate = delegate
}

func (b *LazyBroadcaster) Ready() bool {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.delegate != nil
}

func (b *LazyBroadcaster) current() Broadcaster {
	b.lock.RLock()
	defer b.lock.RUnlock()
	if b.delegate == nil {
		panic("wiretap broadcaster has not been initialized")
	}
	return b.delegate
}

func (b *LazyBroadcaster) RequestValidationErrors(
	request *model.Request,
	errors []*shared.WiretapValidationError,
	txn *transaction.HttpTransaction,
) {
	b.current().RequestValidationErrors(request, errors, txn)
}

func (b *LazyBroadcaster) Request(request *model.Request, txn *transaction.HttpTransaction) {
	b.current().Request(request, txn)
}

func (b *LazyBroadcaster) Response(request *model.Request, txn *transaction.HttpTransaction) {
	b.current().Response(request, txn)
}

func (b *LazyBroadcaster) ResponseError(request *model.Request, txn *transaction.HttpTransaction, err error) {
	b.current().ResponseError(request, txn, err)
}

func (b *LazyBroadcaster) ResponseValidationErrors(
	request *model.Request,
	txn *transaction.HttpTransaction,
	errors []*shared.WiretapValidationError,
) {
	b.current().ResponseValidationErrors(request, txn, errors)
}

type channelBroadcaster struct {
	channel     *bus.Channel
	channelName string
	logger      *slog.Logger
}

func NewBroadcaster(channel *bus.Channel, channelName string) Broadcaster {
	if channel == nil {
		panic("wiretap broadcast channel cannot be nil")
	}
	if channelName == "" {
		panic("wiretap broadcast channel name cannot be empty")
	}
	return &channelBroadcaster{
		channel:     channel,
		channelName: channelName,
		logger:      slog.Default(),
	}
}

func (b *channelBroadcaster) RequestValidationErrors(
	request *model.Request,
	errors []*shared.WiretapValidationError,
	txn *transaction.HttpTransaction,
) {
	txn.RequestValidation = errors
	b.sendTransaction(request, txn)
}

func (b *channelBroadcaster) Request(request *model.Request, txn *transaction.HttpTransaction) {
	b.sendTransaction(request, txn)
}

func (b *channelBroadcaster) Response(request *model.Request, txn *transaction.HttpTransaction) {
	b.sendTransaction(request, txn)
}

func (b *channelBroadcaster) ResponseError(request *model.Request, txn *transaction.HttpTransaction, err error) {
	b.send(request, txn, err)
}

func (b *channelBroadcaster) ResponseValidationErrors(
	request *model.Request,
	txn *transaction.HttpTransaction,
	errors []*shared.WiretapValidationError,
) {
	txn.ResponseValidation = errors
	b.sendTransaction(request, txn)
}

func (b *channelBroadcaster) sendTransaction(request *model.Request, txn *transaction.HttpTransaction) {
	payload, err := json.Marshal(txn)
	if err != nil {
		b.warn("dropping wiretap broadcast; unable to marshal payload", "error", err)
		return
	}
	b.send(request, payload, nil)
}

func (b *channelBroadcaster) send(request *model.Request, payload any, err error) {
	if request == nil || request.Id == nil {
		b.warn("dropping wiretap broadcast; request id is missing")
		return
	}
	id, uuidErr := uuid.NewUUID()
	if uuidErr != nil {
		b.warn("dropping wiretap broadcast; unable to create message id", "error", uuidErr)
		return
	}
	b.channel.Send(&model.Message{
		Id:            &id,
		DestinationId: request.Id,
		Error:         err,
		Channel:       b.channelName,
		Destination:   b.channelName,
		Payload:       payload,
		Direction:     model.ResponseDir,
	})
}

func (b *channelBroadcaster) warn(message string, attrs ...any) {
	if b.logger == nil {
		return
	}
	b.logger.Warn(message, attrs...)
}
