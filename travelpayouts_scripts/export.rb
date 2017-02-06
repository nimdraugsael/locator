ActiveRecord::Base.logger = nil
city_count = 0
city_count_all = City.count
CITIES = {}
City.all.each do |city|
  city_count += 1
  CITIES[city.id] = city
  puts "Cities -> memory: #{city_count}/#{city_count_all}"
end;

country_count = 0
country_count_all = Country.count
COUNTRIES = {}
PRIMARY_CITY_IDS = []
Country.all.each do |country|
  country_count += 1
  COUNTRIES[country.id] = country
  PRIMARY_CITY_IDS << City.where(country: country).primary.try(:id)
  puts "Countries -> memory: #{country_count}/#{country_count_all}"
end;
PRIMARY_CITY_IDS.compact!;

translations_count = 0
translations_count_all = PlaceTranslation.count
TRANSLATIONS = {}
PlaceTranslation.all.each do |pt|
  translations_count += 1
  TRANSLATIONS[pt.place_id] ||= {}
  TRANSLATIONS[pt.place_id][pt.locale] = pt
  puts "Translations -> memory: #{translations_count}/#{translations_count_all}"
end;

result = []
result_count = 0
CITIES.values.each do |city|
  result_count += 1
  country = COUNTRIES[city.parent_id]

  translations = []
  city_translations = TRANSLATIONS[city.id]
  country_translations = TRANSLATIONS[country.id]

  if city_translations.nil? || country_translations.nil?
    puts "Translations not found, skipping"
    next
  end

  city_translations.each do |locale, t|
    translations.push({
      locale: locale,
      country: country_translations[locale].try(:name) || country.english_name,
      city: city_translations[locale].try(:name) || city.english_name,
    })
  end

  result.push({
    city: city.english_name,
    city_iata: city.iata,
    country: country.english_name,
    country_iata: country.iata,
    latitude: city.lat,
    longitude: city.lon,
    timezone: city.time_zone,
    is_primary: PRIMARY_CITY_IDS.include?(city.id),
    translations: translations
  })

  puts "Processing: #{result_count}/#{city_count_all}"
end;

Oj.default_options = {:mode => :compat }
json = Oj.dump(result, symbol_keys: false);

File.write("/Users/nimdraug/Work/go/src/github.com/nimdraugsael/locator/configs/cities.json", json)

