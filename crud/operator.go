package crud

type Operator string

const (
	OP_EQ   Operator = "eq"
	OP_NE   Operator = "ne"
	OP_GT   Operator = "gt"
	OP_GE   Operator = "ge"
	OP_LT   Operator = "lt"
	OP_LE   Operator = "le"
	OP_LIKE Operator = "like"
)

func (op Operator) Valid() bool {
	switch op {
	case OP_EQ:
		return true
	case OP_NE:
		return true
	case OP_GT:
		return true
	case OP_GE:
		return true
	case OP_LT:
		return true
	case OP_LE:
		return true
	case OP_LIKE:
		return true
	}
	return false
}
