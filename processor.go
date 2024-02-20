package pegasus

import (
	"fmt"
	"net/http"
)

type processor struct {
	handlers map[string]func(msg Message, res http.ResponseWriter, req *http.Request)
}

func (ctx *processor) dispatch(message Message, res http.ResponseWriter, req *http.Request) error {
	handler := ctx.handlers[message.Type]
	if handler != nil {
		handler(message, res, req)
		return nil
	}
	return fmt.Errorf("handler %s not found", message.Type)
}

func (ctx *processor) registerHandlers(handlers []ActionHandler) {
	for _, handler := range handlers {
		ctx.handlers[handler.GetActionType()] = handler.Handler
	}
}

var actionProcessor = &processor{
	handlers: make(map[string]func(msg Message, res http.ResponseWriter, req *http.Request)),
}
