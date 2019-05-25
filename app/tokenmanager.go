package app

import (
	"context"
	"fmt"
	"crypto/rand"
)

var TokenMap map[string]TokeData

type TokeData struct {
	Subtitles chan Subtitles
	Cancel    *context.CancelFunc
}

func init()  {
	TokenMap = make(map[string]TokeData)
}

func TokenGenerator() string {
	b := make([]byte, 10)
	rand.Read(b)

	return fmt.Sprintf("%x", b)
}
