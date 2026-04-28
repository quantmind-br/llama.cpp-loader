package domain

// FlagType enumerates the supported llama-server flag value types.
type FlagType int

const (
	FlagTypeBool FlagType = iota
	FlagTypeInt
	FlagTypeFloat
	FlagTypeString
	FlagTypeEnum
)

// FlagSpec describes a single llama-server flag (filled by LlamaHelpParser in slice 2).
type FlagSpec struct {
	Long       string
	Short      string
	Type       FlagType
	EnumValues []string
	Default    any
	HelpText   string
	Group      string // "essential" or "advanced"
}

// FlagSchema is the parsed --help output keyed by long name.
type FlagSchema struct {
	Version string
	Flags   map[string]FlagSpec
}
