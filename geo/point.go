package geo

import (
	"github.com/TomiHiltunen/geohash-golang"
)

// Point is a geohashed latitude longitude
type Point string

// LatLng decodes the geohash in p and returns it's center.
func (p Point) LatLng() (float64, float64) {
	center := geohash.Decode(string(p)).Center()
	return center.Lat(), center.Lng()
}

// NewPoint geohashes a lan, lng pair into a point
func NewPoint(lat, lng float64) Point {
	return Point(geohash.Encode(lat, lng))
}
