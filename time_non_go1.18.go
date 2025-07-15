//go:build !go1.18
// +build !go1.18

package template

import "time"

func unixMilli(t time.Time) int64 {
	return t.UnixNano() / 1e6
}

func unixMicro(t time.Time) int64 {
	return t.UnixNano() / 1e3
}
