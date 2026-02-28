package invoice

type DateField int

const (
	DateFieldTravel DateField = iota
	DateFieldIssue
)

func (f DateField) String() string {
	switch f {
	case DateFieldTravel:
		return "travel"
	case DateFieldIssue:
		return "issue"
	default:
		return "unknown"
	}
}

