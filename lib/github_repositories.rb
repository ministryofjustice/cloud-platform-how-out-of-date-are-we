class GithubRepositories < ItemList

  private

  def read_data
    data = super

    # Discard any github repositories with a "PASS" status - we only care
    # about failures.
    unless data.nil?
      list = data.fetch(@key).reject { |repo| repo["status"] == "PASS" }
      data[@key] = list
    end

    data
  end
end
