package models

import "github.com/jinzhu/gorm"

type Env struct {
	DB         *gorm.DB
	PolygonAdd chan Polygon
}

func NewEnviroment(db *gorm.DB) *Env {
	env := &Env{DB: db, PolygonAdd: make(chan Polygon)}

	return env
}
