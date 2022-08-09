package util

import (
	"time"
)

const timeFormat = "2006-01-02 15:04:05"

func Int64ToTimeString(i int64) string {
	createTime := i
	t := time.Unix(createTime, 0)
	return t.Format(timeFormat)
}
