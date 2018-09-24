#!/usr/bin/env ruby

require "json"
require "date"
require "open-uri"

Dir.glob("looted_json/*").shuffle.map do |file|
  completed_file_name = "completed_json/#{file.split("/").last}"

  next if File.exists?(completed_file_name)
  puts file

  raw_data = JSON.parse(File.read(file))

  code = raw_data["code"] || raw_data["shortcode"]
  timestamp = raw_data["date"] || raw_data["taken_at_timestamp"]

  doc = open("https://www.instagram.com/p/#{code}").read
  page_data = doc.scan(/\{[^\n]+\}/).map { |r| JSON.parse(r) rescue nil }.compact.first

  tags = page_data["caption"].scan(/#\w+/).uniq

  location_url = page_data["contentLocation"]["mainEntityofPage"]["@id"].split("/").reject { |e| e.to_s.length == 0 }
  location = {
    "id" => location_url[-2],
    "name" => page_data["contentLocation"]["name"],
      "slug" => location_url.last,
    "has_public_page" => true
  }

  completed_data = {
    id: raw_data["id"],
    code: code,
    display_url: raw_data["display_url"],
    media_url: raw_data["is_video"] ? raw_data["video_url"] : raw_data["display_url"],
    post_url: "https://www.instagram.com/p/#{code}",
    is_video: raw_data["is_video"] == true,
    caption: page_data["caption"],
    location: location,
    tags: tags,
    timestamp: timestamp,
    dimensions: raw_data["dimensions"]
  }

  File.write(completed_file_name, JSON.pretty_generate(completed_data))
end
