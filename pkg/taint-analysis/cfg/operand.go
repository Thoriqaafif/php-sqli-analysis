package cfg

type VarScope int

const (
	VAR_SCOPE_GLOBAL VarScope = iota
	VAR_SCOPE_LOCAL
	VAR_SCOPE_OBJECT
	VAR_SCOPE_FUNCTION
)

// TODO: AddAssertion method
type Operand interface {
	GetType() string
	AddUsage(op Op)
	AddWriteOp(op Op)
	RemoveUsage(op Op)
	AddAssertion(op Op)
}

type OperandAttr struct {
	Assertions []any // TODO: change type
	Ops        []Op
	Usages     []Op
}

// TODO: AddWriteOp, RemoveUsage, AddAssertion
// func operandGetType() string {
// 	return ""
// }

func operandAddUsage(oper OperandAttr, op Op) {}

func operandAddWriteOp(oper OperandAttr, op Op) {
	oper.Ops = append(oper.Ops, op)
}

func operandRemoveUsage(oper OperandAttr, op Op) {}

func operandAddAssertion(oper OperandAttr, op Op) {}

type Literal struct {
	OperandAttr
	Val any
}

func NewLiteral(val any) *Literal {
	return &Literal{
		OperandAttr: OperandAttr{
			Assertions: make([]any, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
		},
		Val: val,
	}
}

// TODO: remove if not used
func (o *Literal) GetType() string {
	return "Literal"
}

func (o *Literal) AddUsage(op Op) {
	operandAddUsage(o.OperandAttr, op)
}

func (o *Literal) AddWriteOp(op Op) {
	operandAddWriteOp(o.OperandAttr, op)
}

func (o *Literal) RemoveUsage(op Op) {
	operandRemoveUsage(o.OperandAttr, op)
}

func (o *Literal) AddAssertion(op Op) {
	operandAddAssertion(o.OperandAttr, op)
}

type Var struct {
	OperandAttr
	Name Operand
}

func NewVar(name Operand) *Var {
	return &Var{
		OperandAttr: OperandAttr{
			Assertions: make([]any, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
		},
		Name: name,
	}
}

func (o *Var) GetType() string {
	return "Variable"
}

func (o *Var) AddUsage(op Op) {
	operandAddUsage(o.OperandAttr, op)
}

func (o *Var) AddWriteOp(op Op) {
	operandAddWriteOp(o.OperandAttr, op)
}

func (o *Var) RemoveUsage(op Op) {
	operandRemoveUsage(o.OperandAttr, op)
}

func (o *Var) AddAssertion(op Op) {
	operandAddAssertion(o.OperandAttr, op)
}

type BoundVar struct {
	OperandAttr
	Name  Operand
	ByRef bool
	Scope VarScope
	Extra any // TODO: check functionality
}

// Variable immune to SSA
func NewBoundVar(name Operand, byref bool, scope VarScope) *BoundVar {
	return &BoundVar{
		OperandAttr: OperandAttr{
			Assertions: make([]any, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
		},
		Name:  name,
		ByRef: byref,
		Scope: scope,
	}
}

func (o *BoundVar) GetType() string {
	return "Bound_Variable"
}

func (o *BoundVar) AddUsage(op Op) {
	operandAddUsage(o.OperandAttr, op)
}

func (o *BoundVar) AddWriteOp(op Op) {
	operandAddWriteOp(o.OperandAttr, op)
}

func (o *BoundVar) RemoveUsage(op Op) {
	operandRemoveUsage(o.OperandAttr, op)
}

func (o *BoundVar) AddAssertion(op Op) {
	operandAddAssertion(o.OperandAttr, op)
}

type Temporary struct {
	OperandAttr
	Original Operand
}

func NewTemporary(original Operand) *Temporary {
	return &Temporary{
		OperandAttr: OperandAttr{
			Assertions: make([]any, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
		},
		Original: original,
	}
}

func NewEmptyTemporary() *Temporary {
	return &Temporary{
		OperandAttr: OperandAttr{
			Assertions: make([]any, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
		},
		Original: nil,
	}
}

func (o *Temporary) GetType() string {
	return "Temporary"
}

func (o *Temporary) AddUsage(op Op) {
	operandAddUsage(o.OperandAttr, op)
}

func (o *Temporary) AddWriteOp(op Op) {
	operandAddWriteOp(o.OperandAttr, op)
}

func (o *Temporary) RemoveUsage(op Op) {
	operandRemoveUsage(o.OperandAttr, op)
}

func (o *Temporary) AddAssertion(op Op) {
	operandAddAssertion(o.OperandAttr, op)
}

type NullOper struct {
	OperandAttr
}

func NewNull() *NullOper {
	return &NullOper{
		OperandAttr: OperandAttr{
			Assertions: make([]any, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
		},
	}
}

func (o *NullOper) GetType() string {
	return "Null"
}

func (o *NullOper) AddUsage(op Op) {
	operandAddUsage(o.OperandAttr, op)
}

func (o *NullOper) AddWriteOp(op Op) {
	operandAddWriteOp(o.OperandAttr, op)
}

func (o *NullOper) RemoveUsage(op Op) {
	operandRemoveUsage(o.OperandAttr, op)
}

func (o *NullOper) AddAssertion(op Op) {
	operandAddAssertion(o.OperandAttr, op)
}
