package locator

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kellydunn/golang-geo"
)

// Request is an element accepted by the API that contains all search parameters.
type Request struct {
	CountryCode string
	CountryName string
	CityName    string
	Latitude    float64
	Longitude   float64
	Locale      string

	startedAt time.Time
}

// Location is an element returned by all API functions.
type Location struct {
	IATA        string `json:"iata"`
	Name        string `json:"name"`
	CountryName string `json:"country_name"`
	TimeZone    string `json:"-"`
	Coordinates string `json:"coordinates"`

	Approach string        `json:"-"`
	Took     time.Duration `json:"-"`
}

type country struct {
	name         string
	cities       map[string]*city
	primaryCity  *city
	translations map[string]string
}

type city struct {
	iata         string
	lat, long    *float64
	timezone     string
	translations map[string]string
}

const (
	fallbackLocale = "en"

	approachExact   = "exact_match"
	approachPrimary = "primary_city"
	approachClosest = "closest_city"
)

var (
	countries = map[string]*country{}
)

// Lookup tries to locate a city that matches given parameters.
func Lookup(req Request) *Location {
	req.startedAt = time.Now()

	// First look for a country match
	if c, ok := countries[req.CountryCode]; ok {
		// Look for exact city match
		if cit, ok := c.cities[req.CityName]; ok {
			return req.buildResponse(c, cit, approachExact)
		}

		// Fallback to primary city if defined
		if c.primaryCity != nil {
			return req.buildResponse(c, c.primaryCity, approachPrimary)
		}
	}

	// GeoIP returne 0:0 not found place
	if req.Latitude == 0 && req.Longitude == 0 {
		return nil
	}

	// Last resort: try to find the closest city within 100km
	pt := geo.NewPoint(req.Latitude, req.Longitude)
	if c, cit, ok := findCityWithinDistance(pt, 100); ok {
		return req.buildResponse(c, cit, approachClosest)
	}

	return nil
}

func findCityWithinDistance(from *geo.Point, maxDist float64) (c *country, cit *city, ok bool) {
	var closestDistance float64 = -1
	var closestCountry *country
	var closestCity *city

	for _, c := range countries {
		for _, cit := range c.cities {
			if cit.lat != nil && cit.long != nil {
				to := geo.NewPoint(*cit.lat, *cit.long)
				dist := from.GreatCircleDistance(to)
				if dist < maxDist && closestDistance < 0 || dist < closestDistance {
					closestDistance = dist
					closestCountry = c
					closestCity = cit
				}
			}
		}
	}
	if closestCity != nil {
		return closestCountry, closestCity, true
	}

	return nil, nil, false
}

func (r Request) buildResponse(c *country, cit *city, approach string) *Location {
	return &Location{
		IATA:        cit.iata,
		Name:        firstNonEmptyString(cit.translations[r.Locale], cit.translations[fallbackLocale]),
		CountryName: firstNonEmptyString(c.translations[r.Locale], c.translations[fallbackLocale]),
		TimeZone:    cit.timezone,
		Coordinates: cit.coords(),
		Approach:    approach,
		Took:        time.Now().Sub(r.startedAt),
	}
}

func (c *city) coords() string {
	if c.lat != nil && c.long != nil {
		return fmt.Sprintf("%v:%v", strconv.FormatFloat(*c.long, 'f', -1, 64), strconv.FormatFloat(*c.lat, 'f', -1, 64))
	}

	return ""
}

func firstNonEmptyString(vals ...string) string {
	for _, s := range vals {
		if s != "" {
			return s
		}
	}

	return ""
}
