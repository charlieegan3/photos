#!/usr/bin/env ruby

require "json"
require "open-uri"
require "fileutils"

Dir.glob("completed_json/*").shuffle.map do |file|
  data = JSON.parse(File.read(file))

  format = data["is_video"] == true ? "mp4" : "jpg"

  media_file_name = "media/#{file.split("/").last.gsub("json", format)}"

  next if File.exists?(media_file_name)
  puts data["post_url"]

  begin
    File.write(media_file_name, open(data["media_url"]).read)
    FileUtils.touch media_file_name, mtime: data["timestamp"]
  rescue
    puts "#{media_file_name} Failed"
    File.delete(media_file_name) if File.exists?(media_file_name)
  end
end
