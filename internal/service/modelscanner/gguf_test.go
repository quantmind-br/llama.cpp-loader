package modelscanner

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// buildGGUFHeader returns a minimal valid GGUF header bytestream:
// "GGUF" + version + tensorCount + metaCount.
// Caller appends KV bytes if needed.
func buildGGUFHeader(version uint32, tensorCount, metaCount uint64) []byte {
	var buf bytes.Buffer
	buf.WriteString("GGUF")
	binary.Write(&buf, binary.LittleEndian, version)
	binary.Write(&buf, binary.LittleEndian, tensorCount)
	binary.Write(&buf, binary.LittleEndian, metaCount)
	return buf.Bytes()
}

func TestReadGGUFHeader_ValidMagic(t *testing.T) {
	data := buildGGUFHeader(3, 0, 0)
	hdr, err := readGGUFHeader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("readGGUFHeader: unexpected err: %v", err)
	}
	if hdr.Version != 3 {
		t.Fatalf("Version = %d, want 3", hdr.Version)
	}
	if hdr.MetadataCount != 0 {
		t.Fatalf("MetadataCount = %d, want 0", hdr.MetadataCount)
	}
}

func TestReadGGUFHeader_BadMagic(t *testing.T) {
	data := []byte("NOPEXXXXXXXXXXXXXXXXXXXXXXXX")
	if _, err := readGGUFHeader(bytes.NewReader(data)); err == nil {
		t.Fatal("expected error on bad magic, got nil")
	}
}

func TestReadGGUFHeader_Truncated(t *testing.T) {
	data := []byte("GGUF\x03\x00") // magic + 2 bytes of version
	if _, err := readGGUFHeader(bytes.NewReader(data)); err == nil {
		t.Fatal("expected error on truncated header, got nil")
	}
}
