package pegasus

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	di "github.com/nodejayes/generic-di"
)

type (
	Config struct {
		EventUrl  string
		ActionUrl string
		Handlers  map[string]func(msg Message, ctx *gin.Context)
	}
	Response struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
	}
)

func Register(router *gin.Engine, config *Config) {
	router.GET(config.EventUrl, func(ctx *gin.Context) {
		ctx.Stream(func(w io.Writer) bool {
			if msg, ok := <-di.Inject[EventHander]().getChannel(); ok {
				ctx.SSEvent("message", msg)
				return true
			}
			return false
		})
	})
	router.POST(config.ActionUrl, func(ctx *gin.Context) {
		var msg Message
		err := ctx.BindJSON(&msg)
		if err != nil {
			ctx.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: err.Error(),
			})
		}
		err = actionProcessor.dispatch(msg, ctx)
		if err != nil {
			ctx.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: err.Error(),
			})
		}
		ctx.JSON(http.StatusOK, Response{
			Code:  http.StatusOK,
			Error: "",
		})
	})
	actionProcessor.registerHandlers(config.Handlers)
}
