class HostedServices < ItemList
  def todo_count
    namespace_details.inject(0) {|sum, (namespace, items)| sum += items.size}
  end
end