#!/usr/bin/env ruby

require "json"
require "open-uri"
require "fileutils"

Dir.glob("completed_json/*").shuffle.map do |file|
  data = JSON.parse(File.read(file))

  puts data["post_url"]

  format = data["is_video"] == true ? "mp4" : "jpg"

  media_file_name = "media/#{file.split("/").last.gsub("json", format)}"

  next if File.exists?(media_file_name)

  File.write(media_file_name, open(data["media_url"]).read)
  FileUtils.touch media_file_name, mtime: data["timestamp"]
end
