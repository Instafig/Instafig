package utils

import (
	"fmt"
	"time"

	"github.com/gorilla/feeds"
)

const millisecondUit = int64(time.Millisecond/time.Nanosecond)

func GenerateKey() string {
	u := feeds.NewUUID()
	return fmt.Sprintf("%x%x%x%x%x", u[:4], u[4:6], u[6:8], u[8:10], u[10:])
}

func GetNowSecond() int {
	return int(time.Now().Unix())
}

func GetNowMillisecond() int64 {
	return time.Now().UnixNano() / millisecondUit
}

func GetNowStringYMD() string {
	return time.Now().Format("2006-01-02")
}