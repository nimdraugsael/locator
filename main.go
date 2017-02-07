package main

import (
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"os"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/nimdraugsael/locator/locator"
	logging "github.com/op/go-logging"
	geoip2 "github.com/oschwald/geoip2-golang"
	"github.com/tomasen/realip"
)

var (
	geoip2db *geoip2.Reader
	c        *statsd.Client
)

var log = logging.MustGetLogger("example")
var format = logging.MustStringFormatter(
	`%{color}%{time:2006-01-02 15:04:05} > %{level:.4s} %{color:reset} %{message}`,
)

func main() {
	var configDir string
	var port string
	var logFile string
	stdoutLogBackend := logging.NewLogBackend(os.Stdout, "", 0)
	stdoutLogFormatter := logging.NewBackendFormatter(stdoutLogBackend, format)
	logging.SetBackend(stdoutLogFormatter)

	flag.StringVar(&configDir, "configs", "./configs", "Path to configs directory")
	flag.StringVar(&port, "port", "8100", "Web server port")
	flag.StringVar(&logFile, "log_file", "", "Path to log file")
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

	c, err = statsd.New("0.0.0.0:8125")
	c.Namespace = "tp_locator"
	c.Tags = append(c.Tags, "locator")

	log.Info("Starting server at port", port)
	http.HandleFunc("/whereami", handler)
	http.ListenAndServe(":"+port, nil)
}

func handler(rw http.ResponseWriter, req *http.Request) {
	var err error

	ip := net.ParseIP(requestIP(req))
	geoCity, err := geoip2db.City(ip)
	if err != nil {
		log.Errorf("Failed to resolve IP address %s to geo point: %v\n", ip, err)
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
		log.Infof("Location lookup failed: %+v\n", lreq)
		http.Error(rw, "Failed to figure out your location ¯\\_(ツ)_/¯", http.StatusTeapot)
		return
	}

	err = c.Gauge(".request.duration", loc.Took.Seconds()*1000, nil, 1)
	err = c.Incr(".request.count", nil, 1)
	if err != nil {
		log.Info("Sending to Datadog failed")
	}
	log.Infof("Processed %f ms", loc.Took.Seconds()*1000)

	json.NewEncoder(rw).Encode(loc)
}

func requestIP(req *http.Request) string {
	// Use parameter value if provided
	if req.FormValue("ip") != "" {
		return req.FormValue("ip")
	}

	// Use request IP address by default
	ip := realip.RealIP(req)
	log.Info("Realip ", ip)
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
