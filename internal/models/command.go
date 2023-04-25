package models

import "time"

type Command struct {
	Id         int64
	SenderId   int64
	Command    string
	Args       string
	Status     int // 1有效 2无效
	CreateTime time.Time
}
