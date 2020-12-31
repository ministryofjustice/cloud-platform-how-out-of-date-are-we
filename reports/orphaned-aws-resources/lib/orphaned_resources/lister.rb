module OrphanedResources
  class Lister
    private

    def clean_list(list)
      list
        .flatten
        .uniq { |i| i.respond_to?(:id) ? i.id : i }
        .reject(&:nil?)
        .reject(&:empty?)
        .sort
    end
  end
end
