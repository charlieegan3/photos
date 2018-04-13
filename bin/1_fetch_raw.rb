#!/usr/bin/env ruby

require "json"
require "open-uri"

url = "https://www.instagram.com/charlieegan3"

count = 0

raw = open(url).read
script = raw.scan(/<script type="text\/javascript">.*<\/script>/).select { |s| s.include?("graphql") }.first
data = JSON.parse(script.scan(/\{.*\}/).first)

data["entry_data"]["ProfilePage"].first["graphql"]["user"]["edge_owner_to_timeline_media"]["edges"].each do |image|
  image = image["node"]
  filename = Time.at(image["taken_at_timestamp"]).strftime("%Y-%m-%d") + "-" + image["id"] + ".json"
  image.merge!(scraper_version: "v2")

  # images may have different dates if collected in another timezone! Only check IDs
  if `ls looted_json | grep #{image["id"]}` == ""
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
