package utils

import (
	"time"

	"github.com/gorilla/feeds"
)

func GenerateKey() string {
	return feeds.NewUUID().String()
}

func GetNowSecond() int {
	return int(time.Now().Unix())
}
