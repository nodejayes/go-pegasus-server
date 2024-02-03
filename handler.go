package pegasus

import (
	"encoding/json"

	di "github.com/nodejayes/generic-di"
)

func init() {
	di.Injectable(newEventHandler)
}

type (
	Message struct {
		Type    string `json:"type"`
		Payload any    `json:"payload"`
	}
	ChannelMessage struct {
		ClientFilter func(client Client) bool
		Message      string
	}
	EventHander struct {
		channel chan ChannelMessage
	}
)

func newEventHandler() *EventHander {
	return &EventHander{
		channel: make(chan ChannelMessage),
	}
}

func (ctx *EventHander) getChannel() chan ChannelMessage {
	return ctx.channel
}

func (ctx *EventHander) SendMessage(clientFilter func(client Client) bool, message Message) error {
	buffer, err := json.Marshal(message)
	if err != nil {
		return err
	}
	ctx.channel <- ChannelMessage{
		ClientFilter: clientFilter,
		Message:      string(buffer),
	}
	return nil
}
