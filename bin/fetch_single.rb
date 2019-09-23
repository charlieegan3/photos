#!/usr/bin/env ruby

require "json"
require "open-uri"

raw = open(ARGV[0]).read
script = raw.scan(/<script type="text\/javascript">.*<\/script>/).select { |s| s.include?("graphql") }.first
data = JSON.parse(script.scan(/\{.*\}/).first)
image = data["entry_data"]["PostPage"].first["graphql"]["shortcode_media"]

filename = Time.at(image["taken_at_timestamp"]).strftime("%Y-%m-%d") + "-" + image["id"] + ".json"
image.merge!(scraper_version: "v2")

# images may have different dates if collected in another timezone! Only check IDs
if `ls looted_json | grep #{image["id"]}` == ""

  File.write("looted_json/" + filename, JSON.pretty_generate(image, indent: "    "))
  puts filename
end
