package lz77

type position struct {
	offset int
	count  int
}

type link struct {
	offset int
	count  int
	next   *byte
}
