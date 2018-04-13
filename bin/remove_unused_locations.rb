#!/usr/bin/env ruby

require "json"

locations = Dir.glob("completed_json/*").map do |file|
  JSON.parse(File.read(file))["location"]["id"].to_s rescue nil
end.uniq.compact

Dir.glob("locations/*").each do |file|
  id = file.gsub(".json", "").gsub("locations/", "")
  File.delete(file) unless locations.include? id
end
