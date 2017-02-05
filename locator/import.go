package locator

import (
	"encoding/json"
	"os"
)

type rawCity struct {
	City         string   `json:"city"`
	CityIATA     string   `json:"city_iata"`
	Country      string   `json:"country"`
	CountryIATA  string   `json:"country_iata"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	Timezone     string   `json:"timezone"`
	IsPrimary    bool     `json:"is_primary"`
	Translations []struct {
		Locale  string `json:"locale"`
		Country string `json:"country"`
		City    string `json:"city"`
	} `json:"translations"`
}

// ImportCitiesFile loads cities file into index.
func ImportCitiesFile(fileName string) error {
	fd, err := os.Open(fileName)
	if err != nil {
		return err
	}

	var rawCities []rawCity
	err = json.NewDecoder(fd).Decode(&rawCities)
	if err != nil {
		return err
	}

	for _, rc := range rawCities {
		importCity(rc)
	}

	return nil
}

func importCity(rc rawCity) {
	if _, ok := countries[rc.CountryIATA]; !ok {
		countries[rc.CountryIATA] = &country{
			name:         rc.Country,
			cities:       map[string]*city{},
			translations: map[string]string{},
		}
	}

	c := &city{
		iata:         rc.CityIATA,
		lat:          rc.Latitude,
		long:         rc.Longitude,
		timezone:     rc.Timezone,
		translations: map[string]string{},
	}
	for _, tr := range rc.Translations {
		c.translations[tr.Locale] = tr.City
		if _, ok := countries[rc.CountryIATA].translations[tr.Locale]; !ok {
			countries[rc.CountryIATA].translations[tr.Locale] = tr.Country
		}
	}

	countries[rc.CountryIATA].cities[rc.City] = c
	if rc.IsPrimary {
		countries[rc.CountryIATA].primaryCity = c
	}
}
