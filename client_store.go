package pegasus

import (
	"net/http"

	di "github.com/nodejayes/generic-di"
)

func init() {
	di.Injectable(newClientStore)
}

type (
	Client struct {
		ID string
		Response http.ResponseWriter
		Request *http.Request
	}
	ClientStore struct {
		Clients map[string]Client
	}
)

func newClientStore() *ClientStore {
	return &ClientStore{
		Clients: make(map[string]Client),
	}
}

func (ctx *ClientStore) Add(client Client) {
	ctx.Clients[client.ID] = client
}

func (ctx *ClientStore) Remove(client Client) {
	delete(ctx.Clients, client.ID)
}

func (ctx *ClientStore) Get(filter func(client Client) bool) []Client {
	result := make([]Client, 0)
	for _, cl := range ctx.Clients {
		if filter(cl) {
			result = append(result, cl)
		}
	}
	return result
}