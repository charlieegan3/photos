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

excluded_tags = File.readlines("excluded_tags").map(&:chomp).map(&:downcase)
series = File.readlines("series").map(&:chomp).map(&:downcase)
all_tags = Dir.glob("completed_json/*.json").map {|f| JSON.parse(File.read(f))["tags"]}.flatten.uniq.map {|t|t[1..-1]}
reject_pattern = /insta|gram|shots|_|shotz|london|scotland|nature.|photography|filter/
permitted_tags = (all_tags - excluded_tags).reject { |t| t.match(reject_pattern) }

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
  tags.map { |t| t.gsub("#", "").strip.downcase }.uniq
end

def format_location_slug(id, slug)
  if slug == ""
    id
  else
    [slug, id].join("-")
  end
end

`mkdir -p site/content/photos`

class Fixnum
  def ordinalize
    if (11..13).include?(self % 100)
      "#{self}th"
    else
      case self % 10
        when 1; "#{self}st"
        when 2; "#{self}nd"
        when 3; "#{self}rd"
        else    "#{self}th"
      end
    end
  end
end

plurals = {
  january: "Januaries",
  february: "Februaries",
  march: "Marches",
  april: "Aprils",
  may: "Mays",
  june: "Junes",
  july: "Julys",
  august: "Augusts",
  september: "Septembers",
  october: "Octobers",
  november: "Novembers",
  december: "Decembers",
}

archives = {}
posts = Dir.glob("completed_json/*").sort
posts.each_with_index do |file, index|
  @file_reference = file.split("/").last.gsub(".json", "")
  @data = JSON.parse(File.read(file))
  @data["tags"] = (format_tags(@data["tags"]) & permitted_tags)
  @data["tags"] << "video" if @data["is_video"]
  @title = format_title(@data)
  if @data["location"]
    @location_slug = format_location_slug(@data["location"]["id"], @data["location"]["slug"])
  end
  @data["previous"] = index == 0 ? nil : posts[index-1].split("/").last.sub(".json", "")
  @data["next"] = posts[index+1].split("/").last.sub(".json", "") rescue nil

  year, month, month_string, day, week_day = DateTime.strptime(@data["timestamp"].to_s,'%s').strftime("%Y %m %B %d %A").split(" ")
  archive = {
    "#{year}" => { title: "Year of #{year}", related: [], class: "year" },
    "#{month}" => { title: "All #{plurals[month_string.downcase.to_sym]}", related: [], class: "month" },
    "#{week_day.downcase}" => { title: "Taken on a #{week_day}", related: [], class: "week-day" },
    "#{year}-#{month}" => { title: "#{month_string} #{year}", related: ["#{year}", "#{month}"], class: "year-month" },
    "#{month}-#{day}" => { title: "All #{[month_string.capitalize, day.to_i.ordinalize].join(" ")}s", related: ["#{month}"], class: "month-day" },
    "#{year}-#{month}-#{day}" => { title: DateTime.strptime(@data["timestamp"].to_s,'%s').strftime("%A, %B #{day.to_i.ordinalize}, %Y"), related: ["#{year}-#{month}", "#{month}-#{day}", "#{week_day.downcase}", "#{month}", "#{year}"], class: "day" },
  }
  @data["archive"] = archive.keys
  archives.merge!(archive)

  markdown = ERB.new(page_template).result()
  File.write("site/content/photos/#{@file_reference}.md", markdown)
end

years = archives.select { |k,v| v[:class] == "year" }.keys
months = archives.select { |k,v| v[:class] == "month" }.keys
week_days = %w(monday tuesday wednesday thursday friday saturday sunday)

archives.select { |k,v| v[:class] == "year" }.map { |k, _| archives[k][:related] = (years - [k]).sort.reverse }
archives.select { |k,v| v[:class] == "month" }.map { |k, _| archives[k][:related] = (months - [k]).sort }
archives.select { |k,v| v[:class] == "week-day" }.map { |k, _| archives[k][:related] = (week_days - [k]) }

`rm -r site/content/archive || true`
archives.each do |slug, data|
  data = JSON.parse(JSON.generate(data))
  `mkdir -p site/content/archive/#{slug}/`
  File.write("site/content/archive/#{slug}/_index.md", data.to_yaml + "---\n")
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
