package models

import (
	"github.com/jinzhu/gorm"
)

type Env struct {
	DB         *gorm.DB
	PolygonAdd chan Polygon
}

var (
	SERVER_IP = "46.101.106.208"
)

func NewEnviroment(db *gorm.DB) *Env {
	env := &Env{DB: db, PolygonAdd: make(chan Polygon)}

	return env
}
