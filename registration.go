package pegasus

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	di "github.com/nodejayes/generic-di"
)

type (
	Config struct {
		EventUrl          string
		ActionUrl         string
		ClientIDHeaderKey string
		Handlers          map[string]func(msg Message, ctx *gin.Context)
	}
	Response struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
	}
)

func Register(router *gin.Engine, config *Config) {
	router.GET(config.EventUrl, func(ctx *gin.Context) {
		clientStore := di.Inject[ClientStore]()
		clientID := ctx.Param(config.ClientIDHeaderKey)
		if len(clientID) < 1 {
			ctx.JSON(http.StatusBadRequest, Response{
				Code:  http.StatusBadRequest,
				Error: "clientId not found in header",
			})
			return
		}
		clientStore.Add(Client{
			ID:      clientID,
			Context: ctx,
		})
		ctx.Stream(func(w io.Writer) bool {
			if msg, ok := <-di.Inject[EventHander]().getChannel(); ok {
				client := clientStore.Get(msg.ClientFilter)
				if len(client) < 1 {
					return true
				}
				ctx.SSEvent("message", msg.Message)
				return true
			}
			return false
		})
	})
	router.POST(config.ActionUrl, func(ctx *gin.Context) {
		var msg Message
		err := ctx.BindJSON(&msg)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, Response{
				Code:  http.StatusInternalServerError,
				Error: err.Error(),
			})
		}
		err = actionProcessor.dispatch(msg, ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, Response{
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
