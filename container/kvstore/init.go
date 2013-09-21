package kvstore

import (
	"encoding/gob"
)

func init() {
	gob.Register(logEntry{})
}
