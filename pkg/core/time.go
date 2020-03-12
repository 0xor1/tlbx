package core

import "time"

func Now() time.Time {
	return time.Now().UTC()
}

func NowMilli() time.Time {
	return time.Now().UTC().Round(time.Millisecond)
}

func NowUnixNano() int64 {
	return Now().UnixNano()
}

func NowUnixMilli() int64 {
	return NowUnixNano() / 1000000
}
