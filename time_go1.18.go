//go:build go1.18
// +build go1.18

package template

import "time"

func unixMilli(t time.Time) int64 {
	return t.UnixMilli()
}

func unixMicro(t time.Time) int64 {
	return t.UnixMicro()
}
