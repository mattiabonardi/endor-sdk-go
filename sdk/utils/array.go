package utils

func StringElemMatch(slice []string, target string) bool {
	for _, id := range slice {
		if id == target {
			return true
		}
	}
	return false
}
