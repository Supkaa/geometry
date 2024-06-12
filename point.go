package geometry

import (
	"encoding/hex"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/encoding/wkt"
	"github.com/paulmach/orb/geojson"
)

type Point interface {
	GeoJSONType() string
	Dimensions() int
	Lat() float64
	Lon() float64
	Bound() orb.Bound
	ToWKT() string
	ToGeoJSON() geojson.Point
}

func NewPointFromWKB(wkbPoint string) (Point, error) {
	bytes, err := hex.DecodeString(wkbPoint)

	if err != nil {
		return nil, err
	}

	p, err := wkb.Unmarshal(bytes)

	if err != nil {
		return nil, err
	}

	if !isPoint(p) {
		return nil, ErrFailedGeometryType
	}

	return point{
		Point: p.(orb.Point),
	}, nil
}

func NewPointFromWKT(wktPolygon string) (Point, error) {
	p, err := wkt.Unmarshal(wktPolygon)

	if err != nil {
		return nil, err
	}

	if !isPoint(p) {
		return nil, ErrFailedGeometryType
	}

	return point{
		Point: p.(orb.Point),
	}, nil
}

func NewPointFromOrb(orbPoint orb.Geometry) (Point, error) {
	if !isPoint(orbPoint) {
		return nil, ErrFailedGeometryType
	}

	return point{
		Point: orbPoint.(orb.Point),
	}, nil
}

type point struct {
	orb.Point
}

func (p point) ToGeoJSON() geojson.Point {
	return geojson.Point(p.Point)
}

func (p point) ToWKT() string {
	return wkt.MarshalString(p.Point)
}

func isPoint(geom orb.Geometry) bool {
	return geom.GeoJSONType() == "Point"
}
