package utils

func Contains(slice []string, target string) bool {
	for _, i := range slice {
		if i == target {
			return true
		}
	}
	return false
}
