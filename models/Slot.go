package models

import "time"

type Slot struct {
	ID        int       `json:"id"`
	ShopID    int       `json:"shopId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}
