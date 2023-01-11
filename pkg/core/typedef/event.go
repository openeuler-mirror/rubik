package typedef

type (
	EventType int8
	Event     interface{}
)

const (
	ADD EventType = iota
	UPDATE
	DELETE
)

func (t EventType) String() string {
	switch t {
	case ADD:
		return "add"
	case UPDATE:
		return "update"
	case DELETE:
		return "delete"
	default:
		return "unknown"
	}
}
