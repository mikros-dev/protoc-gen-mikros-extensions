package context

// FieldLocation represents the location of a field in an HTTP request,
// such as body, query, path, or header.
type FieldLocation int32

// Supported field locations.
const (
	FieldLocationUnknown FieldLocation = iota
	FieldLocationBody
	FieldLocationQuery
	FieldLocationPath
	FieldLocationHeader
)

func (l FieldLocation) String() string {
	switch l {
	case FieldLocationBody:
		return "body"

	case FieldLocationQuery:
		return "query"

	case FieldLocationPath:
		return "path"

	case FieldLocationHeader:
		return "header"

	default:
	}

	return "unknown"
}
