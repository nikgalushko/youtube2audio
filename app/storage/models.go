package storage

import (
	"encoding/json"
	"time"
)

type User struct {
	ID          int        `json:"_id"`
	Login       string     `json:"login"`
	Pass        string     `json:"pass"`
	LastLogin   time.Time  `json:"last_login"`
	Permissions Permission `json:"permissions"`
}

func (u *User) Unmarshal(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u User) Marshal() ([]byte, error) {
	return json.Marshal(u)
}

type Permission struct {
	RequestPerHour int           `json:"req_per_hour"`
	TTL            time.Duration `json:"ttl"`
}
