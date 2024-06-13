package geometry

import (
	"encoding/hex"
	"errors"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/encoding/wkt"
	"github.com/paulmach/orb/geo"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"reflect"
	"slices"
)

type Polygon interface {
	GeoJSONType() string
	Dimensions() int
	Bound() orb.Bound
	ToWKT() string
	ToGeoJSON() geojson.Geometry
	Centroid() Point
	Area() float64
	Divide(n float64) []Polygon
}

var (
	ErrFailedGeometryType = errors.New("failed geometry type")
)

func NewPolygonFromWKB(wkbPolygon string) (Polygon, error) {
	bytes, err := hex.DecodeString(wkbPolygon)

	if err != nil {
		return nil, err
	}

	p, err := wkb.Unmarshal(bytes)

	if err != nil {
		return nil, err
	}

	if !isPolygon(p) {
		if p.GeoJSONType() != "GeometryCollection" {
			return nil, ErrFailedGeometryType
		}

		p = p.Bound().ToPolygon()
	}

	orbCentroid, area := planar.CentroidArea(p)
	centroid, err := NewPointFromOrb(orbCentroid)

	if err != nil {
		return nil, err
	}

	return polygon{
		centroid: centroid,
		area:     area,
		Geometry: p,
	}, nil
}

func NewPolygonFromWKT(wktPolygon string) (Polygon, error) {
	p, err := wkt.Unmarshal(wktPolygon)

	if err != nil {
		return nil, err
	}

	if !isPolygon(p) {
		return nil, ErrFailedGeometryType
	}

	orbCentroid, area := planar.CentroidArea(p)
	centroid, err := NewPointFromOrb(orbCentroid)

	if err != nil {
		return nil, err
	}

	return polygon{
		centroid: centroid,
		area:     area,
		Geometry: p,
	}, nil
}

func NewPolygonFromOrb(orbPolygon orb.Geometry) (Polygon, error) {
	if !isPolygon(orbPolygon) {
		return nil, ErrFailedGeometryType
	}

	orbCentroid, area := planar.CentroidArea(orbPolygon)
	centroid, err := NewPointFromOrb(orbCentroid)

	if err != nil {
		return nil, err
	}

	return polygon{
		centroid: centroid,
		area:     area,
		Geometry: orbPolygon,
	}, nil
}

func NewPolygonFromGeoJSON() {

}

func NewPolygonFromPlanarPoints(points []point) (Polygon, error) {
	ring := orb.Ring{}

	for _, p := range points {
		ring = append(ring, p.Point)
	}

	if !reflect.DeepEqual(points[0], points[len(points)-1]) {
		ring = append(ring, points[0].Point)
	}

	poly := orb.Polygon{ring}

	return NewPolygonFromOrb(poly)
}

type polygon struct {
	centroid Point
	area     float64
	orb.Geometry
}

func (p polygon) ToGeoJSON() geojson.Geometry {
	return geojson.Geometry{
		Type:        p.GeoJSONType(),
		Coordinates: p.Geometry,
	}
}

func (p polygon) ToWKT() string {
	return wkt.MarshalString(p.Geometry)
}

func (p polygon) Centroid() Point {
	return p.centroid
}

func (p polygon) Area() float64 {
	return p.area
}

// Divide polygon bound into parts less than n square meters
func (p polygon) Divide(n float64) []Polygon {
	var polygons []Polygon
	bbox := p.Bound()
	bboxArea := geo.Area(bbox) / 1_000_000
	if bboxArea <= n {
		poly, _ := NewPolygonFromOrb(bbox)

		return append(polygons, poly)
	}

	for _, half := range divide(bbox) {
		polygons = append(polygons, half.Divide(n)...)
	}

	return polygons
}

func divide(bbox orb.Bound) [2]Polygon {
	var parts [2]Polygon

	width := geo.Distance(bbox.Min, orb.Point{bbox.Max[0], bbox.Min[1]})
	height := geo.Distance(bbox.Min, orb.Point{bbox.Min[0], bbox.Max[1]})

	if width > height {
		centerX := (bbox.Min[0] + bbox.Max[0]) / 2
		parts[0], _ = NewPolygonFromOrb(orb.Bound{
			Min: bbox.Min,
			Max: orb.Point{centerX, bbox.Max[1]},
		})
		parts[1], _ = NewPolygonFromOrb(orb.Bound{
			Min: orb.Point{centerX, bbox.Min[1]},
			Max: bbox.Max,
		})

		return parts
	}

	centerY := (bbox.Min[1] + bbox.Max[1]) / 2
	parts[0], _ = NewPolygonFromOrb(orb.Bound{
		Min: bbox.Min,
		Max: orb.Point{bbox.Max[0], centerY},
	})
	parts[1], _ = NewPolygonFromOrb(orb.Bound{
		Min: orb.Point{bbox.Min[0], centerY},
		Max: bbox.Max,
	})

	return parts
}

func isPolygon(geom orb.Geometry) bool {
	return slices.Contains([]string{"Polygon", "MultiPolygon"}, geom.GeoJSONType())
}
