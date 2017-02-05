package locator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sort"
)

var (
	ByCityAndCountryTranslationsMap = make(map[string]TranslationMapItem)
	ByCountryTranslationsMap        = make(map[string]TranslationMapItem)
	ByCoordsList                    = make([]TranslationMapItem, 0)
)

type (
	City struct {
		CountryIATA  string `json:"country_iata"`
		Country      string
		City         string
		Timezone     string
		Latitude     float64
		Longitude   float64
		Translations []Translation
	}
	Translation struct {
		Locale  string
		Country string
		City    string
	}

	TranslationMapItem struct {
		Country      string
		City         string
		Timezone     string
		Latitude     float64
		Longitude   float64
		Translations map[string]Translation
	}

	ProcessedCoordsChunk struct {
		delta float64
		item  TranslationMapItem
	}
	ProcessedCoordsChunks []ProcessedCoordsChunk
)

type NotFoundError struct {
	msg string
}

func (error *NotFoundError) Error() string {
	return error.msg
}

func (slice ProcessedCoordsChunks) Len() int {
	return len(slice)
}

func (slice ProcessedCoordsChunks) Less(i, j int) bool {
	return slice[i].delta < slice[j].delta
}

func (slice ProcessedCoordsChunks) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func FindByCityAndCountry(city, country, timezone string) (TranslationMapItem, error) {
	var r TranslationMapItem
	key := city + "___" + country + "___" + timezone
	if result, ok := ByCityAndCountryTranslationsMap[key]; ok {
		return result, nil
	} else {
		return r, &NotFoundError{"FindByCityAndCountry"}
	}
}

func FindByCountry(country string) (TranslationMapItem, error) {
	var r TranslationMapItem
	if result, ok := ByCountryTranslationsMap[country]; ok {
		return result, nil
	} else {
		return r, &NotFoundError{"FindByCountry"}
	}
}

func FindByCoords(latitude, longitude float64) (TranslationMapItem, error) {
	processed_coords_chunks := make(ProcessedCoordsChunks, 0)
	for _, city := range ByCoordsList {
		delta := math.Abs(city.Latitude-latitude) + math.Abs(city.Longitude-longitude)

		processed_chunk := ProcessedCoordsChunk{delta: delta, item: city}
		processed_coords_chunks = append(processed_coords_chunks, processed_chunk)
	}
	sort.Sort(processed_coords_chunks)
	return processed_coords_chunks[0].item, nil
}

func InitAllCities(config_path string) {
	file, e := ioutil.ReadFile(config_path)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	cities := make([]City, 0)
	json.Unmarshal(file, &cities)

	for _, city := range cities {
		key := city.Country + "___" + city.City + "___" + city.Timezone
		translation_map_item := TranslationMapItem{
			City:         city.City,
			Country:      city.Country,
			Timezone:     city.Timezone,
			Latitude:     city.Latitude,
			Longitude:   city.Longitude,
			Translations: make(map[string]Translation)}

		for _, translation := range city.Translations {
			translation_map_item.Translations[translation.Locale] = translation
		}

		ByCityAndCountryTranslationsMap[key] = translation_map_item
		ByCoordsList = append(ByCoordsList, translation_map_item)
	}
}

func InitPrimaryCities(config_path string) {
	file, e := ioutil.ReadFile(config_path)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	cities := make([]City, 0)
	json.Unmarshal(file, &cities)

	for _, city := range cities {
		key := city.CountryIATA
		translation_map_item := TranslationMapItem{
			City:         city.City,
			Country:      city.Country,
			Timezone:     city.Timezone,
			Translations: make(map[string]Translation)}

		for _, translation := range city.Translations {
			translation_map_item.Translations[translation.Locale] = translation
		}

		ByCountryTranslationsMap[key] = translation_map_item
	}
}
