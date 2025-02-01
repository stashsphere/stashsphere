package utils

func Contains(input []string, needle string) bool {
	for _, e := range input {
		if e == needle {
			return true
		}
	}
	return false
}
