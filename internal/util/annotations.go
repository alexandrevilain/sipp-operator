package util

// MergeAnnotations merges the provided annotations
func MergeAnnotations(annotations ...map[string]string) map[string]string {
	result := map[string]string{}

	for _, keyValue := range annotations {
		for key, value := range keyValue {
			result[key] = value
		}
	}

	return result
}
