#!/usr/bin/env ruby

require "nokogiri"
require "date"
require "json"
require "net/http"

def page_urls_overdue_for_review(root_url)
  urls_and_review_statuses(root_url, root_url)
    .select { |_url, needs_review| needs_review }
    .keys
end

# Spider a URL looking for links to other pages.
# Return a hash: { unique_page_url => needs_review? }
def urls_and_review_statuses(url, root_url, seen = {})
  return if seen.key?(url)

  doc = get_nokogiri_doc(url)
  seen[url] = needs_review?(doc)
  get_links(doc).map { |path| urls_and_review_statuses([root_url, path].join, root_url, seen) }
  
  seen
end

def needs_review?(doc)
  div = page_expiry_div(doc)

  return false if div.nil?

  # NOTE: 'last-reviewed-on' is overloaded by the gov.uk tech-docs gem. It is actually 'review due date'.
  review_required_date = Date.parse(div.attributes["data-last-reviewed-on"].value)

  review_required_date < Date.today
end

def get_links(doc)
  doc.css("a")
    .map { |link| link.attributes.dig("href").value }
    .map { |href| normalise_href(href) }.compact
    .uniq
end

def get_nokogiri_doc(url)
  page = fetch_url(url)
  Nokogiri::HTML(page.body)
end

def fetch_url(url)
  uri = URI.parse(url)
  Net::HTTP.get_response(uri)
end

def page_expiry_div(doc)
  doc.search("div")
    .select { |div| div.attributes.include?("data-module") }
    .select { |div| div.attributes["data-module"].value == "page-expiry" }
    .first
end

# Return a normalised href value, or nil if we don't like this href
def normalise_href(href)
  # ignore links which are external, or to in-page anchors
  return nil if href[0] == "#" || ["/", "http", "mail", "/ima"].include?(href[0, 4])

  # Remove any trailing anchors, or "/" and leading "./" 
  target = href.sub(/\#.*/, "").sub(/\/$/, "").sub(/\.*/, "").sub(/\/*/, "")

  # Ignore links which don't point to html files
  /html$/.match?(target) ? target : nil
end

############################################################

# DOCUMENTATION_SITES should be something like:
# export DOCUMENTATION_SITES="https://runbooks.cloud-platform.service.justice.gov.uk https://user-guide.cloud-platform.service.justice.gov.uk"
# NB: When this becomes too long for a bash string, we will need to
# use an alternative mechanism for passing the list of URLs
sites = ENV.fetch("DOCUMENTATION_SITES").split(" ")

pages = sites.map { |url| page_urls_overdue_for_review(url) }.flatten

puts({
  pages: pages,
  updated_at: Time.now,
}.to_json)
