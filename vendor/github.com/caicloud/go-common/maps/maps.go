package maps

// DeepCopyStringMap returns a copy of the original map object.
func DeepCopyStringMap(src map[string]string) map[string]string {
	re := make(map[string]string, len(src))
	for k, v := range src {
		re[k] = v
	}
	return re
}
