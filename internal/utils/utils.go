package utils

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func Filter(slice []string, f func(string) bool) []string {
	var r []string
	for _, s := range slice {
		if f(s) {
			r = append(r, s)
		}
	}
	return r
}
