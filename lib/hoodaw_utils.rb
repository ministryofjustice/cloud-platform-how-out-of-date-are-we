module HoodawUtils
  GITHUB_TEAM_URL = "https://github.com/orgs/ministryofjustice/teams/"
  # "foo_bar_baz" => "Foo Bar Baz"
  def snake_case_to_capitalised(str)
    str.split("_").map(&:capitalize).join(" ")
  end

  def decimals(value, decimals = 2)
    sprintf("%0.#{decimals}f", value)
  end

  def commify(value)
    whole, decimal = decimals(value).split(".")
    with_commas = whole.gsub(/\B(?=(...)*\b)/, ",")
    [with_commas, decimal].join(".")
  end

  def link_to_github_team(team_name)
    unless team_name.nil?
      link_to(team_name.to_s, GITHUB_TEAM_URL + team_name.to_s)
    end
  end

  def links_to_domain_names(domain_names)
    domain_names.map { |domain_name| link_to(domain_name, "https://" + domain_name) }
      .join("<br/> ")
  end

  def link_to_repo(url)
    name = url.to_s.split("/").last
    link_to(name, url)
  end

  def link_to(text, href)
    "<a href='#{href}'>#{text}</a>"
  end
end

helpers HoodawUtils
