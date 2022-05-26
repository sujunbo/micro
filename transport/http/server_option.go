package http

type ServerOptions struct {
	addr string
}

type ServerOption func(*ServerOptions)

func Addr(addr string) ServerOption {
	return func(o *ServerOptions) {
		o.addr = addr
	}
}
