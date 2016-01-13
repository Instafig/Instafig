package utils

import (
	"github.com/gorilla/feeds"
)

func GenerateKey() string {
	return feeds.NewUUID().String()
}
