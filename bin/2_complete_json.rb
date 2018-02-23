#!/usr/bin/env ruby

require "json"
require "date"
require "open-uri"

Dir.glob("looted_json/*").shuffle.map do |file|
  completed_file_name = "completed_json/#{file.split("/").last}"

  next if File.exists?(completed_file_name)
  puts file

  raw_data = JSON.parse(File.read(file))

  doc = open("https://www.instagram.com/p/#{raw_data["code"]}").read
  page_data = doc.scan(/\{[^\n]+\}/).map { |r| JSON.parse(r) rescue nil }.compact.first
  graph_image = page_data.dig(*%w(entry_data PostPage)).first.dig(*%w(graphql shortcode_media))

  caption = graph_image["edge_media_to_caption"]["edges"].first["node"]["text"] rescue ""
  tags = caption.scan(/#\w+/).uniq

  completed_data = {
    id: raw_data["id"],
    code: raw_data["code"],
    display_url: graph_image["display_url"],
    media_url: raw_data["is_video"] ? graph_image["video_url"] : graph_image["display_url"],
    post_url: "https://www.instagram.com/p/#{raw_data["code"]}",
    is_video: raw_data["is_video"] == true,
    caption: caption,
    location: graph_image["location"],
    tags: tags,
    timestamp: raw_data["date"],
    dimensions: raw_data["dimensions"]
  }


  File.write(completed_file_name, JSON.pretty_generate(completed_data))
end
