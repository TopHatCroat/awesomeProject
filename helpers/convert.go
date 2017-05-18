package helpers

import (
	"bytes"
	"fmt"
	"github.com/paulmach/go.geo"
)

func ToPolygonWKT(ps geo.PointSet) string {
	if len(ps) == 0 {
		return "EMPTY"
	}

	buff := bytes.NewBuffer(nil)
	fmt.Fprintf(buff, "POLYGON((%g %g", ps[0][0], ps[0][1])

	for i := 1; i < len(ps); i++ {
		fmt.Fprintf(buff, ",%g %g", ps[i][0], ps[i][1])
	}

	fmt.Fprintf(buff, ",%g %g", ps[0][0], ps[0][1])

	buff.Write([]byte("))"))
	return buff.String()
}
