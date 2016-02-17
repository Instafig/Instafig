package utils

import (
	"fmt"
	"time"

	"github.com/gorilla/feeds"
)

func GenerateKey() string {
	u := feeds.NewUUID()
	return fmt.Sprintf("%x%x%x%x%x", u[:4], u[4:6], u[6:8], u[8:10], u[10:])
}

func GetNowSecond() int {
	return int(time.Now().Unix())
}
