package control

import (
	m "github.com/TopHatCroat/awesomeProject/models"
	"github.com/paulmach/go.geo"
	"log"
)

var (
	dangers []m.Polygon
)

func InitDanger(e *m.Env) {
	polys := []m.Polygon{}

	if err := e.DB.Find(&polys).Error; err != nil {
		panic(err)
	}

	//for _, pol := range polys {
	//	if err := e.DB.Model(&pol).Related(&pol.Points, "Points").Error; err != nil {
	//		panic(err)
	//	}
	//}

	dangers = polys
	//
	//for _, poly := range dangers {
	//	fmt.Printf("Added polygon %d %s \n", poly.ID, poly.Typ)
	//}
}

func DangerProcessing(e *m.Env) {
	for {
		select {
		case poly := <-e.PolygonAdd:
			dangers = append(dangers, poly)
			//fmt.Printf("Added polygon %d %s \n", poly.ID, poly.Typ)
			return
		}
	}
}

func CheckPoint(e *m.Env) {
	path := geo.NewPath()
	path.Push(geo.NewPoint(0, 0))
	path.Push(geo.NewPoint(1, 1))

	line := geo.NewLine(geo.NewPoint(0, 1), geo.NewPoint(1, 0))

	// intersects does a simpler check for yes/no
	if path.Intersects(line) {
		// intersection will return the actual points and places on intersection
		points, segments := path.Intersection(line)

		for i, _ := range points {
			log.Printf("Intersection %d at %v with path segment %d", i, points[i], segments[i][0])
		}
	}

}
