package reader

func removeElementFromSlice(slice []string, i int) []string {
	return append(slice[:i], slice[i+1:]...)
}
