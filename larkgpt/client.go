package larkgpt

import (
	"fmt"
	"net/http"

	"github.com/chyroc/lark"
)

type Client struct {
	larkIns                   *larkClient
	chatGPTIns                *chatGPTClient
	metricsIns                IMetrics
	serverPort                string
	maintained                bool
	enableSessionForLarkGroup bool
}

type ClientConfig struct {
	// Lark
	AppID     string
	AppSecret string

	// ChatGPT
	ChatGPTAPIKey string
	ChatGPTAPIURL string
	Maintained    bool

	// server
	ServerPort                string
	Metrics                   IMetrics
	EnableSessionForLarkGroup bool // 给群聊的消息启动 session，session id 是消息的 root id
}

func New(config *ClientConfig) *Client {
	res := new(Client)

	res.metricsIns = config.Metrics
	if res.metricsIns == nil {
		res.metricsIns = new(noneMetrics)
	}

	res.larkIns = newLarkClient(lark.New(lark.WithAppCredential(config.AppID, config.AppSecret)), res.metricsIns)

	res.chatGPTIns = newChatGPTClient(config.ChatGPTAPIURL, config.ChatGPTAPIKey, res.metricsIns)

	res.serverPort = config.ServerPort

	res.maintained = config.Maintained
	res.enableSessionForLarkGroup = config.EnableSessionForLarkGroup

	return res
}

func (r *Client) Start() error {
	r.larkIns.cli.EventCallback.HandlerEventV2IMMessageReceiveV1(r.larkMessageReceiverHandler)

	http.HandleFunc("/event", func(w http.ResponseWriter, req *http.Request) {
		r.larkIns.cli.EventCallback.ListenCallback(req.Context(), req.Body, w)
	})

	fmt.Printf("start server: %s ...\n", r.serverPort)
	return http.ListenAndServe(fmt.Sprint("[::]:", r.serverPort), nil)
}
