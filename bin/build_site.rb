#!/usr/bin/env ruby

require "json"
require "yaml"
require "date"
require "erb"

page_template = <<ERB
<%= @data.to_yaml %>
date: <%= DateTime.strptime(@data["timestamp"].to_s,'%s').iso8601 %>
title: |-
  <%= @title %>
file_reference: <%= @file_reference %>
locations:
<% if @data["location"] %> - <%= @location_slug %><% end %>
---

<%= format_caption(@data["caption"]) %>
ERB

def format_caption(caption)
  if caption.split("\n").reject(&:empty?).length > 1
    caption = caption.split("\n").each_with_index.
      reject { |e, i| i > 0 && (e.strip.length < 3 || e[0] == "#") }.map(&:first).join("\n\n")
  end

  caption.gsub!(/(#\S+ ?)+$/, "")

  caption.strip.gsub(/ \./, "")
end

def format_title(data)
  if data["location"]
    if data["location"]["name"]
      return data["location"]["name"]
    end
  end

  data["caption"] = data["caption"].split("\n").join(" ")

  if data["caption"].length > 15
    return data["caption"][0..14] + "..."
  elsif data["caption"].length > 0
    return data["caption"]
  end

  return "Pretty Picture"
end

def format_tags(tags)
  tags.map { |t| t.gsub("#", "").strip }
end

def format_location_slug(id, slug)
  if slug == ""
    id
  else
    [slug, id].join("-")
  end
end

`mkdir -p site/content/photos`

Dir.glob("completed_json/*").shuffle.each do |file|
  @file_reference = file.split("/").last.gsub(".json", "")
  @data = JSON.parse(File.read(file))
  @data["tags"] = format_tags(@data["tags"])
  @title = format_title(@data)
  if @data["location"]
    @location_slug = format_location_slug(@data["location"]["id"], @data["location"]["slug"])
  end

  markdown = ERB.new(page_template).result()
  File.write("site/content/photos/#{@file_reference}.md", markdown)
end

`rm -r site/content/locations/*`
locations = []
Dir.glob("locations/*").shuffle.each do |file|
  locations << JSON.parse(File.read(file))
end

def distance(point1, point2)
  rad = 6_371_000
  lat1, lat2, lon1, lon2 = [point1.first, point2.first, point1.last, point2.last].map { |v| v * Math::PI / 180 }
  delta_lat = (lat1 - lat2).abs
  delta_lon = (lon1 - lon2).abs

  a = Math.sin(delta_lat / 2) * Math.sin(delta_lat / 2) +
      Math.cos(lat1) * Math.cos(lat2) *
      Math.sin(delta_lon / 2) * Math.sin(delta_lon / 2)

  c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a))

  rad * c
end

locations.map! do |l1|
  near = (locations - [l1]).map { |l2|
    { "id" => l2["id"], "distance" => distance([l1["lat"], l1["long"]], [l2["lat"], l2["long"]]) }
  }.select { |l| l["distance"] < 50000 }.sort_by { |l| l["distance"] }
  l1.merge({"near" => near})
end

locations.each do |location|
  @data = location
  slug = format_location_slug(@data["id"], @data["slug"])

  path = "site/content/locations/#{slug}"
  `mkdir -p #{path}`
  File.write("#{path}/_index.md", @data.to_yaml + "---\n")
end
