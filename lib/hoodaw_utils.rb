module HoodawUtils
  # "foo_bar_baz" => "Foo Bar Baz"
  def snake_case_to_capitalised(str)
    str.split("_").map(&:capitalize).join(" ")
  end
end

helpers HoodawUtils
