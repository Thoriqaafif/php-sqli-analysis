package cfg

type Op interface {
	AddReadRef(o Operand) Operand
	AddWriteRef(o Operand) Operand
}

type OpGeneral struct{}

func (op *OpGeneral) AddReadRefs(opers ...Operand) []Operand {
	result := make([]Operand, 0)
	for _, oper := range opers {
		result = append(result, op.AddReadRef(oper))
	}
	return result
}

func (op *OpGeneral) AddReadRef(oper Operand) Operand {
	oper.AddUsage(op)
	return oper
}

func (op *OpGeneral) AddWriteRef(oper Operand) Operand {
	oper.AddWriteOp(op)
	return oper
}

type Phi struct {
	OpGeneral
	Vars   map[Operand]struct{}
	Result Operand
}

func NewPhi(result Operand) *Phi {
	op := &Phi{
		Vars:   make(map[Operand]struct{}, 0),
		Result: result,
	}
	op.AddWriteRef(result)
	return op
}

// add an operand to phi vars, if not exist
func (op *Phi) AddOperand(oper Operand) {
	var empty struct{}
	if _, ok := op.Vars[oper]; !ok && op.Result != oper {
		op.Vars[oper] = empty
	}
}

// remove an operand from phi vars
func (op *Phi) RemoveOperand(oper Operand) {
	if _, ok := op.Vars[oper]; ok {
		oper.RemoveUsage(op)
		delete(op.Vars, oper)
	}
}

// function parameter
type Param struct {
	OpGeneral
}

func NewParam(result Operand) *Param {
	op := &Param{}
	op.AddWriteRef(result)
	return op
}

// TODO: Find if it useful
// type CallableOp interface {
// 	GetFunc() Func
// }
