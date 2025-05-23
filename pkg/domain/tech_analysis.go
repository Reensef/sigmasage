package domain

import (
	"time"
)

type SMAInfo struct {
	MarketData MarketData
	Length     int
}

type SMA struct {
	Info  SMAInfo
	Value float64
	Time  time.Time
}

type SMASrc struct {
	Value float64
	Time  time.Time
}
