package gin

type Options struct {
	Codec        Codec
	ServiceName  string
	ProtoName    string
	Prefix       string
	interceptors []Interceptor
}

type Option func(*Options)

func WithServiceName(name string) Option {
	return func(o *Options) {
		o.ServiceName = name
	}
}

func WithProtoName(name string) Option {
	return func(o *Options) {
		o.ProtoName = name
	}
}

func WithCodec(codec Codec) Option {
	return func(o *Options) {
		o.Codec = codec
	}
}

func WithInterceptor(interceptors ...Interceptor) Option {
	return func(o *Options) {
		o.interceptors = interceptors
	}
}

func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.Prefix = prefix
	}
}
