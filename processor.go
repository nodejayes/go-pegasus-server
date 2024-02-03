package pegasus

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type processor struct {
	handlers map[string]func(msg Message, ctx *gin.Context)
}

func (ctx *processor) dispatch(message Message, context *gin.Context) error {
	handler := ctx.handlers[message.Type]
	if handler != nil {
		handler(message, context)
		return nil
	}
	return fmt.Errorf("handler %s not found", message.Type)
}

func (ctx *processor) registerHandlers(handlers map[string]func(msg Message, ctx *gin.Context)) {
	ctx.handlers = handlers
}

var actionProcessor = &processor{}
