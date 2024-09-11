package context

type FieldLocation int32

const (
	FieldLocation_Unknown FieldLocation = iota
	FieldLocation_Body
	FieldLocation_Query
	FieldLocation_Path
	FieldLocation_Header
)

func (l FieldLocation) String() string {
	switch l {
	case FieldLocation_Body:
		return "body"

	case FieldLocation_Query:
		return "query"

	case FieldLocation_Path:
		return "path"

	case FieldLocation_Header:
		return "header"

	default:
	}

	return "unknown"
}
