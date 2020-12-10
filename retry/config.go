package retry

import (
	"github.com/gojekfarm/ziggurat/zlog"
	"github.com/gojekfarm/ziggurat/ztype"
	"strings"
)

type RabbitMQConfig struct {
	Hosts                string `mapstructure:"hosts"`
	DelayQueueExpiration string `mapstructure:"delay-queue-expiration"`
	DialTimeoutInS       int    `mapstructure:"dial-timeout-seconds"`
}

func parseRabbitMQConfig(config ztype.ConfigStore) *RabbitMQConfig {
	rmqcfg := &RabbitMQConfig{}
	if err := config.UnmarshalByKey("rabbitmq", rmqcfg); err != nil {
		zlog.LogError(err, "rmq config unmarshall error", nil)
		return &RabbitMQConfig{
			Hosts:                "amqp://user:bitnami@localhost:5672/",
			DelayQueueExpiration: "2000",
		}
	}
	return rmqcfg
}

func splitHosts(hosts string) []string {
	return strings.Split(hosts, ",")
}
