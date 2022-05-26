package gin

import (
	"context"
	"github.com/gin-gonic/gin"
)

type HandleFunc func(ctx context.Context, req interface{}) (reply interface{}, err error)

type Interceptor func(ctx context.Context, req interface{}, handler HandleFunc) (reply interface{}, err error)

var NopInterceptor = func(ctx context.Context, req interface{}, handler HandleFunc) (reply interface{}, err error) {
	return handler(ctx, req)
}

// 链式拦截器
func ChainInterceptors(interceptors ...Interceptor) Interceptor {
	return func(ctx context.Context, req interface{}, handler HandleFunc) (reply interface{}, err error) {
		buildChain := func(current Interceptor, next HandleFunc) HandleFunc {
			return func(currentCtx context.Context, currentReq interface{}) (reply interface{}, err error) {
				return current(currentCtx, currentReq, next)
			}
		}

		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = buildChain(interceptors[i], chain)
		}
		return chain(ctx, req)
	}
}

func ginInterceptors(interceptors ...Interceptor) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ChainInterceptors(interceptors...)
	}
}
