package http

import (
	"fmt"
	netutil "github.com/sujunbo/micro/util/net"
	"log"
	"net"
	"net/http"
)

type HTTPServer struct {
	*http.Server
	*http.ServeMux
	opts ServerOptions
}

func NewServer(opts ...ServerOption) *HTTPServer {
	var opt ServerOptions
	for _, o := range opts {
		o(&opt)
	}

	mux := http.NewServeMux()
	return &HTTPServer{
		opts:     opt,
		ServeMux: mux,
		Server: &http.Server{
			Handler: mux,
		},
	}
}

func (s *HTTPServer) Start() {
	ls, err := netutil.ListenAddr(s.opts.addr, func(addr string) (net.Listener, error) {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			fmt.Printf("err:%v addr:%v\n", err, addr)
			return nil, err
		}
		return netutil.TcpKeepAliveListener{ln.(*net.TCPListener)}, nil
	})
	if err != nil {
		panic(err)
	}

	log.Println("HTTPServer listen on:", ls.Addr().String())
	s.opts.addr = ls.Addr().String()
	go s.Serve(ls)
}
