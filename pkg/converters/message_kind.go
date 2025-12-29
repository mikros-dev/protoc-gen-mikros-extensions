package converters

// MessageKind represents the type of message.
type MessageKind int

// Supported message kinds.
const (
	UnknownMessageKind MessageKind = iota
	WireMessage
	WireInputMessage
	WireOutputMessage
)

func (k MessageKind) String() string {
	switch k {
	case WireMessage:
		return "wire"
	case WireInputMessage:
		return "wire_input"
	case WireOutputMessage:
		return "wire_output"
	default:
	}

	return "unknown"
}
