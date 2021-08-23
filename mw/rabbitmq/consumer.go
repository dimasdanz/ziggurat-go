package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gojekfarm/ziggurat"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/makasim/amqpextra/logger"
	"github.com/streadway/amqp"
)

func startConsumer(ctx context.Context, d *amqpextra.Dialer, c QueueConfig, h ziggurat.Handler, l logger.Logger, ogl ziggurat.StructuredLogger) (*consumer.Consumer, error) {
	pfc := 1
	wc := 1

	if c.ConsumerPrefetchCount > 1 {
		pfc = c.ConsumerPrefetchCount
	}
	if c.WorkerCount > 1 {
		wc = c.WorkerCount
	}

	qname := fmt.Sprintf("%s_%s_%s", c.QueueName, "instant", "queue")
	parallelWorker := consumer.NewParallelWorker(wc)
	consumerName := fmt.Sprintf("%s_consumer", c.QueueName)
	cons, err := d.Consumer(
		consumer.WithContext(ctx),
		consumer.WithQueue(qname),
		consumer.WithLogger(l),
		consumer.WithQos(pfc, false),
		consumer.WithWorker(parallelWorker),
		consumer.WithHandler(consumer.HandlerFunc(func(ctx context.Context, msg amqp.Delivery) interface{} {
			bb := msg.Body
			var event ziggurat.Event
			err := json.Unmarshal(bb, &event)
			if err != nil {
				ogl.Error("error rejecting message:", err)
				return msg.Reject(true)
			}
			ogl.Info("amqp processing message", map[string]interface{}{"consumer": consumerName}, event.Metadata)
			err = h.Handle(ctx, &event)
			if err != nil {
				ogl.Error("error processing message", err)
			}
			return msg.Ack(true)
		})),
	)

	if err != nil {
		return nil, err
	}
	return cons, nil
}
