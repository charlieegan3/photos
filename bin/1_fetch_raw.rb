#!/usr/bin/env ruby

require "json"
require "open-uri"

url = "https://www.instagram.com/charlieegan3/?__a=1"

count = 0

data = JSON.parse(open(url).read)
data["graphql"]["user"]["edge_owner_to_timeline_media"]["edges"].each do |image|
  image = image["node"]
  filename = Time.at(image["taken_at_timestamp"]).strftime("%Y-%m-%d") + "-" + image["id"] + ".json"
  image.merge!(scraper_version: "v2")

  unless File.exists?("looted_json/" + filename)
    count += 1

    File.write("looted_json/" + filename, JSON.pretty_generate(image, indent: "    "))
  end
end

if count > 0
  puts "#{count} new images"
  `date > updated_at`
end

if count >= 12
  raise "Potentially missing images"
end
