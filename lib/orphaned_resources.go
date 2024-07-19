package lib

// OrphanedResources represents a list of orphaned resources
type OrphanedResources struct {
	List map[string][]string
}

// TodoCount calculates the total number of items across all resource types
func (or *OrphanedResources) TodoCount() int {
	sum := 0
	for _, items := range or.List {
		sum += len(items)
	}
	return sum
}
