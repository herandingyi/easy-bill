package models

import "time"

type Command struct {
	Id         int64
	SenderId   int64
	Command    string
	Args       string
	CreateTime time.Time
}
