package storage

import (
	"encoding/json"
	"time"
)

type User struct {
	Login       string     `json:"login"`
	Pass        string     `json:"pass"`
	LastLogin   time.Time  `json:"last_login"`
	Permissions Permission `json:"permissions"`
	History     []string   `json:"history"`
}

func (u *User) Unmarshal(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u *User) Marshal() ([]byte, error) {
	return json.Marshal(u)
}

type Permission struct {
	RequestPerHour int           `json:"req_per_hour"`
	TTL            time.Duration `json:"ttl"`
}

type Converter struct {
	Adress       string    `json:"adress"`
	RegisterTime time.Time `json:"register_time"`
	Token        string    `json:"token"`
}

func (c *Converter) Unmarshal(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *Converter) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

type HistoryItem struct {
	Time   time.Time `json:"time"`
	Title  string    `json:"title"`
	Link   string    `json:"audio_link"`
	Status string    `json:"status"`
}

func (j *HistoryItem) Unmarshal(data []byte) error {
	return json.Unmarshal(data, j)
}

func (j *HistoryItem) Marshal() ([]byte, error) {
	return json.Marshal(j)
}
