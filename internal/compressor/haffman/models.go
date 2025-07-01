package haffman

type frequencyMap map[byte]int

type node struct {
	data      byte
	frequency int
	left      *node
	right     *node
}
