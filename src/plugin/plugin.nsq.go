package plugin

import (
	"time"

	_nsq "github.com/nsqio/go-nsq"
)

type NsqConfig struct {
	Url     string `form:"url" bson:"url" json:"url"`
	Channel string `form:"channel" bson:"channel" json:"channel"`
}

type NsqClient struct {
	Client *_nsq.Producer
	Config NsqConfig
}

func NewNsqClient(config NsqConfig) (nsqClient NsqClient) {

	nsqClient.Config = config
	return
}

func (nsq *NsqClient) Connect() (err error) {

	nsq.Client, err = _nsq.NewProducer(nsq.Config.Url, _nsq.NewConfig())
	if nil != err {
		return
	}

	nsq.Client.SetLogger(nil, 0)

	return
}

func (nsq *NsqClient) Publish(topic string, message string) (err error) {

	err = nsq.Client.Publish(topic, []byte(message))

	return
}

type NsqServer struct {
	Client *_nsq.Consumer
	Config NsqConfig
}

func NewNsqServer(config NsqConfig) (nsqServer NsqServer) {

	nsqServer.Config = config
	return
}

type nsqHandler struct {
	handle func(string)
}

func (n *nsqHandler) HandleMessage(msg *_nsq.Message) error {
	n.handle(string(msg.Body))
	return nil
}

func (nsq *NsqServer) Subscribe(topic string, handle func(string)) (err error) {

	cfg := _nsq.NewConfig()

	cfg.LookupdPollInterval = 15 * time.Second

	nsq.Client, err = _nsq.NewConsumer(topic, nsq.Config.Channel, cfg)
	if err != nil {
		return
	}

	nsq.Client.SetLogger(nil, 0)

	nsq.Client.AddHandler(&nsqHandler{handle: handle})

	// 连接 nsqd
	err = nsq.Client.ConnectToNSQD(nsq.Config.Url)

	return
}
