package modelscanner

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// ggufMagic is the 4-byte file signature.
var ggufMagic = [4]byte{'G', 'G', 'U', 'F'}

// errBadMagic is returned when the first 4 bytes do not match "GGUF".
var errBadMagic = errors.New("modelscanner: not a GGUF file (bad magic)")

// ggufHeader holds the fixed-width prefix every GGUF file starts with.
type ggufHeader struct {
	Version       uint32
	TensorCount   uint64
	MetadataCount uint64
}

// readGGUFHeader reads and validates the GGUF magic + version + counts.
// It does NOT advance into metadata KV pairs.
func readGGUFHeader(r io.Reader) (ggufHeader, error) {
	var magic [4]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		return ggufHeader{}, fmt.Errorf("read magic: %w", err)
	}
	if magic != ggufMagic {
		return ggufHeader{}, errBadMagic
	}
	var hdr ggufHeader
	if err := binary.Read(r, binary.LittleEndian, &hdr.Version); err != nil {
		return ggufHeader{}, fmt.Errorf("read version: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr.TensorCount); err != nil {
		return ggufHeader{}, fmt.Errorf("read tensor count: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr.MetadataCount); err != nil {
		return ggufHeader{}, fmt.Errorf("read metadata count: %w", err)
	}
	return hdr, nil
}
