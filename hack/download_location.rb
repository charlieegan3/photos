#!/usr/bin/env ruby

require "json"
require "open-uri"

location = { id: ARGV[0] }

location_file_name = "locations/#{location[:id]}.json"

html = open("https://www.instagram.com/explore/locations/#{location[:id]}").read
page_data = html.scan(/\{[^\n]+\}/).map { |r| JSON.parse(r) rescue nil }.compact.first

location_data = page_data.dig(*%w(entry_data LocationsPage)).first["location"] ||
  page_data["entry_data"]["LocationsPage"].first["graphql"]["location"]

location.merge!(lat: location_data["lat"], long: location_data["lng"], name: location_data["name"], slug: location_data["slug"])
location.delete("has_public_page")
File.write(location_file_name, JSON.pretty_generate(location) + "\n")

location.delete(:lat)
location.delete(:long)
location["has_public_page"] = true

puts JSON.pretty_generate({ location: location })
