package gin

import (
	"net/http"
)

type Codec interface {
	Encode(r *http.Request, w http.ResponseWriter, v interface{}) error
	Decode(r *http.Request, v interface{}) error
}
