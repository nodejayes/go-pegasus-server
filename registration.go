package pegasus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	di "github.com/nodejayes/generic-di"
)

type (
	ActionHandler interface {
		GetActionType() string
		Handler(msg Message, res http.ResponseWriter, req *http.Request)
	}
	Config struct {
		EventUrl             string
		ActionUrl            string
		ClientIDHeaderKey    string
		SendConnectedAfterMs int
		Handlers             []ActionHandler
	}
	Response struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
	}
)

func jsonResponse(res http.ResponseWriter, statusCode int, data any) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(statusCode)
	str, _ := json.Marshal(data)
	res.Write(str)
}

func formatMessage(event, data string) (string, error) {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("event: %s\n", event))
	sb.WriteString(fmt.Sprintf("data: %v\n\n", data))

	return sb.String(), nil
}

func Register(router *http.ServeMux, config *Config) {
	if config.SendConnectedAfterMs < 1 {
		config.SendConnectedAfterMs = 500
	}
	handleOutgoing(router, config)
	handleIncoming(router, config)
	actionProcessor.registerHandlers(config.Handlers)
}

func handleIncoming(router *http.ServeMux, config *Config) {
	router.HandleFunc(fmt.Sprintf("POST %s", config.ActionUrl), func(res http.ResponseWriter, req *http.Request) {
		var msg Message
		err := json.NewDecoder(req.Body).Decode(&msg)
		if err != nil {
			jsonResponse(res, http.StatusInternalServerError, Response{
				Code:  http.StatusInternalServerError,
				Error: err.Error(),
			})
			return
		}
		err = actionProcessor.dispatch(msg, res, req)
		if err != nil {
			jsonResponse(res, http.StatusInternalServerError, Response{
				Code:  http.StatusInternalServerError,
				Error: err.Error(),
			})
			return
		}
		jsonResponse(res, http.StatusOK, Response{
			Code:  http.StatusOK,
			Error: "",
		})
	})
}

func handleOutgoing(router *http.ServeMux, config *Config) {
	router.HandleFunc(fmt.Sprintf("GET %s", config.EventUrl), func(res http.ResponseWriter, req *http.Request) {
		rc := http.NewResponseController(res)

		clientStore := di.Inject[ClientStore]()
		clientID := req.URL.Query().Get(config.ClientIDHeaderKey)
		_, err := uuid.Parse(clientID)
		if err != nil {
			jsonResponse(res, http.StatusBadRequest, Response{
				Code:  http.StatusBadRequest,
				Error: "clientId not found in header",
			})
			return
		}
		clientStore.Add(Client{
			ID:       clientID,
			Response: res,
			Request:  req,
		})
		go func() {
			time.Sleep(time.Duration(config.SendConnectedAfterMs) * time.Millisecond)
			di.Inject[EventHander]().SendMessage(func(client Client) bool {
				return client.ID == clientID
			}, Message{
				Type:    "connected",
				Payload: clientID,
			})
		}()

		res.Header().Set("Content-Type", "text/event-stream")
		res.Header().Set("Cache-Control", "no-cache")
		res.Header().Set("Connection", "keep-alive")
		for msg := range di.Inject[EventHander]().getChannel() {
			client := clientStore.Get(msg.ClientFilter)
			if len(client) < 1 {
				break
			}
			data, err := formatMessage("message", msg.Message)
			if err != nil {
				fmt.Println(err.Error())
				break
			}
			fmt.Fprint(res, data)
			err = rc.Flush()
			if err != nil {
				return
			}
		}
	})
}
