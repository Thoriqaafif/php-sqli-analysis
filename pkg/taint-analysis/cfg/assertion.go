package cfg

type AssertMode int

const (
	ASSERT_MODE_INTERSECT AssertMode = iota + 192
	ASSERT_MODE_UNION
)

type Assertion interface {
	GetNegation() Assertion
	Negated() bool
}

type TypeAssertion struct {
	Val Operand

	IsNegated bool
}

func NewTypeAssertion(val Operand, isNegated bool) *TypeAssertion {
	return &TypeAssertion{
		Val:       val,
		IsNegated: isNegated,
	}
}

func (a *TypeAssertion) GetNegation() Assertion {
	return &TypeAssertion{
		Val:       a.Val,
		IsNegated: !a.IsNegated,
	}
}

func (a *TypeAssertion) Negated() bool {
	return a.IsNegated
}

type CompositeAssertion struct {
	Val  []Assertion
	Mode AssertMode

	IsNegated bool
}

func NewCompositeAssertion(val []Assertion, mode AssertMode, isNegated bool) *CompositeAssertion {
	return &CompositeAssertion{
		Val:       val,
		Mode:      mode,
		IsNegated: isNegated,
	}
}

func (a *CompositeAssertion) GetNegation() Assertion {
	return &CompositeAssertion{
		Val:       a.Val,
		Mode:      a.Mode,
		IsNegated: !a.IsNegated,
	}
}

func (a *CompositeAssertion) Negated() bool {
	return a.IsNegated
}
