package slice

// StringInSlice returns true if the string exists in the slice.
func StringInSlice(array []string, str string) bool {
	for _, v := range array {
		if v == str {
			return true
		}
	}
	return false
}

// RemoveStringInSlice deletes the string in the given slice.
func RemoveStringInSlice(array []string, str string) []string {
	newArray := make([]string, 0, len(array))
	for _, v := range array {
		if v != str {
			newArray = append(newArray, v)
		}
	}
	return newArray
}

// GetStartLimitEnd returns the valid end position of the array based on the start and limit.
func GetStartLimitEnd(start, limit, arrayLen int) int {
	if limit == 0 { // no limit
		return arrayLen
	}
	end := start + limit
	if end > arrayLen {
		end = arrayLen
	}
	return end
}
