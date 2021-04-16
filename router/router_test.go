package router

import (
	"context"
	"reflect"
	"testing"

	"github.com/gojekfarm/ziggurat/mock"

	"github.com/gojekfarm/ziggurat"
)

func TestDefaultRouter_HandleMessage(t *testing.T) {
	dr := New()
	expectedEvent := mock.CreateMockEvent()
	expectedEvent.ValueFunc = func() []byte {
		return nil
	}
	expectedEvent.HeadersFunc = func() map[string]string {
		return map[string]string{ziggurat.HeaderMessageRoute: "bar"}
	}
	dr.HandleFunc("foo", func(ctx context.Context, event ziggurat.Event) interface{} {
		if !reflect.DeepEqual(event, expectedEvent) {
			t.Errorf("expected event %+v, got %+v", expectedEvent, event)
		}
		return nil
	})
	dr.Handle(context.Background(), mock.Event{
		ValueFunc: func() []byte {
			return nil
		},
		HeadersFunc: func() map[string]string {
			return map[string]string{ziggurat.HeaderMessageRoute: "bar"}
		},
	})
}

func TestDefaultRouter_NotFoundHandler(t *testing.T) {
	notFoundHandlerCalled := false
	dr := New(WithNotFoundHandler(func(ctx context.Context, event ziggurat.Event) interface{} {
		notFoundHandlerCalled = true
		return nil
	}))

	dr.HandleFunc("foo", func(ctx context.Context, event ziggurat.Event) interface{} {
		return nil
	})

	dr.Handle(context.Background(), mock.Event{
		ValueFunc: func() []byte {
			return []byte{}
		},
		HeadersFunc: func() map[string]string {
			return map[string]string{ziggurat.HeaderMessageRoute: "bar"}
		},
	})

	if !notFoundHandlerCalled {
		t.Errorf("expected %v got %v", true, notFoundHandlerCalled)
	}

}
