package models

import "time"

type Appointment struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	SlotID    int       `json:"slotId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}
