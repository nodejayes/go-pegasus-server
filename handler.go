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
	EventHander struct {
		channel chan string
	}
)

func newEventHandler() *EventHander {
	return &EventHander{
		channel: make(chan string),
	}
}

func (ctx *EventHander) getChannel() chan string {
	return ctx.channel
}

func (ctx *EventHander) SendMessage(message Message) error {
	buffer, err := json.Marshal(message)
	if err != nil {
		return err
	}
	ctx.channel <- string(buffer)
	return nil
}
