ActiveRecord::Base.logger = nil
result = []
count = 0
count_all = City.count
City.joins(:country).limit(100).load.each do |city|
  count += 1
  puts "#{count} / #{count_all}"
  translations = []

  result.push({
    country_iata: city.country.iata,
    country: city.country.english_name,
    city: city.english_name,
    timezone: city.time_zone,
    latitude: city.lat,
    longitude: city.lon,
    translations: translations
  })
  locales = city.country.translations.map(&:locale)
  locales.each do |locale|
    translations.push({
      locale: locale,
      country: city.country.translations.where(locale: locale).try(:first).try(:name) || city.country.english_name,
      city: city.translations.where(locale: locale).try(:first).try(:name) || city.country.english_name,
    })
  end;
end;

r = result.to_json();
File.write('/Users/nimdraug/Work/go/src/github.com/nimdraugsael/locator/configs/all_cities_sample.json', r)

