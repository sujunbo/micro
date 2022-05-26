package json

import "encoding/json"

func MustString(v interface{}) string {
	s, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(s)
}
