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

  def github_team_url(team_name)
    link_to(team_name.to_s,GITHUB_TEAM_URL+team_name.to_s)
  end

  def link_each_to(domain_names)
    domain_names.map { |domain_name| link_to(domain_name,"https://"+domain_name) }
                .join("<br> ")
  end

  def link_to(text, href)
    "<a href='#{href}'>#{split(text)}</a>" 
  end

  def split(text)
    text.to_s.split('/').last
  end

end

helpers HoodawUtils
