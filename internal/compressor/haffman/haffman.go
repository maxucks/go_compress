package haffman

import (
	"bytes"
	"errors"
	"sort"

	"github.com/maxucks/go_compress.git/internal/compressor/core"
	"github.com/maxucks/go_compress.git/internal/compressor/encoder"
	"github.com/maxucks/go_compress.git/internal/pkg/collections"
)

type HuffmanCompressor struct {
	encoder core.Encoder
}

func NewCompressor() *HuffmanCompressor {
	return &HuffmanCompressor{
		encoder: &encoder.CompactVLQ{},
	}
}

func (c *HuffmanCompressor) Compress(data []byte) (*bytes.Buffer, error) {
	frmap := c.computeFrequencyMap(data)
	root := c.buildTree(frmap)
	codes := c.generateCodes(root, "")

	var buf bytes.Buffer

	bitLen := 0
	for _, b := range data {
		bitLen += len(codes[b])
	}

	if err := c.encodeMeta(&buf, frmap, bitLen); err != nil {
		return nil, err
	}
	err := c.encodeData(&buf, data, codes)
	return &buf, err
}

func (c *HuffmanCompressor) Decompress(buf *bytes.Buffer) ([]byte, error) {
	frmap, bitLen, err := c.decodeMeta(buf)
	if err != nil {
		return nil, err
	}
	root := c.buildTree(frmap)
	return c.decodeData(buf, root, bitLen)
}

func (c *HuffmanCompressor) computeFrequencyMap(data []byte) frequencyMap {
	fr := make(frequencyMap)
	for _, sym := range data {
		fr[sym]++
	}
	return fr
}

func (c *HuffmanCompressor) buildTree(frmap frequencyMap) *node {
	keys := make([]byte, 0, len(frmap))
	for k := range frmap {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	que := collections.NewPriorityQue[*node]()

	for _, sym := range keys {
		n := &node{
			data:      sym,
			frequency: frmap[sym],
		}
		que.Push(&collections.QueItem[*node]{
			Value:    n,
			Priority: n.frequency,
		})
	}

	que.InitHeap()

	for que.Len() > 1 {
		a := que.HeapPop()
		b := que.HeapPop()

		freq := a.Value.frequency + b.Value.frequency

		que.HeapPush(&collections.QueItem[*node]{
			Value: &node{
				frequency: freq,
				left:      a.Value,
				right:     b.Value,
			},
			Priority: freq,
		})
	}

	return que.HeapPop().Value
}

func (c *HuffmanCompressor) generateCodes(node *node, prefix string) map[byte]string {
	codes := make(map[byte]string)
	c.generateCodesRec(node, prefix, codes)
	return codes
}

func (c *HuffmanCompressor) generateCodesRec(node *node, prefix string, codes map[byte]string) {
	if node.left == nil && node.right == nil {
		codes[node.data] = prefix
		return
	}
	c.generateCodesRec(node.left, prefix+"0", codes)
	c.generateCodesRec(node.right, prefix+"1", codes)
}

func (c *HuffmanCompressor) encodeData(buf *bytes.Buffer, data []byte, codes map[byte]string) error {
	var bitStr string

	for _, b := range data {
		bitStr += codes[b]
	}
	var encoded []byte
	var byteVal byte
	bitCount := 0

	for _, bit := range bitStr {
		byteVal <<= 1
		if bit == '1' {
			byteVal |= 1
		}
		bitCount++
		if bitCount == 8 {
			encoded = append(encoded, byteVal)
			byteVal = 0
			bitCount = 0
		}
	}

	if bitCount > 0 {
		byteVal <<= (8 - bitCount)
		encoded = append(encoded, byteVal)
	}

	_, err := buf.Write(encoded)
	return err
}

func (c *HuffmanCompressor) decodeData(buf *bytes.Buffer, root *node, bitLen int) ([]byte, error) {
	encodedData := make([]byte, buf.Len())
	if _, err := buf.Read(encodedData); err != nil {
		return nil, err
	}
	if len(encodedData)*8 < bitLen {
		return nil, errors.New("not enough encoded bits")
	}

	var data []byte
	node := root

	for i := 0; i < bitLen; i++ {
		bytePos := i / 8
		bitPos := 7 - (i % 8)
		bit := (encodedData[bytePos] >> bitPos) & 1

		if bit == 0 {
			node = node.left
		} else {
			node = node.right
		}

		if node.left == nil && node.right == nil {
			data = append(data, node.data)
			node = root
		}
	}

	return data, nil
}

func (c *HuffmanCompressor) encodeMeta(buf *bytes.Buffer, frmap frequencyMap, bitLen int) error {
	if err := c.encoder.EncodeInt(buf, len(frmap)); err != nil {
		return err
	}

	for sym, fr := range frmap {
		if err := c.encoder.EncodeInt(buf, int(sym)); err != nil {
			return err
		}
		if err := c.encoder.EncodeInt(buf, fr); err != nil {
			return err
		}
	}

	return c.encoder.EncodeInt(buf, bitLen)
}

func (c *HuffmanCompressor) decodeMeta(buf *bytes.Buffer) (frequencyMap, int, error) {
	frmapLen, err := c.encoder.DecodeInt(buf)
	if err != nil {
		return nil, 0, err
	}

	frmap := make(frequencyMap, frmapLen)

	for range frmapLen {
		sym, err := c.encoder.DecodeInt(buf)
		if err != nil {
			return nil, 0, err
		}
		fr, err := c.encoder.DecodeInt(buf)
		if err != nil {
			return nil, 0, err
		}
		frmap[byte(sym)] = fr
	}

	bitLen, err := c.encoder.DecodeInt(buf)
	if err != nil {
		return nil, 0, err
	}

	return frmap, bitLen, nil
}
