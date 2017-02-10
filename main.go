package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	newrelic "github.com/newrelic/go-agent"
	"github.com/nimdraugsael/locator/locator"
	logging "github.com/op/go-logging"
	geoip2 "github.com/oschwald/geoip2-golang"
	"github.com/tomasen/realip"
)

var (
	geoip2db *geoip2.Reader
)

var log = logging.MustGetLogger("example")
var format = logging.MustStringFormatter(
	`%{color}%{time:2006-01-02 15:04:05} > %{level:.4s} %{color:reset} %{message}`,
)

func main() {
	var configDir, port, logFile, newrelicKey string

	stdoutLogBackend := logging.NewLogBackend(os.Stdout, "", 0)
	stdoutLogFormatter := logging.NewBackendFormatter(stdoutLogBackend, format)
	logging.SetBackend(stdoutLogFormatter)

	flag.StringVar(&configDir, "configs", "./configs", "Path to configs directory")
	flag.StringVar(&port, "port", "8100", "Web server port")
	flag.StringVar(&logFile, "log_file", "", "Path to log file")
	flag.StringVar(&newrelicKey, "newrelic_key", "", "NewRelic key for monitoring")
	flag.Parse()

	if logFile != "" {
		fd, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			log.Panicf("Error creating log file %v", logFile)
		}
		fileLogBackend := logging.NewLogBackend(fd, "", 0)
		fileLogFormatter := logging.NewBackendFormatter(fileLogBackend, format)
		logging.SetBackend(stdoutLogFormatter, fileLogFormatter)
	}

	var err error
	geoip2db, err = geoip2.Open(configDir + "/GeoIP2-City.mmdb")
	if err != nil {
		log.Panicf("Failed to open GeoIP database: %v", err)
	}

	err = locator.ImportCitiesFile(configDir + "/cities.json")
	if err != nil {
		log.Panicf("Failed to open cities database file: %v", err)
	}

	if newrelicKey != "" {
		nrc := newrelic.NewConfig("TpLocator", newrelicKey)
		nrapp, _ := newrelic.NewApplication(nrc)
		log.Info("NewRelic monitoring ON")
		http.HandleFunc(newrelic.WrapHandleFunc(nrapp, "/whereami", handler))
		http.HandleFunc(newrelic.WrapHandleFunc(nrapp, "/whereami2", handler))
	} else {
		log.Info("NewRelic monitoring OFF")
		http.HandleFunc("/whereami", handler)
		http.HandleFunc("/whereami2", handler)
	}

	log.Info("Starting server at port", port)
	http.ListenAndServe(":"+port, nil)
}

func handler(rw http.ResponseWriter, req *http.Request) {
	var err error
	var loc *locator.Location
	req.Header.Set("Access-Control-Allow-Origin", "*")

	ip := net.ParseIP(requestIP(req))
	geoCity, err := geoip2db.City(ip)
	if err != nil {
		log.Errorf("Failed to resolve IP address %s to geo point: %v\n", ip, err)
		loc = locator.GetDefaultReponse(requestLocale(req))
	}

	lreq := locator.Request{
		CountryCode: geoCity.Country.IsoCode,
		CountryName: geoCity.Country.Names["en"],
		CityName:    geoCity.City.Names["en"],
		Latitude:    geoCity.Location.Latitude,
		Longitude:   geoCity.Location.Longitude,
		Locale:      requestLocale(req),
	}

	loc = locator.Lookup(lreq)
	if loc == nil {
		log.Infof("Location lookup failed: %+v\n", lreq)
		loc = locator.GetDefaultReponse(requestLocale(req))
	}

	log.Infof("Processed %f ms", loc.Took.Seconds()*1000)

	result, _ := json.Marshal(loc)

	if cb := req.FormValue("callback"); cb != "" {
		req.Header.Set("Content-Type", "Type:application/x-javascript; charset=utf-8")
		fmt.Fprintf(rw, "%v(%v);", cb, string(result))
	} else {
		req.Header.Set("Content-Type", "Type:application/json; charset=utf-8")
		fmt.Fprint(rw, string(result))
	}
}

func requestIP(req *http.Request) string {
	// Use parameter value if provided
	if req.FormValue("ip") != "" {
		return req.FormValue("ip")
	}

	// Use request IP address by default
	ip := realip.RealIP(req)

	return ip
}

func requestLocale(req *http.Request) string {
	if req.FormValue("locale") != "" {
		return req.FormValue("locale")
	}

	return "ru"
}
