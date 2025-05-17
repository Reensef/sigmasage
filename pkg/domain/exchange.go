package domain

import "time"

type Deal struct {
	Direction  DealDirection
	Time       time.Time
	Price      float64
	Count      int
	LotPrice   float64
	Commission float64
}

type DealDirection int

const (
	DealDirection_BUY DealDirection = iota
	DealDirection_SELL
)
