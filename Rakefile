# frozen_string_literal: true

# Copyright (c) 2025 kk
#
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT

require 'time'

task default: %w[release_push]

desc '自动叠加版本号并推送 release'
task :release_push do
  # 获取最新 tag
  latest_tag = `git tag --list 'v*' --sort=-v:refname`.lines.first&.strip
  if latest_tag.nil? || latest_tag.empty?
    new_tag = 'v1.0.0'
  else
    major, minor, patch = latest_tag[1..].split('.').map(&:to_i)
    patch += 1
    new_tag = "v#{major}.#{minor}.#{patch}"
  end

  puts "[kver] New release tag: #{new_tag}"

  system 'git add .'
  system "git commit -m 'Release #{new_tag} at #{Time.now.utc.iso8601}'"
  system 'git pull'
  system 'git push origin main'

  system "git tag #{new_tag}"
  system "git push origin #{new_tag}"

  puts "[kver] Release #{new_tag} pushed!"
end
