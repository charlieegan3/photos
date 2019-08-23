#!/usr/bin/env ruby

require "json"
require "open-uri"

locations_with_missing_data = %w()

Dir.glob("completed_json/*").map do |file|
  JSON.parse(File.read(file))["location"]
end.uniq.compact.each do |location|
  location_file_name = "locations/#{location["id"]}.json"

  next if File.exists?(location_file_name)
  next if locations_with_missing_data.include?(location["id"])

  html = open("https://facebook.com/pages/locations/#{location["id"]}").read
  center = html.scan(/(;|&)(center|markers)(=)([\-\d\.]+)(%2C|\u00252C)([\-\d\.]+)/).first

  raise "Failed to get location #{location["id"]}" if center.nil?

  lat, long = center[3], center[5]

  begin
    location.merge!(lat: lat.to_f, long: long.to_f)
    location.delete("has_public_page")
    File.write(location_file_name, JSON.pretty_generate(location))
  rescue
    puts "Place missing location data (#{location["name"]} - #{location["id"]})"
  end
end
