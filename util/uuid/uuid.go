package uuid

import (
	"github.com/google/uuid"
	"strings"
)

func New() string {
	return strings.Join(strings.Split(uuid.New().String(), "-"), "")
}
