package zig

import (
	"github.com/rs/zerolog/log"
	"time"
)

func messageHandler(app *App, handlerFunc HandlerFunc) func(event MessageEvent) {
	return func(event MessageEvent) {
		metricTags := map[string]string{
			"topic_entity": event.TopicEntity,
			"kafka_topic":  event.Topic,
		}
		funcExecStartTime := time.Now()
		status := handlerFunc(event, app)
		funcExecEndTime := time.Now()
		app.metricPublisher.Gauge("handler_func_exec_time", funcExecEndTime.Sub(funcExecStartTime).Milliseconds(), metricTags)
		switch status {
		case ProcessingSuccess:
			if publishErr := app.metricPublisher.IncCounter("message_processing_success", 1, metricTags); publishErr != nil {
				log.Error().Err(publishErr).Msg("")
			}
			log.Info().Msg("successfully processed message")
		case SkipMessage:
			if publishErr := app.metricPublisher.IncCounter("message_processing_failure_skip", 1, metricTags); publishErr != nil {
				log.Error().Err(publishErr).Msg("")
			}
			log.Info().Msg("skipping message")

		case RetryMessage:
			log.Info().Msgf("retrying message")
			if publishErr := app.metricPublisher.IncCounter("message_processing_failure_skip", 1, metricTags); publishErr != nil {
				log.Error().Err(publishErr).Msg("")
			}
			if retryErr := app.retrier.Retry(app, event); retryErr != nil {
				log.Error().Err(retryErr).Msg("error retrying message")
			}
		default:
			log.Error().Err(ErrInvalidReturnCode).Msg("return code must be one of `zig.ProcessingSuccess OR zig.RetryMessage OR zig.SkipMessage`")
		}
	}
}
