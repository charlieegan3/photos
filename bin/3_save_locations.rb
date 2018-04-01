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

  html = open("https://www.instagram.com/explore/locations/#{location["id"]}").read
  page_data = html.scan(/\{[^\n]+\}/).map { |r| JSON.parse(r) rescue nil }.compact.first

  location_data = page_data.dig(*%w(entry_data LocationsPage)).first["location"] ||
    page_data["entry_data"]["LocationsPage"].first["graphql"]["location"]

  begin
    location.merge!(lat: location_data["lat"], long: location_data["lng"])
    location.delete("has_public_page")
    File.write(location_file_name, JSON.pretty_generate(location))
  rescue
    puts "Place missing location data (#{location["name"]} - #{location["id"]})"
  end
end
