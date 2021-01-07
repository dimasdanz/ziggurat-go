package router

import "github.com/gojekfarm/ziggurat"

var PipeHandlers = func(funcs ...func(handler ziggurat.Handler) ziggurat.Handler) func(origHandler ziggurat.Handler) ziggurat.Handler {
	return func(next ziggurat.Handler) ziggurat.Handler {
		return ziggurat.HandlerFunc(func(event ziggurat.Event) ziggurat.ProcessStatus {
			var handlerResult = next
			lastIdx := len(funcs) - 1
			for i := lastIdx; i >= 0; i-- {
				f := funcs[i]
				if i == lastIdx {
					handlerResult = f(next)
				} else {
					handlerResult = f(handlerResult)
				}
			}
			return handlerResult.HandleEvent(event)
		})
	}
}