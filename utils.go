package bear

func parseStructColumns(i interface{}, ignore ...string) []string {
	panic("not implemented")
}

func repeatString(s string, n int) []string {
	r := make([]string, 0, n)
	for i := 0; i < n; i++ {
		r = append(r, s)
	}
	return r
}
