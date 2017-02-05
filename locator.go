package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/nimdraugsael/locator/locator"
	geoip2 "github.com/oschwald/geoip2-golang"
)

var GeoIP *geoip2.Reader

type (
	JsonResponse struct {
		Iata        string `json:"iata"`
		Name        string `json:"name"`
		CountryName string `json:"country_name"`
	}
)

func handler(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		ip = "81.2.69.142"
	}
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "ru"
	}

	record, err := GeoIP.City(net.ParseIP(ip))
	if err != nil {
		log.Fatal(err)
	}

	var result locator.TranslationMapItem

	start_time := time.Now()

	result, err = locator.FindByCityAndCountry(record.Country.Names["en"], record.City.Names["en"], record.Location.TimeZone)
	if nfe, ok := err.(*locator.NotFoundError); ok {
		fmt.Printf("Error: %v\n", nfe.Error())
		result, err = locator.FindByCountry(record.Country.IsoCode)
		if nfe, ok := err.(*locator.NotFoundError); ok {
			fmt.Printf("Error: %v\n", nfe.Error())

			result, err = locator.FindByCoords(record.Location.Latitude, record.Location.Longitude)
			if nfe, ok := err.(*locator.NotFoundError); ok {
				fmt.Printf("Error: %v\n", nfe.Error())
			}
		}
	}

	duration := time.Since(start_time).Seconds() * 1000
	fmt.Printf("Request %v %f ms\n", r.URL, duration)

	json_response := &JsonResponse{Iata: record.Country.IsoCode,
		Name:        result.Translations[locale].City,
		CountryName: result.Translations[locale].Country}

	json_result, _ := json.Marshal(json_response)

	fmt.Fprintf(w, string(json_result))
}

func main() {
	start_time := time.Now()
	db, err := geoip2.Open("/Users/nimdraug/Work/GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	GeoIP = db
	duration := time.Since(start_time)
	fmt.Printf("Init GeoIP: %v\n", duration)
	defer db.Close()

	start_time = time.Now()
	locator.InitAllCities("./configs/all_cities.json")
	duration = time.Since(start_time)
	fmt.Printf("Init by_city_and_country: %v\n", duration)

	start_time = time.Now()
	locator.InitPrimaryCities("./configs/primary_cities.json")
	duration = time.Since(start_time)
	fmt.Printf("Init by_country: %v\n", duration)

	http.HandleFunc("/whereami", handler)
	http.ListenAndServe(":8100", nil)
}
