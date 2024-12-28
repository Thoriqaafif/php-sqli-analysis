package cfg

import (
	"fmt"
	"reflect"
	"strconv"
)

type VarScope int

const (
	VAR_SCOPE_GLOBAL VarScope = iota
	VAR_SCOPE_LOCAL
	VAR_SCOPE_OBJECT
	VAR_SCOPE_FUNCTION
)

type VarAssert struct {
	Var    Operand
	Assert Assertion
}

// TODO: AddAssertion method
type Operand interface {
	AddUsage(op Op)
	AddWriteOp(op Op)
	RemoveUsage(op Op)
	AddAssertion(op Operand, assert Assertion, mode AssertMode)
	GetAssertions() []VarAssert
	GetUsage() []Op
	GetWriteOp() Op
	AddCondUsage(block *Block)
	GetCondUsages() []*Block
	IsTainted() bool
	String() string
	IsWritten() bool
}

func GetOperNamed(oper Operand) *OperString {
	if opT, ok := oper.(*OperTemporary); ok {
		if orig, ok := opT.Original.(*OperVariable); ok {
			if name, ok := orig.Name.(*OperString); ok {
				return name
			}
		}
	}
	return nil
}

func GetOperName(oper Operand) (string, error) {
	switch o := oper.(type) {
	case *OperBoundVar:
		return GetOperName(o.Name)
	case *OperVariable:
		return GetOperName(o.Name)
	case *OperString:
		return o.Val, nil
	case *OperTemporary:
		return GetOperName(o.Original)
	case *OperBool, *OperNumber, *OperNull, *OperSymbolic:
		return "", fmt.Errorf("operand doesn't have name '%v'", reflect.TypeOf(o))
	}
	return "", fmt.Errorf("operand doesn't have name '%v'", reflect.TypeOf(oper))
}

func GetStringOper(oper Operand) (string, bool) {
	if os, ok := oper.(*OperString); ok {
		return os.Val, true
	}
	return "", false
}

func GetOperVal(oper Operand) Operand {
	switch o := oper.(type) {
	case *OperBool, *OperNumber, *OperNull, *OperObject, *OperString, *OperSymbolic:
		return oper
	case *OperTemporary:
		return GetOperVal(o.Original)
	case *OperVariable:
		return GetOperVal(o.Value)
	case *OperBoundVar:
		return GetOperVal(o.Value)
	}

	// operand doesn't have value
	return NewOperNull()
}

func SetOperVal(oper Operand, val Operand) {
	switch o := oper.(type) {
	case *OperTemporary:
		SetOperVal(o.Original, val)
	case *OperVariable:
		o.Value = val
	case *OperBoundVar:
		o.Value = val
	}
}

func IsScalarOper(oper Operand) bool {
	switch oper.(type) {
	case *OperBool, *OperNumber, *OperString:
		return true
	}
	return false
}

// func GetOperVal(oper Operand)

// Basic operand attributes and methods
// - Assertions:
// - Ops: Set of Op define the operand value
// - Usages: Set of Op using the operand value
type OperandAttr struct {
	Assertions []VarAssert
	Ops        []Op // op which define the operand
	Usages     []Op // op which use the operand value
	CondUsages []*Block

	Tainted bool
}

func (oa *OperandAttr) AddUsage(op Op) {
	for _, usage := range oa.Usages {
		if usage == op {
			return
		}
	}
	oa.Usages = append(oa.Usages, op)
}

func (oa *OperandAttr) RemoveUsage(op Op) {
	for i, usage := range oa.Usages {
		// if find, remove
		if usage == op {
			oa.Usages = append(oa.Usages[:i], oa.Usages[i+1:]...)
		}
	}
}

func (oa *OperandAttr) AddWriteOp(op Op) {
	for _, writeOp := range oa.Ops {
		if writeOp == op {
			return
		}
	}

	oa.Ops = append(oa.Ops, op)
}

// TODO: Implement assertion in operand
func (oa *OperandAttr) AddAssertion(oper Operand, assert Assertion, mode AssertMode) {
	// find in the current assertion if there
	// is a curr assertion can be merged
	for i, varAssert := range oa.Assertions {
		if varAssert.Var == oper {
			mergedAssert := []Assertion{varAssert.Assert, assert}
			oa.Assertions[i].Assert = NewCompositeAssertion(mergedAssert, mode, false)
			return
		}
		operName := GetOperNamed(oper)
		varName := GetOperNamed(varAssert.Var)
		if operName != nil && varName != nil && operName.Val == varName.Val {
			mergedAssert := []Assertion{varAssert.Assert, assert}
			oa.Assertions[i].Assert = NewCompositeAssertion(mergedAssert, mode, false)
			return
		}
	}

	// if no one can be merged, append to the list
	oa.Assertions = append(oa.Assertions, VarAssert{Var: oper, Assert: assert})
}

func (oa *OperandAttr) GetAssertions() []VarAssert {
	return oa.Assertions
}

func (oa *OperandAttr) GetUsage() []Op {
	return oa.Usages
}

func (oa *OperandAttr) GetWriteOp() Op {
	if len(oa.Ops) > 0 {
		return oa.Ops[len(oa.Ops)-1]
	}
	return nil
}

func (oa *OperandAttr) AddCondUsage(newBlock *Block) {
	for _, block := range oa.CondUsages {
		if newBlock == block {
			return
		}
	}
	oa.CondUsages = append(oa.CondUsages, newBlock)
}

func (oa *OperandAttr) GetCondUsages() []*Block {
	return oa.CondUsages
}

func (oa *OperandAttr) IsTainted() bool {
	return oa.Tainted
}

func (oa *OperandAttr) IsWritten() bool {
	return len(oa.Ops) > 0
}

func (oa *OperandAttr) String() string {
	return ""
}

type OperSymbolic struct {
	OperandAttr
	Val string
}

func NewOperSymbolic(val string, tainted bool) *OperSymbolic {
	return &OperSymbolic{
		OperandAttr: OperandAttr{
			Assertions: make([]VarAssert, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
			Tainted:    tainted,
		},
		Val: val,
	}
}

func (oper *OperSymbolic) String() string {
	return "SYMBOLIC(" + oper.Val + ")"
}

type OperString struct {
	OperandAttr
	Val string
}

func NewOperString(val string) *OperString {
	return &OperString{
		OperandAttr: OperandAttr{
			Assertions: make([]VarAssert, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
			Tainted:    false,
		},
		Val: val,
	}
}

func (oper *OperString) String() string {
	return "LITERAL(" + oper.Val + ")"
}

type OperNumber struct {
	OperandAttr
	Val float64
}

func NewOperNumber(val float64) *OperNumber {
	return &OperNumber{
		OperandAttr: OperandAttr{
			Assertions: make([]VarAssert, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
			Tainted:    false,
		},
		Val: val,
	}
}

func (oper *OperNumber) String() string {
	return "LITERAL(" + strconv.FormatFloat(oper.Val, 'f', -1, 64) + ")"
}

type OperBool struct {
	OperandAttr
	Val bool
}

func NewOperBool(val bool) *OperBool {
	return &OperBool{
		OperandAttr: OperandAttr{
			Assertions: make([]VarAssert, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
			Tainted:    false,
		},
		Val: val,
	}
}

func (oper *OperBool) String() string {
	if oper.Val {
		return "LITERAL(TRUE)"
	}
	return "LITERAL(FALSE)"
}

type OperObject struct {
	OperandAttr
	ClassName string
}

func NewOperObject(className string) *OperObject {
	return &OperObject{
		OperandAttr: OperandAttr{
			Assertions: make([]VarAssert, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
			Tainted:    false,
		},
		ClassName: className,
	}
}

func (oper *OperObject) String() string {
	return "OBJECT(" + oper.ClassName + ")"
}

type OperVariable struct {
	OperandAttr
	Name  Operand
	Value Operand // if the variable have been defined
}

func NewOperVar(name Operand, value Operand) *OperVariable {
	if value == nil {
		value = NewOperNull()
	}
	return &OperVariable{
		OperandAttr: OperandAttr{
			Assertions: make([]VarAssert, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
			Tainted:    false,
		},
		Name:  name,
		Value: value,
	}
}

func (oper *OperVariable) String() string {
	return "NAME(" + oper.Name.String() + "):(" + oper.Value.String() + ")"
}

type OperBoundVar struct {
	OperandAttr
	Name  Operand
	Value Operand
	ByRef bool
	Scope VarScope
	Extra Operand // TODO: check functionality
}

// Variable immune to SSA
func NewOperBoundVar(name Operand, value Operand, byref bool, scope VarScope, extra Operand) *OperBoundVar {
	return &OperBoundVar{
		OperandAttr: OperandAttr{
			Assertions: make([]VarAssert, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
			Tainted:    false,
		},
		Name:  name,
		Value: value,
		ByRef: byref,
		Scope: scope,
		Extra: extra,
	}
}

func (oper *OperBoundVar) String() string {
	return "NAME(" + oper.Name.String() + ")"
}

type OperTemporary struct {
	OperandAttr
	Original Operand
}

func NewOperTemporary(original Operand) *OperTemporary {
	return &OperTemporary{
		OperandAttr: OperandAttr{
			Assertions: make([]VarAssert, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
			Tainted:    false,
		},
		Original: original,
	}
}

func (oper *OperTemporary) String() string {
	return "Var(" + oper.Original.String() + ")"
}

type OperNull struct {
	OperandAttr
}

func NewOperNull() *OperNull {
	return &OperNull{
		OperandAttr: OperandAttr{
			Assertions: make([]VarAssert, 0),
			Ops:        make([]Op, 0),
			Usages:     make([]Op, 0),
			Tainted:    false,
		},
	}
}

func (oper *OperNull) String() string {
	return "NULL"
}
