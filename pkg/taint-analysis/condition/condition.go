package condition

type CondType int

const (
	GT CondType = iota
	LT
	GE
	LE
	EQ
	NEQ
)

type CondOperand struct{}

type Condition struct {
	IsNegated bool
	Type      CondType
	V1        any
	V2        any
}
