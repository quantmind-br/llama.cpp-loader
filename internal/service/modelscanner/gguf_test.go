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

func TestReadGGUFParams_ParameterCount(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(buildGGUFHeader(3, 0, 1))
	writeGGUFString(&buf, "general.parameter_count")
	binary.Write(&buf, binary.LittleEndian, uint32(10)) // type uint64
	binary.Write(&buf, binary.LittleEndian, uint64(32_000_000_000))

	got, err := readGGUFParams(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("readGGUFParams: %v", err)
	}
	if got != "32B" {
		t.Fatalf("params = %q, want %q", got, "32B")
	}
}

func TestReadGGUFParams_SizeLabelString(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(buildGGUFHeader(3, 0, 1))
	writeGGUFString(&buf, "general.size_label")
	binary.Write(&buf, binary.LittleEndian, uint32(8)) // type string
	writeGGUFString(&buf, "7B")

	got, err := readGGUFParams(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("readGGUFParams: %v", err)
	}
	if got != "7B" {
		t.Fatalf("params = %q, want %q", got, "7B")
	}
}

func TestReadGGUFParams_NoMatchReturnsEmpty(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(buildGGUFHeader(3, 0, 1))
	writeGGUFString(&buf, "general.architecture")
	binary.Write(&buf, binary.LittleEndian, uint32(8)) // type string
	writeGGUFString(&buf, "qwen")

	got, err := readGGUFParams(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("readGGUFParams: %v", err)
	}
	if got != "" {
		t.Fatalf("params = %q, want empty", got)
	}
}

func TestReadGGUFParams_UnsupportedTypeAborts(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(buildGGUFHeader(3, 0, 2))
	writeGGUFString(&buf, "tokenizer.ggml.tokens")
	binary.Write(&buf, binary.LittleEndian, uint32(9)) // type array — unsupported, abort scan
	// Don't bother appending payload; reader should bail out at type check.

	got, err := readGGUFParams(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("readGGUFParams: %v", err)
	}
	if got != "" {
		t.Fatalf("params = %q, want empty (unsupported type aborts)", got)
	}
}

// writeGGUFString writes the GGUF string format: u64 length + utf8 bytes.
func writeGGUFString(buf *bytes.Buffer, s string) {
	binary.Write(buf, binary.LittleEndian, uint64(len(s)))
	buf.WriteString(s)
}

func TestFormatParams(t *testing.T) {
	cases := []struct {
		n    uint64
		want string
	}{
		{32_000_000_000, "32B"},
		{7_000_000_000, "7B"},
		{6_700_000_000, "6.7B"},
		{1_950_000_000, "2B"},   // tenths == 10 must roll into whole
		{1_949_999_999, "1.9B"}, // round down boundary
		{500_000_000, "500M"},
		{12_345, "12345"},
	}
	for _, tc := range cases {
		got := formatParams(tc.n)
		if got != tc.want {
			t.Errorf("formatParams(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
}
