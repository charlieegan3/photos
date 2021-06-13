#!/usr/bin/env ruby

def run(command)
  puts "running: #{command}"
  fail unless system(command)
end

# build the frontend app
run("make -C frontend vue_build")

# copy app to public
run("rm -rf public")
run("mkdir -p public")
run("cp -r frontend/dist/* public")

# make the output data dir
run("mkdir -p public/data")
# generate the site data
run("photos site debug --output public/data")

# commit the result
email = `git config --global user.email`.chomp
name = `git config --global user.name`.chomp
if name == "" || email == ""
  puts "setting gh actions git identity"
  run('git config --global user.email "githubactions@example.com"')
  run('git config --global user.name "GitHub Actions"')
end
run("git checkout -b netlify")
run("git add public")
run("git -c commit.gpgsign=false commit -m generate-site")
run("git push -f origin netlify")
