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

  puts "https://hotsta.org/geo/#{location["id"]}"
  html = open("http://hotsta.org/geo/#{location["id"]}").read
  lat, long = html.scan(/(google.com\/maps)\S+q=(\S+)"/).first.last.split(",").map(&:to_f)

  begin
    location.merge!(lat: lat.to_f, long: long.to_f)
    location.delete("has_public_page")
    File.write(location_file_name, JSON.pretty_generate(location))
  rescue
    puts "Place missing location data (#{location["name"]} - #{location["id"]})"
  end
end
