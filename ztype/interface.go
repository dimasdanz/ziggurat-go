package ztype

import (
	"context"
	"github.com/gojekfarm/ziggurat/zbase"
	"net/http"
)

type StartStopper interface {
	Start(a App) error
	Stop(a App)
}

type Server interface {
	ConfigureRoutes(a App, configFunc func(a App, h http.Handler))
	Handler() http.Handler
	StartStopper
}

type MetricPublisher interface {
	StartStopper
	IncCounter(metricName string, value int64, arguments map[string]string) error
	Gauge(metricName string, value int64, arguments map[string]string) error
}

type MessageRetry interface {
	Retry(app App, payload zbase.MessageEvent) error
	Replay(app App, topicEntity string, count int) error
	StartStopper
}

type App interface {
	Context() context.Context
	Routes() []zbase.Stream
	MessageRetry() MessageRetry
	Handler() MessageHandler
	MetricPublisher() MetricPublisher
	HTTPServer() Server
}

type MessageHandler interface {
	HandleMessage(event zbase.MessageEvent, app App) ProcessStatus
}

type Streams interface {
	Start(a App) (chan struct{}, error)
	Stop()
}
