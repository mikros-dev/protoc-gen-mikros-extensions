package mapping

// MessageKind represents the type of message.
type MessageKind int

// Supported message kinds.
const (
	UnknownMessageKind MessageKind = iota
	Wire
	WireInput
	WireOutput
)

var (
	kinds = map[MessageKind]string{
		Wire:       "wire",
		WireInput:  "wire_input",
		WireOutput: "wire_output",
	}
)

func (k MessageKind) String() string {
	if s, ok := kinds[k]; ok {
		return s
	}

	return "unknown"
}
