#!/usr/bin/env ruby

# build the cli
system("make -C cli get-binary")
# build the frontend app
system("make -C frontend vue_build")

# copy app to public
system("rm -rf public")
system("mkdir -p public")
system("cp -r frontend/dist/* public")
# make the output dir
system("mkdir -p public/data")
# generate the site data
system("./cli/photos site debug --output public/data")

# commit the result
email = `git config --global user.email`.chomp
name = `git config --global user.name`.chomp
if name == "" || email == ""
  puts "setting gh actions git identity"
  system('git config --global user.email "githubactions@example.com"')
  system('git config --global user.name "GitHub Actions"')
end
system("git checkout -b netlify")
system("git add public")
system("git -c commit.gpgsign=false commit -m generate-site")
system("git push -f origin netlify")
