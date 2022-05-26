package gin

import (
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"net/http"
	"strings"
)

type defaultCodec struct {
}

func (c defaultCodec) Encode(r *http.Request, w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(200)

	if err, _ := v.(error); err != nil {
		var code, message string
		parts := strings.Split(grpc.ErrorDesc(err), ":")
		if len(parts) > 1 {
			code = strings.TrimSpace(parts[0])
			message = strings.TrimSpace(strings.Join(parts[1:], ": "))
		} else {
			message = err.Error()
		}
		return json.NewEncoder(w).Encode(map[string]string{
			"code":    code,
			"message": message,
		})
	}

	if vv, ok := v.(proto.Message); ok {
		m := &jsonpb.Marshaler{
			EmitDefaults: true,
		}
		return m.Marshal(w, vv)
	}

	return json.NewEncoder(w).Encode(v)
}

func (c defaultCodec) Decode(r *http.Request, v interface{}) (err error) {
	return json.NewDecoder(r.Body).Decode(v)
}
