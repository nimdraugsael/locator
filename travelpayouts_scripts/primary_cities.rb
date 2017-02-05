ActiveRecord::Base.logger = nil
result = []
count = 0
count_all = Country.count
Country.all.load.each do |country|
  count += 1
  puts "#{count} / #{count_all}"

  city = City.where(country: country).primary
  if city
    translations = []
    result.push({
      country: country.english_name,
      country_iata: country.iata,
      city: city.english_name,
      timezone: city.time_zone,
      translations: translations
    })
    locales = country.translations.map(&:locale)
    locales.each do |locale|
      translations.push({
        locale: locale,
        country: country.translations.where(locale: locale).try(:first).try(:name) || country.english_name,
        city: city.translations.where(locale: locale).try(:first).try(:name) || city.english_name
      })
    end;
  end;
end;

r = result.to_json()
File.write('/Users/nimdraug/Work/go/src/github.com/nimdraugsael/locator/configs/primary_cities.json', r)
