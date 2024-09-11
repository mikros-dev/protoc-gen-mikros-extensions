package converters

type MessageKind int

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
