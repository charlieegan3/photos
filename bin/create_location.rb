#!/usr/bin/env ruby

require "json"
require "digest"

puts "Enter name:"
name = gets.strip

slug = name.downcase.gsub(/\W+/, "-")

puts "Enter slug: (default: #{slug})"
if (input = gets.chomp).length > 0
  slug = input
end

puts "Enter lat:"
lat = gets.chomp

puts "Enter lon:"
lon = gets.chomp

data = {
  id: Digest::SHA2.hexdigest(name+lat+lon),
  name: name,
  slug: slug,
  lat: lat.to_f,
  long: lon.to_f,
}

puts json = JSON.pretty_generate(data)

filename = "locations/#{data[:id]}.json"

puts "Write #{filename}?"
if gets.chomp == ""
  File.write(filename, json)
end
