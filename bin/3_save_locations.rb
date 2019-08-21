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
  center = html.scan(/;center=[\-\d\.]+%2C[\-\d\.]+/).first

  raise "Failed to get location #{location["id"]}" if center.nil?

  lat, long = center.split(/=|%2C/)[1..2]

  begin
    location.merge!(lat: lat, long: long)
    location.delete("has_public_page")
    File.write(location_file_name, JSON.pretty_generate(location))
  rescue
    puts "Place missing location data (#{location["name"]} - #{location["id"]})"
  end
end
