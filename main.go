package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/nimdraugsael/locator/locator"
	geoip2 "github.com/oschwald/geoip2-golang"
	"github.com/tomasen/realip"
)

var (
	geoip2db *geoip2.Reader
)

func main() {
	var configDir string
	var port string
	flag.StringVar(&configDir, "configs", "./configs", "Path to configs directory")
	flag.StringVar(&port, "port", "8100", "Web server port")
	flag.Parse()

	var err error
	geoip2db, err = geoip2.Open(configDir + "/GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal("Failed to open GeoIP database: %v", err)
	}

	err = locator.ImportCitiesFile(configDir + "/cities.json")
	if err != nil {
		log.Fatal("Failed to open cities database file: %v", err)
	}

	log.Println("Starting server at port", port)
	http.HandleFunc("/whereami", handler)
	http.ListenAndServe(":"+port, nil)
}

func handler(rw http.ResponseWriter, req *http.Request) {
	ip := net.ParseIP(requestIP(req))
	geoCity, err := geoip2db.City(ip)
	if err != nil {
		log.Printf("Failed to resolve IP address %s to geo point: %v\n", ip, err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	lreq := locator.Request{
		CountryCode: geoCity.Country.IsoCode,
		CountryName: geoCity.Country.Names["en"],
		CityName:    geoCity.City.Names["en"],
		Latitude:    geoCity.Location.Latitude,
		Longitude:   geoCity.Location.Longitude,
		Locale:      requestLocale(req),
	}

	loc := locator.Lookup(lreq)
	if loc == nil {
		log.Printf("Location lookup failed: %+v\n", lreq)
		http.Error(rw, "Failed to figured out your location ¯\\_(ツ)_/¯", http.StatusTeapot)
		return
	}

	json.NewEncoder(rw).Encode(loc)
}

func requestIP(req *http.Request) string {
	// Use parameter value if provided
	if req.FormValue("ip") != "" {
		return req.FormValue("ip")
	}

	// Use request IP address by default
	ip := realip.RealIP(req)
	if ip == "[::1]" || ip == "127.0.0.1" {
		// Substitute localhost value with a fake address
		return "81.2.69.142"
	}

	return ip
}

func requestLocale(req *http.Request) string {
	if req.FormValue("locale") != "" {
		return req.FormValue("locale")
	}

	return "ru"
}
