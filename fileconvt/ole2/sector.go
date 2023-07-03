package ole2

import (
	"bytes"
	"encoding/binary"
)

type Sector []byte

func (s *Sector) Uint32(bit uint32) uint32 {
	return binary.LittleEndian.Uint32((*s)[bit : bit+4])
}

func (s *Sector) NextSid(size uint32) int32 {
	return int32(s.Uint32(size - 4))
}

func (s *Sector) MsatValues(size uint32) []int32 {

	return s.values(size, int(size/4-1))
}

func (s *Sector) AllValues(size uint32) []int32 {

	return s.values(size, int(size/4))
}

func (s *Sector) values(size uint32, length int) []int32 {

	var res = make([]int32, length)

	buf := bytes.NewBuffer((*s))

	binary.Read(buf, binary.LittleEndian, res)

	return res
}
