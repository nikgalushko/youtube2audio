package storage

import (
	"encoding/json"
	"time"
)

type User struct {
	ID          int
	Login       string
	Pass        string
	LastLogin   time.Time
	Permissions Permission
}

func (u *User) Unmarshal(data []byte) error {
	return json.Unmarshal(data, u)
}

type Permission struct {
	RequestPerHour int
	TTL            time.Duration
}
