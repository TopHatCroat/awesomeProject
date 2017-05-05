package models

type Session struct {
	Id         int
	Token      string `json:"token"`
	CreatedAt  int64  `json:"-"`
	LastUsedAt int64  `json:"-"`
}
