class OrphanedResources < ItemList
  def todo_count
    list.inject(0) { |sum, (resource_type, items)| sum += items.size }
  end
end
