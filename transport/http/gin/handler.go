package gin

import (
	"fmt"
	"net/http"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	*gin.Engine
	opts Options
}

func NewHandler(opts ...Option) *Handler {
	opt := Options{
		interceptors: []Interceptor{NopInterceptor},
		Codec:        defaultCodec{},
	}

	for _, o := range opts {
		o(&opt)
	}

	handler := &Handler{
		Engine: gin.Default(),
		opts:   opt,
	}

	return handler
}

func (h *Handler) RegisterService(srv interface{}) {
	service, err := register(srv)
	if err != nil {
		panic(err)
	}

	pe, err := newProtoExtend(h.opts.ProtoName)
	if err != nil {
		panic(err)
	}

	h.opts.ServiceName = service.name
	prefix := h.opts.Prefix
	if prefix == "" {
		prefix = fmt.Sprintf("/%s/", h.opts.ServiceName)
	}
	for method, v := range service.methods {
		rules := pe.methodHttpRules(method)
		for _, rule := range rules {
			h.Handle(rule.method, path.Join(prefix, rule.pattern), h.ginHandler(v))
		}
	}
}

func (h *Handler) Option() Options {
	return h.opts
}

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idx := strings.LastIndex(r.URL.Path, "/")
	if idx < 0 {
		panic(fmt.Errorf("rpc: no method in path %q", r.URL.Path))
	}

	if s.opts.Prefix == "" {
		r.URL.Path = "/" + s.opts.ServiceName + r.URL.Path[idx:]
	}

	s.Engine.ServeHTTP(w, r)
}

func (h *Handler) ginHandler(methodSpec *ServiceMethod) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		r := ctx.Request
		md := NewMetadata(r.Header, "")
		md.Set("request-time", fmt.Sprintf("%d", time.Now().UnixNano()))
		md.Set("method-name", methodSpec.method.Name)
		md.Set("client-ip", ctx.ClientIP())
		md = Join(md, Metadata(r.URL.Query()), ginParams(ctx.Params))

		newCtx := NewContextFromMetadata(r.Context(), md)
		r = r.WithContext(newCtx)

		reply, err := h.doHandle(r, methodSpec)
		if err != nil {
			h.opts.Codec.Encode(r, ctx.Writer, err)
		} else {
			h.opts.Codec.Encode(r, ctx.Writer, reply)
		}
		return
	}
}

func (s *Handler) doHandle(r *http.Request, methodSpec *ServiceMethod) (reply interface{}, err error) {
	ctx := r.Context()

	reqValue, err := s.getRequestValue(methodSpec.ReqType, r)
	if err != nil {
		return
	}

	return ChainInterceptors(s.opts.interceptors...)(ctx, reqValue, methodSpec.call)
}

func (s *Handler) getRequestValue(p reflect.Type, r *http.Request) (value interface{}, err error) {
	reqValue := reflect.New(p)
	err = s.opts.Codec.Decode(r, reqValue.Interface())
	return reqValue.Interface(), err
}

func ToUnderLine(src string) string {
	var dest []byte
	for _, s := range src {
		if s >= 'A' && s <= 'Z' {
			dest = append(dest, byte('_'), byte(s+32))
		} else {
			dest = append(dest, byte(s))
		}
	}

	if dest[0] == '_' {
		dest = dest[1:]
	}
	return string(dest)
}

func getMethodName(path string) (method string) {
	return path[strings.LastIndex(path, "/")+1:]
}

func ginParams(params gin.Params) Metadata {
	md := Metadata{}
	for _, param := range params {
		md.Set(param.Key, param.Value)
	}

	return md
}
