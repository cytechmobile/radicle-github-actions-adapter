package readerwriterbroker

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"radicle-github-actions-adapter/app/broker"
	"strconv"
	"strings"
)

type ReaderWriterBroker struct {
	brokerReader io.Reader
	brokerWriter io.Writer
	logger       *slog.Logger
}

func NewReaderWriterBroker(reader io.Reader, writer io.Writer, logger *slog.Logger) *ReaderWriterBroker {
	return &ReaderWriterBroker{
		brokerReader: reader,
		brokerWriter: writer,
		logger:       logger,
	}
}

// ParseRequestMessage reads a Request Message from broker message through ReaderWriterBroker.Reader
// and parses it in to an app.RequestMessage.
func (sb *ReaderWriterBroker) ParseRequestMessage(ctx context.Context) (*broker.RequestMessage, error) {
	input, err := io.ReadAll(sb.brokerReader)
	if err != nil {
		sb.logger.Error("could not read request message", "error", err.Error())
		return nil, err
	}
	input = []byte(strings.ReplaceAll(string(input), "\n", ""))
	sb.logger.Debug("received message from broker", "message", string(input))
	messageType, err := sb.parseRequestMessageType(input)
	if err != nil {
		sb.logger.Error("could not parse request message", "error", err.Error())
		return nil, err
	}
	requestMessage := broker.RequestMessage{}
	switch messageType {
	case broker.RequestMessageTypePush:
		requestMessageTypePush := broker.RequestPushEventMessage{}
		err := json.Unmarshal(input, &requestMessageTypePush)
		if err != nil {
			sb.logger.Error("could not unmarshal request message to push event message")
			return nil, errors.New("could not unmarshal request message to push event message")
		}
		requestMessage.PushEvent = &requestMessageTypePush
		requestMessage.Repo = requestMessageTypePush.Repository.ID
		requestMessage.Commit = requestMessageTypePush.After
		return &requestMessage, nil
	case broker.RequestMessageTypePatch:
		requestMessageTypePatch := broker.RequestPatchEventMessage{}
		err := json.Unmarshal(input, &requestMessageTypePatch)
		if err != nil {
			sb.logger.Error("could not unmarshal request message to patch event message")
			return nil, errors.New("could not unmarshal request message to patch event message")
		}
		requestMessage.PatchEvent = &requestMessageTypePatch
		requestMessage.Repo = requestMessageTypePatch.Repository.ID
		requestMessage.Commit = requestMessageTypePatch.Patch.After
		return &requestMessage, nil
	}
	sb.logger.Error("not supported event type", "event type", messageType)
	return nil, errors.New("not supported event type: " + string(messageType))
}

func (sb *ReaderWriterBroker) parseRequestMessageType(input []byte) (broker.RequestMessageType, error) {
	brokerMessage := broker.RequestTypeMessage{}
	err := json.Unmarshal(input, &brokerMessage)
	if err != nil {
		sb.logger.Error("could not unmarshal request message", "error", err.Error())
		return "", errors.New("could not unmarshal request")
	}
	if brokerMessage.Request != "trigger" {
		sb.logger.Error("not supported message request", "request", brokerMessage.Request)
		return "", errors.New("not supported message request: " + brokerMessage.Request)
	}
	if val, ok := broker.SupportedProtocolVersions[brokerMessage.Version]; !ok || !val {
		sb.logger.Error("not supported message protocol version", "version", brokerMessage.Version)
		return "", errors.New("not supported message protocol version: " + strconv.Itoa(int(brokerMessage.Version)))
	}
	switch brokerMessage.EventType {
	case broker.RequestMessageTypePush:
		return broker.RequestMessageTypePush, nil
	case broker.RequestMessageTypePatch:
		return broker.RequestMessageTypePatch, nil
	}
	sb.logger.Error("not supported event type", "event type", brokerMessage.EventType)
	return "", errors.New("not supported event type: " + string(brokerMessage.EventType))
}

// ServeResponse writes the responseMessage to the ReaderWriterBroker.Writer
func (sb *ReaderWriterBroker) ServeResponse(ctx context.Context, responseMessage broker.ResponseMessage) error {
	encoder := json.NewEncoder(sb.brokerWriter)
	return encoder.Encode(responseMessage)
}
