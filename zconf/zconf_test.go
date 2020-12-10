package zconf

import (
	"github.com/gojekfarm/ziggurat/zbase"
	"os"
	"reflect"
	"testing"
)

const testConfPath = "../config/config.test.yaml"

func TestViperConfig_Parse(t *testing.T) {
	vc := NewViperConfig()
	expectedConfig := zbase.Config{
		StreamRouter: map[string]zbase.StreamRouterConfig{
			"plain-text-log": {
				InstanceCount:    2,
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				GroupID:          "plain_text_consumer",
			},
		},
		LogLevel:    "debug",
		ServiceName: "test-app",
		Retry: zbase.RetryConfig{
			Enabled: true,
			Count:   5,
		},
		HTTPServer: zbase.HTTPServerConfig{
			Port: "8080",
		},
	}

	err := vc.Parse(zbase.CommandLineOptions{ConfigFilePath: testConfPath})
	if err != nil {
		t.Errorf("error parsing config %v", err.Error())
	}
	actualConfig := vc.Config()
	if !reflect.DeepEqual(expectedConfig, *actualConfig) {
		t.Errorf("expected config %+v, actual cfgReader %+v", expectedConfig, actualConfig)
	}

}

func TestViperConfig_EnvOverride(t *testing.T) {
	overriddenValue := "localhost:9094"
	vc := NewViperConfig()
	if err := os.Setenv("ZIGGURAT_STREAM_ROUTER_PLAIN_TEXT_LOG_BOOTSTRAP_SERVERS", overriddenValue); err != nil {
		t.Error(err)
	}
	vc.Parse(zbase.CommandLineOptions{ConfigFilePath: testConfPath})
	config := vc.Config()
	actualValue := config.StreamRouter["plain-text-log"].BootstrapServers
	if !(actualValue == overriddenValue) {
		t.Errorf("expected value of bootstrap servers to be %s but got %s", overriddenValue, actualValue)
	}
}

func TestViperConfig_GetByKey(t *testing.T) {
	vc := NewViperConfig()
	vc.Parse(zbase.CommandLineOptions{ConfigFilePath: testConfPath})
	expectedStatsDConf := map[string]interface{}{"host": "localhost:8125"}
	statsdCfg := vc.GetByKey("statsd").(map[string]interface{})

	if !reflect.DeepEqual(expectedStatsDConf, statsdCfg) {
		t.Errorf("expected %v got %v", expectedStatsDConf, statsdCfg)
	}
}
