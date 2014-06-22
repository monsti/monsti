package main

// inStringSlice checks if the string value is in the given string slice.
func inStringSlice(value string, slice []string) bool {
	for _, v := range slice {
		if value == v {
			return true
		}
	}
	return false
}
