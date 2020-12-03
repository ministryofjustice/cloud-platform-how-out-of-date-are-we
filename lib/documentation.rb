class Documentation < ItemList
  private

  def read_data
    data = super
    unless data.nil?
      list = data.fetch(@key)

      list.each_with_index do |url, i|
        # Turn the URL into site/title/url tuples e.g.
        #   "https://runbooks.cloud-platform.service.justice.gov.uk/create-cluster.html" -> site: "runbooks", title: "create-cluster"
        site, _, _, _, _, title = url.split(".").map { |s| s.sub(/.*\//, "") }
        list[i] = {"site" => site, "title" => title, "url" => url}
      end
    end

    data
  end
end
