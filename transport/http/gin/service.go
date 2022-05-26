package gin

import (
	"fmt"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"reflect"
	"unicode"
	"unicode/utf8"
)

var (
	// Precompute the reflect type for error.
	typeOfOsError = reflect.TypeOf((*error)(nil)).Elem()
	// Same as above, this time for http.Request.
	typeOfRequest = reflect.TypeOf((*http.Request)(nil)).Elem()
	// Precompute the reflect type for context.Context.
	typeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()
)

// RPCService represents a service registered with a specific Server.
type RPCService struct {
	name     string                    // name of service
	rcvr     reflect.Value             // receiver of methods for the service
	rcvrType reflect.Type              // type of the receiver
	methods  map[string]*ServiceMethod // registered methods
}

// Name returns service method name
// TODO: remove or use info.Name here?
func (s *RPCService) Name() string {
	return s.name
}

// Methods returns a slice of all service's registered methods
func (s *RPCService) Methods() []*ServiceMethod {
	items := make([]*ServiceMethod, 0, len(s.methods))
	for _, m := range s.methods {
		items = append(items, m)
	}
	return items
}

// MethodByName returns a ServiceMethod of a registered service's method or nil.
func (s *RPCService) MethodByName(name string) *ServiceMethod {
	return s.methods[name]
}

// ServiceMethod is what represents a method of a registered service
type ServiceMethod struct {
	// Type of the request data structure
	ReqType reflect.Type
	// Type of the response data structure
	RespType reflect.Type
	// method's receiver
	method *reflect.Method

	RPCService *RPCService
}

func (m *ServiceMethod) call(ctx context.Context, req interface{}) (interface{}, error) {
	args := []reflect.Value{m.RPCService.rcvr, reflect.ValueOf(ctx), reflect.ValueOf(req)}
	res := m.method.Func.Call(args)
	if len(res) != 2 {
		return nil, fmt.Errorf("response length not equal 2 len:%v", res)
	}

	var reply, err interface{}
	reply = res[0].Interface()
	err = res[1].Interface()
	if err != nil {
		return nil, err.(error)
	}
	return reply, nil
}

func register(srv interface{}) (s *RPCService, err error) {
	// Setup service.
	s = &RPCService{
		rcvr:     reflect.ValueOf(srv),
		rcvrType: reflect.TypeOf(srv),
		methods:  make(map[string]*ServiceMethod),
	}
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	if !isExported(s.name) {
		return nil, fmt.Errorf("no service name for type %q", s.rcvrType.String())
	}

	// Setup methods.
	for i := 0; i < s.rcvrType.NumMethod(); i++ {
		method := s.rcvrType.Method(i)
		srvMethod := newServiceMethod(s, &method)
		if srvMethod != nil {
			s.methods[method.Name] = srvMethod
		}
	}

	return
}

func newServiceMethod(s *RPCService, m *reflect.Method) *ServiceMethod {
	// Method must be exported.
	if m.PkgPath != "" {
		log.Printf("method %#v is not exported", m)
		return nil
	}
	mtype := m.Type
	numIn, numOut := mtype.NumIn(), mtype.NumOut()

	if numIn != 3 || numOut != 2 {
		return nil
	}
	// First value must be of context type Last return value must be of error type
	if !isContext(mtype.In(1)) || !isErrType(mtype.Out(mtype.NumOut()-1)) {
		panic(fmt.Errorf("context Type:%v errType:%v", mtype.In(0), mtype.Out(mtype.NumOut()-1)))
	}

	reqType := mtype.In(2)
	// Second argument must be a pointer and must be exported.
	if reqType.Kind() != reflect.Ptr || !isExportedOrBuiltin(reqType) {
		return nil
	}

	return &ServiceMethod{
		ReqType:    mtype.In(2).Elem(),
		RespType:   mtype.Out(0).Elem(),
		method:     m,
		RPCService: s,
	}
}

// isExported returns true of a string is an exported (upper case) name.
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// isExportedOrBuiltin returns true if a type is exported or a builtin.
func isExportedOrBuiltin(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

func isContext(t reflect.Type) bool {
	return t.Implements(typeOfContext)
}

func isErrType(t reflect.Type) bool {
	return t.Implements(typeOfOsError)
}
