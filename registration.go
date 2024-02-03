package pegasus

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	di "github.com/nodejayes/generic-di"
)

type (
	ActionHandler interface {
		GetActionType() string
		Handler(msg Message, ctx *gin.Context)
	}
	Config struct {
		EventUrl          string
		ActionUrl         string
		ClientIDHeaderKey string
		Handlers          []ActionHandler
	}
	Response struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
	}
)

func Register(router *gin.Engine, config *Config) {
	router.GET(config.EventUrl, func(ctx *gin.Context) {
		clientStore := di.Inject[ClientStore]()
		clientID, ok := ctx.GetQuery(config.ClientIDHeaderKey)
		_, err := uuid.Parse(clientID)
		if !ok || err != nil {
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
