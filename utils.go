package main

import "time"

func ParseTime(t int64, format string) string {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return time.Unix(t/1000, 0).Format(format)
}
