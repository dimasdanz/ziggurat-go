package statsd

import (
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/gojekfarm/ziggurat"
	"runtime"
	"strings"
	"time"
)

type StatsDClient struct {
	client  statsd.Statter
	host    string
	prefix  string
	handler ziggurat.MessageHandler
}

func NewStatsD(opts ...func(s *StatsDClient)) *StatsDClient {
	s := &StatsDClient{}
	for _, opt := range opts {
		opt(s)
	}
	if s.prefix == "" {
		s.prefix = "ziggurat_statsd"
	}
	if s.host == "" {
		s.host = "localhost:8125"
	}
	return s
}

func constructTags(tags map[string]string) string {
	var tagSlice []string
	for k, v := range tags {
		tagSlice = append(tagSlice, k+"="+v)
	}
	return strings.Join(tagSlice, ",")
}

func (s *StatsDClient) Start(app ziggurat.App) error {
	config := &statsd.ClientConfig{
		Prefix:  s.prefix,
		Address: s.host,
	}
	client, clientErr := statsd.NewClientWithConfig(config)
	if clientErr != nil {
		ziggurat.LogError(clientErr, "ziggurat statsD", nil)
		return clientErr
	}
	s.client = client
	go func() {
		ziggurat.LogInfo("statsd: starting go-routine publisher", nil)
		done := app.Context().Done()
		t := time.NewTicker(10 * time.Second)
		tickerChan := t.C
		for {
			select {
			case <-done:
				t.Stop()
				ziggurat.LogInfo("statsd: halting go-routine publisher", nil)
				return
			case <-tickerChan:
				s.client.Gauge("go_routine_count", int64(runtime.NumGoroutine()), 1.0)
			}
		}
	}()
	return nil
}

func (s *StatsDClient) Stop() {
	if s.client != nil {
		ziggurat.LogError(s.client.Close(), "error stopping statsd client", nil)
	}
}

func (s *StatsDClient) constructFullMetricStr(metricName, tags string) string {
	return metricName + "," + tags + "," + "app_name=" + s.prefix
}

func (s *StatsDClient) IncCounter(metricName string, value int64, arguments map[string]string) error {
	tags := constructTags(arguments)
	finalMetricName := s.constructFullMetricStr(metricName, tags)

	return s.client.Inc(finalMetricName, value, 1.0)
}

func (s *StatsDClient) Gauge(metricName string, value int64, arguments map[string]string) error {
	tags := constructTags(arguments)
	finalMetricName := s.constructFullMetricStr(metricName, tags)
	return s.client.Gauge(finalMetricName, value, 1.0)
}

func (s *StatsDClient) PublishHandlerMetrics(handler ziggurat.MessageHandler) ziggurat.MessageHandler {
	return ziggurat.HandlerFunc(func(messageEvent ziggurat.MessageEvent, app ziggurat.App) ziggurat.ProcessStatus {
		arguments := map[string]string{"route": messageEvent.StreamRoute}
		startTime := time.Now()
		status := handler.HandleMessage(messageEvent, app)
		endTime := time.Now()
		diffTimeInMS := endTime.Sub(startTime).Milliseconds()
		s.Gauge("handler_func_exec_time", diffTimeInMS, arguments)
		switch status {
		case ziggurat.RetryMessage, ziggurat.SkipMessage:
			s.IncCounter("message_processing_failure_skip_count", 1, arguments)
		default:
			s.IncCounter("message_processing_success_count", 1, arguments)
		}
		return status
	})
}

func (s *StatsDClient) PublishKafkaLag(handler ziggurat.MessageHandler) ziggurat.MessageHandler {
	return ziggurat.HandlerFunc(func(messageEvent ziggurat.MessageEvent, app ziggurat.App) ziggurat.ProcessStatus {
		actualTS := messageEvent.ActualTimestamp
		now := time.Now()
		diff := now.Sub(actualTS).Milliseconds()
		s.Gauge("kafka_message_lag", diff, map[string]string{
			"route": messageEvent.StreamRoute,
		})
		return handler.HandleMessage(messageEvent, app)
	})
}