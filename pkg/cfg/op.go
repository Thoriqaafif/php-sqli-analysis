package cfg

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/VKCOM/php-parser/pkg/position"
)

type Op interface {
	// AddReadRef(o Operand) Operand
	// AddWriteRef(o Operand) Operand
	GetType() string
	GetPosition() *position.Position
	SetFilePath(filePath string)
	GetFilePath() string
	GetOpVars() map[string]Operand
	GetOpListVars() map[string][]Operand
	ChangeOpVar(vrName string, vr Operand)
	ChangeOpListVar(vrName string, vr []Operand)
	GetOpVarPos(name string) *position.Position
	GetOpVarListPos(name string, index int) *position.Position
	SetBlock(*Block)
	GetBlock() *Block
	Clone() Op
}

type OpCallable interface {
	GetFunc() *OpFunc
}

type OpGeneral struct {
	Position *position.Position
	FilePath string
	OpBlock  *Block
}

func AddReadRefs(op Op, opers ...Operand) []Operand {
	result := make([]Operand, 0)
	for _, oper := range opers {
		if oper != nil {
			result = append(result, AddReadRef(op, oper))
		}
	}
	return result
}

func AddReadRef(op Op, oper Operand) Operand {
	if oper == nil {
		return nil
	}

	oper.AddUsage(op)
	return oper
}

func AddWriteRefs(op Op, opers ...Operand) []Operand {
	result := make([]Operand, 0)
	for _, oper := range opers {
		if oper != nil {
			result = append(result, AddWriteRef(op, oper))
		}
	}
	return result
}

func AddWriteRef(op Op, oper Operand) Operand {
	if oper == nil {
		return nil
	}

	oper.AddWriteOp(op)
	return oper
}

func GetSubBlocks(op Op) map[string]*Block {
	m := make(map[string]*Block)
	switch o := op.(type) {
	case *OpExprParam:
		if o.DefaultBlock != nil {
			m["DefaultBlock"] = o.DefaultBlock
		}
	case *OpStmtInterface:
		if o.Stmts != nil {
			m["Stmts"] = o.Stmts
		}
	case *OpStmtClass:
		if o.Stmts != nil {
			m["Stmts"] = o.Stmts
		} else {
			log.Fatal("nil stmts")
		}
	case *OpStmtTrait:
		if o.Stmts != nil {
			m["Stmts"] = o.Stmts
		}
	case *OpStmtJump:
		if o.Target != nil {
			m["Target"] = o.Target
		}
	case *OpStmtJumpIf:
		if o.If != nil {
			m["If"] = o.If
		}
		if o.Else != nil {
			m["Else"] = o.Else
		}
	case *OpStmtProperty:
		if o.DefaultBlock != nil {
			m["DefaultBlock"] = o.DefaultBlock
		}
	case *OpStmtSwitch:
		for i, subBlock := range o.Targets {
			s := fmt.Sprintf("Target[%d]", i)
			m[s] = subBlock
		}
	case *OpConst:
		if o.ValueBlock != nil {
			m["ValueBlock"] = o.ValueBlock
		}
	case *OpStaticVar:
		if o.DefaultBlock != nil {
			m["DefaultBlock"] = o.DefaultBlock
		}
	}

	return m
}

func ChangeSubBlock(op Op, subBlockName string, newBlock *Block) {
	switch o := op.(type) {
	case *OpExprParam:
		if subBlockName == "DefaultBlock" {
			o.DefaultBlock = newBlock
		} else {
			log.Fatalf("Error: Unknown OpExprParam subblock '%s'", subBlockName)
		}
	case *OpStmtInterface:
		if subBlockName == "Stmts" {
			o.Stmts = newBlock
		} else {
			log.Fatalf("Error: Unknown OpStmtInterface subblock '%s'", subBlockName)
		}
	case *OpStmtClass:
		if subBlockName == "Stmts" {
			o.Stmts = newBlock
		} else {
			log.Fatalf("Error: Unknown OpStmtClass subblock '%s'", subBlockName)
		}
	case *OpStmtTrait:
		if subBlockName == "Stmts" {
			o.Stmts = newBlock
		} else {
			log.Fatalf("Error: Unknown OpStmtTrait subblock '%s'", subBlockName)
		}
	case *OpStmtJump:
		if subBlockName == "Target" {
			o.Target = newBlock
		} else {
			log.Fatalf("Error: Unknown OpStmtJump subblock '%s'", subBlockName)
		}
	case *OpStmtJumpIf:
		if subBlockName == "If" {
			o.If = newBlock
		} else if subBlockName == "Else" {
			o.Else = newBlock
		} else {
			log.Fatalf("Error: Unknown OpStmtJumpIf subblock '%s'", subBlockName)
		}
	case *OpStmtProperty:
		if subBlockName == "DefaultBlock" {
			o.DefaultBlock = newBlock
		} else {
			log.Fatalf("Error: Unknown OpStmtProperty subblock '%s'", subBlockName)
		}
	case *OpStmtSwitch:
		startIdx := strings.Index(subBlockName, "[")
		endIdx := strings.Index(subBlockName, "]")
		if startIdx == -1 || endIdx == -1 {
			log.Fatalf("Error: Unknown OpStmtSwitch subblock '%s'", subBlockName)
		}
		idx, err := strconv.Atoi(subBlockName[startIdx+1 : endIdx])
		if err != nil || idx >= len(o.Targets) {
			log.Fatalf("Error: Unknown OpStmtSwitch subblock '%s'", subBlockName)
		}
		o.Targets[idx] = newBlock
	case *OpConst:
		if subBlockName == "ValueBlock" {
			o.ValueBlock = newBlock
		} else {
			log.Fatalf("Error: Unknown OpConst subblock '%s'", subBlockName)
		}
	case *OpStaticVar:
		if subBlockName == "DefaultBlock" {
			o.DefaultBlock = newBlock
		} else {
			log.Fatalf("Error: Unknown OpStaticVar subblock '%s'", subBlockName)
		}
	}
}

func IsWriteVar(op Op, varName string) bool {
	if varName == "Result" {
		return true
	} else if varName == "Var" {
		switch op.(type) {
		case *OpStaticVar:
			return true
		case *OpExprAssign:
			return true
		case *OpExprAssignRef:
			return true
		}
	}
	return false
}

// func (op *OpGeneral) GetType() string {
// 	return "aaa"
// }

func (op *OpGeneral) GetPosition() *position.Position {
	return op.Position
}

func (op *OpGeneral) SetFilePath(filePath string) {
	op.FilePath = filePath
}

func (op *OpGeneral) GetFilePath() string {
	return op.FilePath
}

func (op *OpGeneral) GetOpVars() map[string]Operand {
	return map[string]Operand{}
}

func (op *OpGeneral) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{}
}

func (op *OpGeneral) ChangeOpVar(vrName string, vr Operand) {
}

func (op *OpGeneral) ChangeOpListVar(vrName string, vr []Operand) {
}

func (op *OpGeneral) GetOpVarPos(vrName string) *position.Position {
	return nil
}

func (op *OpGeneral) GetOpVarListPos(vrName string, index int) *position.Position {
	return nil
}

func (op *OpGeneral) SetBlock(block *Block) {
	op.OpBlock = block
}

func (op *OpGeneral) GetBlock() *Block {
	return op.OpBlock
}

type FuncModifFlag int

const (
	FUNC_MODIF_PUBLIC      FuncModifFlag = 1
	FUNC_MODIF_PROTECTED   FuncModifFlag = 1 << 1
	FUNC_MODIF_PRIVATE     FuncModifFlag = 1 << 2
	FUNC_MODIF_STATIC      FuncModifFlag = 1 << 3
	FUNC_MODIF_ABSTRACT    FuncModifFlag = 1 << 4
	FUNC_MODIF_FINAL       FuncModifFlag = 1 << 5
	FUNC_MODIF_RETURNS_REF FuncModifFlag = 1 << 6
	FUNC_MODIF_CLOSURE     FuncModifFlag = 1 << 7
)

type OpFunc struct {
	OpGeneral
	Name       string
	Flags      FuncModifFlag
	ReturnType OpType
	Class      *OperString
	Params     []*OpExprParam
	Cfg        *Block
	CallableOp Op

	// helper
	ContaintTainted bool
	Sources         []Op
	Calls           []Op
}

func NewClassFunc(name string, flag FuncModifFlag, returnType OpType, class OperString, entryBlock *Block, position *position.Position) (*OpFunc, error) {
	if entryBlock == nil {
		return nil, errors.New("entry block cannot be nil")
	}
	f := &OpFunc{
		OpGeneral: OpGeneral{
			Position: position,
		},
		Name:       name,
		Flags:      flag,
		ReturnType: returnType,
		Class:      &class,
		Params:     make([]*OpExprParam, 0),
		Cfg:        entryBlock,
	}
	return f, nil
}

func NewFunc(name string, flags FuncModifFlag, returnType OpType, entryBlock *Block, position *position.Position) (*OpFunc, error) {
	if entryBlock == nil {
		return nil, errors.New("entry block cannot be nil")
	}
	return &OpFunc{
		OpGeneral: OpGeneral{
			Position: position,
		},
		Name:       name,
		Flags:      flags,
		ReturnType: returnType,
		Class:      nil,
		Params:     make([]*OpExprParam, 0),
		Cfg:        entryBlock,
	}, nil
}

func (op *OpFunc) GetScopedName() string {
	if op.Class != nil {
		className := op.Class.Val
		return className + "::" + op.Name
	}
	return op.Name
}

func (op *OpFunc) AddModifier(flag FuncModifFlag) {
	op.Flags |= flag
}

func (op *OpFunc) GetVisibility() FuncModifFlag {
	return FuncModifFlag(op.Flags & 7)
}

func (op *OpFunc) IsPublic() bool {
	return op.Flags&FUNC_MODIF_PUBLIC != 0
}

func (op *OpFunc) IsPrivate() bool {
	return op.Flags&FUNC_MODIF_PRIVATE != 0
}

func (op *OpFunc) IsProtected() bool {
	return op.Flags&FUNC_MODIF_PROTECTED != 0
}

func (op *OpFunc) IsStatic() bool {
	return op.Flags&FUNC_MODIF_STATIC != 0
}

func (op *OpFunc) IsAbstract() bool {
	return op.Flags&FUNC_MODIF_ABSTRACT != 0
}

func (op *OpFunc) IsFinal() bool {
	return op.Flags&FUNC_MODIF_FINAL != 0
}

func (op *OpFunc) IsReturnRef() bool {
	return op.Flags&FUNC_MODIF_RETURNS_REF != 0
}

func (op *OpFunc) IsClosure() bool {
	return op.Flags&FUNC_MODIF_CLOSURE != 0
}

func (op *OpFunc) GetType() string {
	return "Func"
}

func (op *OpFunc) Clone() Op {
	params := make([]*OpExprParam, len(op.Params))
	copy(params, op.Params)
	return &OpFunc{
		OpGeneral:       op.OpGeneral,
		Name:            op.Name,
		Flags:           op.Flags,
		ReturnType:      op.ReturnType,
		Class:           op.Class,
		Params:          params,
		Cfg:             op.Cfg,
		CallableOp:      op.CallableOp,
		ContaintTainted: op.ContaintTainted,
	}
}

type OpPhi struct {
	OpGeneral
	Vars   map[Operand]struct{}
	Block  *Block
	Result Operand
}

func NewOpPhi(result Operand, block *Block, position *position.Position) *OpPhi {
	op := &OpPhi{
		OpGeneral: OpGeneral{
			Position: position,
		},
		Vars:   make(map[Operand]struct{}),
		Block:  block,
		Result: result,
	}

	AddWriteRef(op, result)

	return op
}

// add an operand to phi vars, if not exist
func (op *OpPhi) AddOperand(oper Operand) {
	var empty struct{}
	// add if operand have not been in vars and not phi itself
	if _, ok := op.Vars[oper]; !ok && op.Result != oper {
		tmp := AddReadRef(op, oper)
		op.Vars[tmp] = empty
	}
}

// remove an operand from phi vars
func (op *OpPhi) RemoveOperand(oper Operand) {
	if _, ok := op.Vars[oper]; ok {
		oper.RemoveUsage(op)
		delete(op.Vars, oper)
	}
}

func (op *OpPhi) GetVars() []Operand {
	vars := make([]Operand, 0, len(op.Vars))
	for vr := range op.Vars {
		vars = append(vars, vr)
	}
	return vars
}

func (op *OpPhi) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Result": op.Result,
	}
}

func (op *OpPhi) GetType() string {
	return "Phi"
}

func (op *OpPhi) HasOperand(oper Operand) bool {
	if _, ok := op.Vars[oper]; ok {
		return true
	}
	return false
}

func (op *OpPhi) Clone() Op {
	return &OpPhi{
		OpGeneral: op.OpGeneral,
		Vars:      op.Vars,
		Block:     op.Block,
		Result:    op.Result,
	}
}

type OpAttributeGroup struct {
	OpGeneral
	Attrs []*OpAttribute
}

func NewOpAttributeGroup(attrs []*OpAttribute, pos *position.Position) *OpAttributeGroup {
	Op := &OpAttributeGroup{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Attrs: attrs,
	}

	return Op
}

func (op *OpAttributeGroup) GetType() string {
	return "AttributeGroup"
}

func (op *OpAttributeGroup) Clone() Op {
	attrs := make([]*OpAttribute, len(op.Attrs))
	copy(attrs, op.Attrs)
	return &OpAttributeGroup{
		OpGeneral: op.OpGeneral,
		Attrs:     attrs,
	}
}

type OpAttribute struct {
	OpGeneral
	Name Operand
	Args []Operand
}

func NewOpAttribute(name Operand, args []Operand, pos *position.Position) *OpAttribute {
	Op := &OpAttribute{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name: name,
		Args: args,
	}

	AddReadRef(Op, name)

	return Op
}

func (op *OpAttribute) GetType() string {
	return "Attribute"
}

func (op *OpAttribute) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name": op.Name,
	}
}

func (op *OpAttribute) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	}
}

func (op *OpAttribute) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Args": op.Args,
	}
}

func (op *OpAttribute) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Args":
		op.Args = vr
	}
}

func (op *OpAttribute) Clone() Op {
	args := make([]Operand, len(op.Args))
	copy(args, op.Args)
	return &OpAttribute{
		OpGeneral: op.OpGeneral,
		Name:      op.Name,
		Args:      args,
	}
}

type OpExprAssertion struct {
	OpGeneral
	Expr      Operand
	Assertion Assertion
	Result    Operand
}

func NewOpExprAssertion(read, write Operand, assertion Assertion, pos *position.Position) *OpExprAssertion {
	Op := &OpExprAssertion{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:      read,
		Assertion: assertion,
		Result:    write,
	}

	AddReadRef(Op, read)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprAssertion) GetType() string {
	return "ExprAssertion"
}

func (op *OpExprAssertion) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprAssertion) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprAssertion) Clone() Op {
	return &OpExprAssertion{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Assertion: op.Assertion,
		Result:    op.Result,
	}
}

type OpExprArray struct {
	OpGeneral
	Keys   []Operand
	Vals   []Operand
	ByRef  []bool
	Result Operand
}

func NewOpExprArray(keys, vals []Operand, byRef []bool, pos *position.Position) *OpExprArray {
	Op := &OpExprArray{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Keys:   keys,
		Vals:   vals,
		ByRef:  byRef,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, keys...)
	AddReadRefs(Op, vals...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprArray) GetType() string {
	return "ExprArray"
}

func (op *OpExprArray) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Result": op.Result,
	}
}

func (op *OpExprArray) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprArray) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Keys":   op.Keys,
		"Values": op.Vals,
	}
}

func (op *OpExprArray) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Keys":
		op.Keys = vr
	case "Values":
		op.Vals = vr
	}
}

func (op *OpExprArray) Clone() Op {
	keys := make([]Operand, len(op.Keys))
	vals := make([]Operand, len(op.Vals))
	copy(keys, op.Keys)
	copy(vals, op.Vals)
	return &OpExprArray{
		OpGeneral: op.OpGeneral,
		Keys:      keys,
		Vals:      vals,
		ByRef:     op.ByRef,
		Result:    op.Result,
	}
}

type OpExprArrayDimFetch struct {
	OpGeneral
	Var    Operand
	Dim    Operand
	Result Operand
}

func NewOpExprArrayDimFetch(vr, dim Operand, pos *position.Position) *OpExprArrayDimFetch {
	Op := &OpExprArrayDimFetch{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var:    vr,
		Dim:    dim,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, vr, dim)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprArrayDimFetch) GetType() string {
	return "ExprArrayDimFetch"
}

func (op *OpExprArrayDimFetch) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Dim":    op.Dim,
		"Result": op.Result,
	}
}

func (op *OpExprArrayDimFetch) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Dim":
		op.Dim = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprArrayDimFetch) Clone() Op {
	return &OpExprArrayDimFetch{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
		Dim:       op.Dim,
		Result:    op.Result,
	}
}

type OpExprAssign struct {
	OpGeneral
	Var    Operand
	Expr   Operand
	Result Operand

	VarPos  *position.Position
	ExprPos *position.Position
}

func NewOpExprAssign(vr, expr Operand, vrPos, exprPos, pos *position.Position) *OpExprAssign {
	Op := &OpExprAssign{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var:     vr,
		Expr:    expr,
		Result:  NewOperTemporary(nil),
		VarPos:  vrPos,
		ExprPos: exprPos,
	}

	AddReadRef(Op, expr)
	AddWriteRefs(Op, Op.Result, vr)

	return Op
}

func (op *OpExprAssign) GetType() string {
	return "ExprAssign"
}

func (op *OpExprAssign) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprAssign) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprAssign) GetOpVarPos(vrName string) *position.Position {
	switch vrName {
	case "Var":
		return op.VarPos
	case "Expr":
		return op.ExprPos
	}
	return nil
}

func (op *OpExprAssign) Clone() Op {
	return &OpExprAssign{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Var:       op.Var,
		Result:    op.Result,
	}
}

type OpExprAssignRef struct {
	OpGeneral
	Var    Operand
	Expr   Operand
	Result Operand
}

func NewOpExprAssignRef(vr, expr Operand, pos *position.Position) *OpExprAssignRef {
	Op := &OpExprAssignRef{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var:    vr,
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRefs(Op, Op.Result, vr)

	return Op
}

func (op *OpExprAssignRef) GetType() string {
	return "ExprAssignRef"
}

func (op *OpExprAssignRef) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprAssignRef) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprAssignRef) Clone() Op {
	return &OpExprAssignRef{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Var:       op.Var,
		Result:    op.Result,
	}
}

type OpExprBinaryBitwiseAnd struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryBitwiseAnd(left, right Operand, pos *position.Position) *OpExprBinaryBitwiseAnd {
	Op := &OpExprBinaryBitwiseAnd{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryBitwiseAnd) GetType() string {
	return "ExprBinaryBitwiseAnd"
}

func (op *OpExprBinaryBitwiseAnd) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryBitwiseAnd) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryBitwiseAnd) Clone() Op {
	return &OpExprBinaryBitwiseAnd{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryBitwiseOr struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryBitwiseOr(left, right Operand, pos *position.Position) *OpExprBinaryBitwiseOr {
	Op := &OpExprBinaryBitwiseOr{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryBitwiseOr) GetType() string {
	return "ExprBinaryBitwiseOr"
}

func (op *OpExprBinaryBitwiseOr) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryBitwiseOr) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryBitwiseOr) Clone() Op {
	return &OpExprBinaryBitwiseOr{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryBitwiseXor struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryBitwiseXor(left, right Operand, pos *position.Position) *OpExprBinaryBitwiseXor {
	Op := &OpExprBinaryBitwiseXor{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryBitwiseXor) GetType() string {
	return "ExprBinaryBitwiseXor"
}

func (op *OpExprBinaryBitwiseXor) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryBitwiseXor) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryBitwiseXor) Clone() Op {
	return &OpExprBinaryBitwiseXor{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryCoalesce struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryCoalesce(left, right Operand, pos *position.Position) *OpExprBinaryCoalesce {
	Op := &OpExprBinaryCoalesce{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryCoalesce) GetType() string {
	return "ExprBinaryCoalesce"
}

func (op *OpExprBinaryCoalesce) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryCoalesce) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryCoalesce) Clone() Op {
	return &OpExprBinaryCoalesce{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryConcat struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand

	LeftPos  *position.Position
	RightPos *position.Position
}

func NewOpExprBinaryConcat(left, right Operand, leftPos, rightPos, pos *position.Position) *OpExprBinaryConcat {
	Op := &OpExprBinaryConcat{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:     left,
		Right:    right,
		Result:   NewOperTemporary(nil),
		LeftPos:  leftPos,
		RightPos: rightPos,
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryConcat) GetType() string {
	return "ExprBinaryConcat"
}

func (op *OpExprBinaryConcat) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryConcat) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryConcat) GetOpVarPos(vrName string) *position.Position {
	switch vrName {
	case "Left":
		return op.LeftPos
	case "Right":
		return op.RightPos
	}

	return nil
}

func (op *OpExprBinaryConcat) Clone() Op {
	return &OpExprBinaryConcat{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryDiv struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryDiv(left, right Operand, pos *position.Position) *OpExprBinaryDiv {
	Op := &OpExprBinaryDiv{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryDiv) GetType() string {
	return "ExprBinaryDiv"
}

func (op *OpExprBinaryDiv) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryDiv) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryDiv) Clone() Op {
	return &OpExprBinaryDiv{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryEqual struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryEqual(left, right Operand, pos *position.Position) *OpExprBinaryEqual {
	Op := &OpExprBinaryEqual{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryEqual) GetType() string {
	return "ExprBinaryEqual"
}

func (op *OpExprBinaryEqual) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryEqual) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryEqual) Clone() Op {
	return &OpExprBinaryEqual{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryGreater struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryGreater(left, right Operand, pos *position.Position) *OpExprBinaryGreater {
	Op := &OpExprBinaryGreater{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryGreater) GetType() string {
	return "ExprBinaryGreater"
}

func (op *OpExprBinaryGreater) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryGreater) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryGreater) Clone() Op {
	return &OpExprBinaryGreater{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryGreaterOrEqual struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryGreaterOrEqual(left, right Operand, pos *position.Position) *OpExprBinaryGreaterOrEqual {
	Op := &OpExprBinaryGreaterOrEqual{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryGreaterOrEqual) GetType() string {
	return "ExprBinaryGreaterOrEqual"
}

func (op *OpExprBinaryGreaterOrEqual) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryGreaterOrEqual) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryGreaterOrEqual) Clone() Op {
	return &OpExprBinaryGreaterOrEqual{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryIdentical struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryIdentical(left, right Operand, pos *position.Position) *OpExprBinaryIdentical {
	Op := &OpExprBinaryIdentical{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryIdentical) GetType() string {
	return "ExprBinaryIdentical"
}

func (op *OpExprBinaryIdentical) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryIdentical) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryIdentical) Clone() Op {
	return &OpExprBinaryIdentical{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryLogicalAnd struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryLogicalAnd(left, right Operand, pos *position.Position) *OpExprBinaryLogicalAnd {
	Op := &OpExprBinaryLogicalAnd{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryLogicalAnd) GetType() string {
	return "ExprBinaryLogicalAnd"
}

func (op *OpExprBinaryLogicalAnd) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryLogicalAnd) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryLogicalAnd) Clone() Op {
	return &OpExprBinaryLogicalAnd{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryLogicalOr struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryLogicalOr(left, right Operand, pos *position.Position) *OpExprBinaryLogicalOr {
	Op := &OpExprBinaryLogicalOr{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryLogicalOr) GetType() string {
	return "ExprBinaryLogicalOr"
}

func (op *OpExprBinaryLogicalOr) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryLogicalOr) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryLogicalOr) Clone() Op {
	return &OpExprBinaryLogicalOr{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryLogicalXor struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryLogicalXor(left, right Operand, pos *position.Position) *OpExprBinaryLogicalXor {
	Op := &OpExprBinaryLogicalXor{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryLogicalXor) GetType() string {
	return "ExprBinaryLogicalXor"
}

func (op *OpExprBinaryLogicalXor) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryLogicalXor) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryLogicalXor) Clone() Op {
	return &OpExprBinaryLogicalXor{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryMinus struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryMinus(left, right Operand, pos *position.Position) *OpExprBinaryMinus {
	Op := &OpExprBinaryMinus{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryMinus) GetType() string {
	return "ExprBinaryMinus"
}

func (op *OpExprBinaryMinus) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryMinus) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryMinus) Clone() Op {
	return &OpExprBinaryMinus{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryMod struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryMod(left, right Operand, pos *position.Position) *OpExprBinaryMod {
	Op := &OpExprBinaryMod{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryMod) GetType() string {
	return "ExprBinaryMod"
}

func (op *OpExprBinaryMod) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryMod) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryMod) Clone() Op {
	return &OpExprBinaryMod{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryMul struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryMul(left, right Operand, pos *position.Position) *OpExprBinaryMul {
	Op := &OpExprBinaryMul{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryMul) GetType() string {
	return "ExprBinaryMul"
}

func (op *OpExprBinaryMul) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryMul) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryMul) Clone() Op {
	return &OpExprBinaryMul{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryNotEqual struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryNotEqual(left, right Operand, pos *position.Position) *OpExprBinaryNotEqual {
	Op := &OpExprBinaryNotEqual{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryNotEqual) GetType() string {
	return "ExprBinaryNotEqual"
}

func (op *OpExprBinaryNotEqual) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryNotEqual) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryNotEqual) Clone() Op {
	return &OpExprBinaryNotEqual{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryNotIdentical struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryNotIdentical(left, right Operand, pos *position.Position) *OpExprBinaryNotIdentical {
	Op := &OpExprBinaryNotIdentical{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryNotIdentical) GetType() string {
	return "ExprBinaryNotIdentical"
}

func (op *OpExprBinaryNotIdentical) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryNotIdentical) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryNotIdentical) Clone() Op {
	return &OpExprBinaryNotIdentical{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryPlus struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryPlus(left, right Operand, pos *position.Position) *OpExprBinaryPlus {
	Op := &OpExprBinaryPlus{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryPlus) GetType() string {
	return "ExprBinaryPlus"
}

func (op *OpExprBinaryPlus) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryPlus) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryPlus) Clone() Op {
	return &OpExprBinaryPlus{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryPow struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryPow(left, right Operand, pos *position.Position) *OpExprBinaryPow {
	Op := &OpExprBinaryPow{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryPow) GetType() string {
	return "ExprBinaryPow"
}

func (op *OpExprBinaryPow) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryPow) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryPow) Clone() Op {
	return &OpExprBinaryPow{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryShiftLeft struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryShiftLeft(left, right Operand, pos *position.Position) *OpExprBinaryShiftLeft {
	Op := &OpExprBinaryShiftLeft{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryShiftLeft) GetType() string {
	return "ExorBinaryShiftLeft"
}

func (op *OpExprBinaryShiftLeft) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryShiftLeft) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryShiftLeft) Clone() Op {
	return &OpExprBinaryShiftLeft{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinaryShiftRight struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinaryShiftRight(left, right Operand, pos *position.Position) *OpExprBinaryShiftRight {
	Op := &OpExprBinaryShiftRight{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinaryShiftRight) GetType() string {
	return "ExprBinaryShiftRight"
}

func (op *OpExprBinaryShiftRight) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinaryShiftRight) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinaryShiftRight) Clone() Op {
	return &OpExprBinaryShiftRight{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinarySmaller struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinarySmaller(left, right Operand, pos *position.Position) *OpExprBinarySmaller {
	Op := &OpExprBinarySmaller{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinarySmaller) GetType() string {
	return "ExprBinarySmaller"
}

func (op *OpExprBinarySmaller) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinarySmaller) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinarySmaller) Clone() Op {
	return &OpExprBinarySmaller{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinarySmallerOrEqual struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinarySmallerOrEqual(left, right Operand, pos *position.Position) *OpExprBinarySmallerOrEqual {
	Op := &OpExprBinarySmallerOrEqual{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinarySmallerOrEqual) GetType() string {
	return "ExprBinarySmallerOrEqual"
}

func (op *OpExprBinarySmallerOrEqual) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinarySmallerOrEqual) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinarySmallerOrEqual) Clone() Op {
	return &OpExprBinarySmallerOrEqual{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBinarySpaceship struct {
	OpGeneral
	Left   Operand
	Right  Operand
	Result Operand
}

func NewOpExprBinarySpaceship(left, right Operand, pos *position.Position) *OpExprBinarySpaceship {
	Op := &OpExprBinarySpaceship{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Left:   left,
		Right:  right,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, left, right)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBinarySpaceship) GetType() string {
	return "ExprBinarySpaceship"
}

func (op *OpExprBinarySpaceship) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Left":   op.Left,
		"Right":  op.Right,
		"Result": op.Result,
	}
}

func (op *OpExprBinarySpaceship) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Left":
		op.Left = vr
	case "Right":
		op.Right = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBinarySpaceship) Clone() Op {
	return &OpExprBinarySpaceship{
		OpGeneral: op.OpGeneral,
		Left:      op.Left,
		Right:     op.Right,
		Result:    op.Result,
	}
}

type OpExprBitwiseNot struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprBitwiseNot(expr Operand, pos *position.Position) *OpExprBitwiseNot {
	Op := &OpExprBitwiseNot{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBitwiseNot) GetType() string {
	return "ExprBitwiseNot"
}

func (op *OpExprBitwiseNot) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprBitwiseNot) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBitwiseNot) Clone() Op {
	return &OpExprBitwiseNot{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprBooleanNot struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprBooleanNot(expr Operand, pos *position.Position) *OpExprBooleanNot {
	Op := &OpExprBooleanNot{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprBooleanNot) GetType() string {
	return "ExprBooleanNot"
}

func (op *OpExprBooleanNot) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprBooleanNot) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprBooleanNot) Clone() Op {
	return &OpExprBooleanNot{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprCastArray struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprCastArray(expr Operand, pos *position.Position) *OpExprCastArray {
	Op := &OpExprCastArray{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprCastArray) GetType() string {
	return "ExprCastArray"
}

func (op *OpExprCastArray) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprCastArray) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprCastArray) Clone() Op {
	return &OpExprCastArray{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprCastBool struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprCastBool(expr Operand, pos *position.Position) *OpExprCastBool {
	Op := &OpExprCastBool{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprCastBool) GetType() string {
	return "ExprCastBool"
}

func (op *OpExprCastBool) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprCastBool) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprCastBool) Clone() Op {
	return &OpExprCastBool{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprCastDouble struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprCastDouble(expr Operand, pos *position.Position) *OpExprCastDouble {
	Op := &OpExprCastDouble{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprCastDouble) GetType() string {
	return "ExprCastDouble"
}

func (op *OpExprCastDouble) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprCastDouble) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprCastDouble) Clone() Op {
	return &OpExprCastDouble{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprCastInt struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprCastInt(expr Operand, pos *position.Position) *OpExprCastInt {
	Op := &OpExprCastInt{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprCastInt) GetType() string {
	return "ExprCastInt"
}

func (op *OpExprCastInt) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprCastInt) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprCastInt) Clone() Op {
	return &OpExprCastInt{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprCastObject struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprCastObject(expr Operand, pos *position.Position) *OpExprCastObject {
	Op := &OpExprCastObject{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprCastObject) GetType() string {
	return "ExprCastObject"
}

func (op *OpExprCastObject) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprCastObject) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprCastObject) Clone() Op {
	return &OpExprCastObject{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprCastString struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprCastString(expr Operand, pos *position.Position) *OpExprCastString {
	Op := &OpExprCastString{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprCastString) GetType() string {
	return "ExprCastString"
}

func (op *OpExprCastString) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprCastString) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprCastString) Clone() Op {
	return &OpExprCastString{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprCastUnset struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprCastUnset(expr Operand, pos *position.Position) *OpExprCastUnset {
	Op := &OpExprCastUnset{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprCastUnset) GetType() string {
	return "ExprCastUnset"
}

func (op *OpExprCastUnset) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprCastUnset) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprCastUnset) Clone() Op {
	return &OpExprCastUnset{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprClassConstFetch struct {
	OpGeneral
	Class  Operand
	Name   Operand
	Result Operand
}

func NewOpExprClassConstFetch(class, name Operand, pos *position.Position) *OpExprClassConstFetch {
	Op := &OpExprClassConstFetch{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Class:  class,
		Name:   name,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, class, name)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprClassConstFetch) GetType() string {
	return "ExprClassConstFetch"
}

func (op *OpExprClassConstFetch) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Class":  op.Class,
		"Name":   op.Name,
		"Result": op.Result,
	}
}

func (op *OpExprClassConstFetch) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Class":
		op.Class = vr
	case "Name":
		op.Name = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprClassConstFetch) Clone() Op {
	return &OpExprClassConstFetch{
		OpGeneral: op.OpGeneral,
		Class:     op.Class,
		Name:      op.Name,
		Result:    op.Result,
	}
}

type OpExprClone struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprClone(expr Operand, pos *position.Position) *OpExprClone {
	Op := &OpExprClone{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprClone) GetType() string {
	return "ExprClone"
}

func (op *OpExprClone) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprClone) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprClone) Clone() Op {
	return &OpExprClone{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprClosure struct {
	OpGeneral
	Func    *OpFunc
	UseVars []Operand
	Result  Operand
}

func NewOpExprClosure(Func *OpFunc, useVars []Operand, pos *position.Position) *OpExprClosure {
	Op := &OpExprClosure{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Func:    Func,
		UseVars: useVars,
		Result:  NewOperTemporary(nil),
	}

	AddReadRefs(Op, useVars...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprClosure) GetType() string {
	return "ExprClosure"
}

func (op *OpExprClosure) GetFunc() *OpFunc {
	return op.Func
}

func (op *OpExprClosure) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Result": op.Result,
	}
}

func (op *OpExprClosure) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprClosure) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"UseVars": op.UseVars,
	}
}

func (op *OpExprClosure) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "UseVars":
		op.UseVars = vr
	}
}

func (op *OpExprClosure) Clone() Op {
	useVars := make([]Operand, len(op.UseVars))
	copy(useVars, op.UseVars)
	return &OpExprClosure{
		OpGeneral: op.OpGeneral,
		Func:      op.Func,
		UseVars:   useVars,
		Result:    op.Result,
	}
}

type OpExprConcatList struct {
	OpGeneral
	List   []Operand
	Result Operand

	ListPos []*position.Position
}

func NewOpExprConcatList(list []Operand, listPos []*position.Position, pos *position.Position) *OpExprConcatList {
	Op := &OpExprConcatList{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		List:    list,
		Result:  NewOperTemporary(nil),
		ListPos: listPos,
	}

	AddReadRefs(Op, list...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprConcatList) GetType() string {
	return "ExprConcatList"
}

func (op *OpExprConcatList) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Result": op.Result,
	}
}

func (op *OpExprConcatList) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprConcatList) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"List": op.List,
	}
}

func (op *OpExprConcatList) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "List":
		op.List = vr
	}
}

func (op *OpExprConcatList) GetOpVarListPos(vrName string, index int) *position.Position {
	switch vrName {
	case "List":
		return op.ListPos[index]
	}
	return nil
}

func (op *OpExprConcatList) Clone() Op {
	list := make([]Operand, len(op.List))
	copy(list, op.List)
	return &OpExprConcatList{
		OpGeneral: op.OpGeneral,
		List:      list,
		Result:    op.Result,
	}
}

type OpExprConstFetch struct {
	OpGeneral
	Name   Operand
	Result Operand
}

func NewOpExprConstFetch(name Operand, pos *position.Position) *OpExprConstFetch {
	Op := &OpExprConstFetch{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name:   name,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, name)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprConstFetch) GetType() string {
	return "ExprConstFetch"
}

func (op *OpExprConstFetch) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name":   op.Name,
		"Result": op.Result,
	}
}

func (op *OpExprConstFetch) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprConstFetch) Clone() Op {
	return &OpExprConstFetch{
		OpGeneral: op.OpGeneral,
		Name:      op.Name,
		Result:    op.Result,
	}
}

type OpExprEmpty struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprEmpty(expr Operand, pos *position.Position) *OpExprEmpty {
	Op := &OpExprEmpty{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprEmpty) GetType() string {
	return "ExprEmpty"
}

func (op *OpExprEmpty) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprEmpty) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprEmpty) Clone() Op {
	return &OpExprEmpty{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprEval struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprEval(expr Operand, pos *position.Position) *OpExprEval {
	Op := &OpExprEval{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprEval) GetType() string {
	return "ExprEval"
}

func (op *OpExprEval) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprEval) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprEval) Clone() Op {
	return &OpExprEval{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprFunctionCall struct {
	OpGeneral
	Name   Operand
	Args   []Operand
	Result Operand

	CalledFunc *OpFunc
	NamePos    *position.Position
	ArgsPos    []*position.Position
}

func NewOpExprFunctionCall(name Operand, args []Operand, namePos *position.Position, argsPos []*position.Position, pos *position.Position) *OpExprFunctionCall {
	Op := &OpExprFunctionCall{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name:    name,
		Args:    args,
		Result:  NewOperTemporary(nil),
		NamePos: namePos,
		ArgsPos: argsPos,
	}

	AddReadRef(Op, name)
	AddReadRefs(Op, args...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprFunctionCall) GetType() string {
	return "ExprFunctionCall"
}

func (op *OpExprFunctionCall) GetName() string {
	funcName, _ := GetOperName(op.Name)
	// if err != nil {
	// 	log.Fatalf("Error in GetFuncCallName: %v", err)
	// }
	return funcName
}

func (op *OpExprFunctionCall) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name":   op.Name,
		"Result": op.Result,
	}
}

func (op *OpExprFunctionCall) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprFunctionCall) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Args": op.Args,
	}
}

func (op *OpExprFunctionCall) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Args":
		op.Args = vr
	}
}

func (op *OpExprFunctionCall) GetOpVarPos(vrName string) *position.Position {
	switch vrName {
	case "Name":
		return op.NamePos
	}
	return nil
}

func (op *OpExprFunctionCall) GetOpVarListPos(vrName string, index int) *position.Position {
	switch vrName {
	case "Args":
		return op.ArgsPos[index]
	}
	return nil
}

func (op *OpExprFunctionCall) Clone() Op {
	args := make([]Operand, len(op.Args))
	copy(args, op.Args)
	return &OpExprFunctionCall{
		OpGeneral:  op.OpGeneral,
		Name:       op.Name,
		Args:       args,
		CalledFunc: op.CalledFunc,
		Result:     op.Result,
	}
}

type INCLUDE_TYPE int

const (
	TYPE_INCLUDE INCLUDE_TYPE = iota
	TYPE_INCLUDE_ONCE
	TYPE_REQUIRE
	TYPE_REQUIRE_ONCE
)

type OpExprInclude struct {
	OpGeneral
	Type   INCLUDE_TYPE
	Expr   Operand
	Result Operand
}

func NewOpExprInclude(expr Operand, tp INCLUDE_TYPE, pos *position.Position) *OpExprInclude {
	Op := &OpExprInclude{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Type:   tp,
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprInclude) IncludeTypeStr() string {
	switch op.Type {
	case TYPE_INCLUDE:
		return "Include"
	case TYPE_INCLUDE_ONCE:
		return "IncludeOnce"
	case TYPE_REQUIRE:
		return "Require"
	case TYPE_REQUIRE_ONCE:
		return "RequireOnce"
	}
	return ""
}

func (op *OpExprInclude) GetType() string {
	return "ExprInclude"
}

func (op *OpExprInclude) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprInclude) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprInclude) Clone() Op {
	return &OpExprInclude{
		OpGeneral: op.OpGeneral,
		Type:      op.Type,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprInstanceOf struct {
	OpGeneral
	Expr   Operand
	Class  Operand
	Result Operand
}

func NewOpExprInstanceOf(expr Operand, class Operand, pos *position.Position) *OpExprInstanceOf {
	Op := &OpExprInstanceOf{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Class:  class,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, expr, class)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprInstanceOf) GetType() string {
	return "ExprInstanceOf"
}

func (op *OpExprInstanceOf) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Class":  op.Class,
		"Result": op.Result,
	}
}

func (op *OpExprInstanceOf) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Class":
		op.Class = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprInstanceOf) Clone() Op {
	return &OpExprInstanceOf{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Class:     op.Class,
		Result:    op.Result,
	}
}

type OpExprIsset struct {
	OpGeneral
	Vars   []Operand
	Result Operand
}

func NewOpExprIsset(vars []Operand, pos *position.Position) *OpExprIsset {
	Op := &OpExprIsset{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Vars:   vars,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, vars...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprIsset) GetType() string {
	return "ExprIsset"
}

func (op *OpExprIsset) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Result": op.Result,
	}
}

func (op *OpExprIsset) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprIsset) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Vars": op.Vars,
	}
}

func (op *OpExprIsset) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Vars":
		op.Vars = vr
	}
}

func (op *OpExprIsset) Clone() Op {
	vars := make([]Operand, len(op.Vars))
	copy(vars, op.Vars)
	return &OpExprIsset{
		OpGeneral: op.OpGeneral,
		Vars:      vars,
		Result:    op.Result,
	}
}

type OpExprMethodCall struct {
	OpGeneral
	Var    Operand
	Name   Operand
	Args   []Operand
	Result Operand

	IsNullSafe bool
	CalledFunc OpCallable
	VarPos     *position.Position
	NamePos    *position.Position
	ArgsPos    []*position.Position
}

func NewOpExprMethodCall(vr, name Operand, args []Operand, varPos, namePos *position.Position, argsPos []*position.Position, pos *position.Position) *OpExprMethodCall {
	Op := &OpExprMethodCall{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var:        vr,
		Name:       name,
		Args:       args,
		Result:     NewOperTemporary(nil),
		IsNullSafe: false,
		VarPos:     varPos,
		NamePos:    namePos,
		ArgsPos:    argsPos,
	}

	AddReadRefs(Op, vr, name)
	AddReadRefs(Op, args...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func NewOpExprNullSafeMethodCall(vr, name Operand, args []Operand, varPos, namePos *position.Position, argsPos []*position.Position, pos *position.Position) *OpExprMethodCall {
	Op := &OpExprMethodCall{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var:        vr,
		Name:       name,
		Args:       args,
		Result:     NewOperTemporary(nil),
		IsNullSafe: true,
		VarPos:     varPos,
		NamePos:    namePos,
		ArgsPos:    argsPos,
	}

	AddReadRefs(Op, vr, name)
	AddReadRefs(Op, args...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprMethodCall) GetType() string {
	return "ExprMethodCall"
}

// func (op *OpExprMethodCall) GetName() string {
// 	className := op.Var.(*OperObjectVar).ClassName
// 	return GetOperName(op.Name)
// }

func (op *OpExprMethodCall) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Name":   op.Name,
		"Result": op.Result,
	}
}

func (op *OpExprMethodCall) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Name":
		op.Name = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprMethodCall) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Args": op.Args,
	}
}

func (op *OpExprMethodCall) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Args":
		op.Args = vr
	}
}

func (op *OpExprMethodCall) GetOpVarPos(vrName string) *position.Position {
	switch vrName {
	case "Var":
		return op.VarPos
	case "Name":
		return op.NamePos
	}
	return nil
}

func (op *OpExprMethodCall) GetOpVarListPos(vrName string, index int) *position.Position {
	switch vrName {
	case "Args":
		return op.ArgsPos[index]
	}
	return nil
}

func (op *OpExprMethodCall) GetName() string {
	// get class name
	className := ""
	switch c := op.Var.(type) {
	case *OperObject:
		className = c.ClassName
	case *OperVariable:
		if cv, ok := c.Value.(*OperObject); ok {
			className = cv.ClassName
		}
	case *OperTemporary:
		if co, ok := c.Original.(*OperVariable); ok {
			if cv, ok := co.Value.(*OperObject); ok {
				className = cv.ClassName
			}
		}
	}
	funcName, _ := GetOperName(op.Name)
	// if err != nil {
	// 	log.Fatalf("Error in GetMethodCallName: %v", err)
	// }
	return className + "::" + funcName
}

func (op *OpExprMethodCall) Clone() Op {
	args := make([]Operand, len(op.Args))
	copy(args, op.Args)
	return &OpExprMethodCall{
		OpGeneral:  op.OpGeneral,
		Var:        op.Var,
		Name:       op.Name,
		Args:       args,
		IsNullSafe: op.IsNullSafe,
		CalledFunc: op.CalledFunc,
		Result:     op.Result,
	}
}

type OpExprNew struct {
	OpGeneral
	Class  Operand
	Args   []Operand
	Result Operand
}

func NewOpExprNew(class Operand, args []Operand, pos *position.Position) *OpExprNew {
	Op := &OpExprNew{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Class:  class,
		Args:   args,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, class)
	AddReadRefs(Op, args...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprNew) GetType() string {
	return "ExprNew"
}

func (op *OpExprNew) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Class":  op.Class,
		"Result": op.Result,
	}
}

func (op *OpExprNew) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Class":
		op.Class = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprNew) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Args": op.Args,
	}
}

func (op *OpExprNew) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Args":
		op.Args = vr
	}
}

func (op *OpExprNew) Clone() Op {
	args := make([]Operand, len(op.Args))
	copy(args, op.Args)
	return &OpExprNew{
		OpGeneral: op.OpGeneral,
		Args:      args,
		Class:     op.Class,
		Result:    op.Result,
	}
}

type OpExprParam struct {
	OpGeneral
	Name         Operand
	ByRef        bool
	Variadic     bool
	AttrGroups   []*OpAttributeGroup
	DefaultVar   Operand
	DefaultBlock *Block
	DeclaredType OpType
	Result       Operand

	Func *OpFunc // Helper
}

func NewOpExprParam(name, defaultVar Operand, defaultBlock *Block, byRef, variadic bool, attrGroups []*OpAttributeGroup, declaredType OpType, pos *position.Position) *OpExprParam {
	Op := &OpExprParam{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name:         name,
		ByRef:        byRef,
		Variadic:     variadic,
		AttrGroups:   attrGroups,
		DefaultVar:   defaultVar,
		DefaultBlock: defaultBlock,
		DeclaredType: declaredType,
		Result:       NewOperTemporary(nil),
	}

	AddReadRefs(Op, name, defaultVar)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprParam) GetType() string {
	return "ExprParam"
}

func (op *OpExprParam) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name":       op.Name,
		"DefaultVar": op.DefaultVar,
		"Result":     op.Result,
	}
}

func (op *OpExprParam) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	case "DefaultVar":
		op.DefaultVar = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprParam) Clone() Op {
	attrGroups := make([]*OpAttributeGroup, len(op.AttrGroups))
	copy(attrGroups, op.AttrGroups)
	return &OpExprParam{
		OpGeneral:    op.OpGeneral,
		Name:         op.Name,
		ByRef:        op.ByRef,
		Variadic:     op.Variadic,
		AttrGroups:   attrGroups,
		DefaultVar:   op.DefaultVar,
		DefaultBlock: op.DefaultBlock,
		DeclaredType: op.DeclaredType,
		Func:         op.Func,
		Result:       op.Result,
	}
}

type OpExprPrint struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprPrint(expr Operand, pos *position.Position) *OpExprPrint {
	Op := &OpExprPrint{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprPrint) GetType() string {
	return "ExprPrint"
}

func (op *OpExprPrint) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprPrint) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprPrint) Clone() Op {
	return &OpExprPrint{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprPropertyFetch struct {
	OpGeneral
	Var    Operand
	Name   Operand
	Result Operand
}

func NewOpExprPropertyFetch(vr, name Operand, pos *position.Position) *OpExprPropertyFetch {
	Op := &OpExprPropertyFetch{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var:    vr,
		Name:   name,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, vr, name)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprPropertyFetch) GetType() string {
	return "ExprPropertyFetch"
}

func (op *OpExprPropertyFetch) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Name":   op.Name,
		"Result": op.Result,
	}
}

func (op *OpExprPropertyFetch) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Name":
		op.Name = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprPropertyFetch) Clone() Op {
	return &OpExprPropertyFetch{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
		Name:      op.Name,
		Result:    op.Result,
	}
}

type OpExprStaticCall struct {
	OpGeneral
	Class  Operand
	Name   Operand
	Args   []Operand
	Result Operand

	CalledFunc *OpFunc
	ClassPos   *position.Position
	NamePos    *position.Position
	ArgsPos    []*position.Position
}

func NewOpExprStaticCall(class, name Operand, args []Operand, classPos, namePos *position.Position, argsPos []*position.Position, pos *position.Position) *OpExprStaticCall {
	Op := &OpExprStaticCall{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Class:    class,
		Name:     name,
		Args:     args,
		Result:   NewOperTemporary(nil),
		ClassPos: classPos,
		NamePos:  namePos,
		ArgsPos:  argsPos,
	}

	AddReadRefs(Op, class, name)
	AddReadRefs(Op, args...)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprStaticCall) GetType() string {
	return "ExprStaticCall"
}

func (op *OpExprStaticCall) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Class":  op.Class,
		"Name":   op.Name,
		"Result": op.Result,
	}
}

func (op *OpExprStaticCall) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Class":
		op.Class = vr
	case "Name":
		op.Name = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprStaticCall) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Args": op.Args,
	}
}

func (op *OpExprStaticCall) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Args":
		op.Args = vr
	}
}

func (op *OpExprStaticCall) GetOpVarPos(vrName string) *position.Position {
	switch vrName {
	case "Class":
		return op.ClassPos
	case "Name":
		return op.NamePos
	}
	return nil
}

func (op *OpExprStaticCall) GetOpVarListPos(vrName string, index int) *position.Position {
	switch vrName {
	case "Args":
		return op.ArgsPos[index]
	}
	return nil
}

func (op *OpExprStaticCall) GetName() string {
	// get class name
	className := ""
	switch c := op.Class.(type) {
	case *OperObject:
		className = c.ClassName
	case *OperVariable:
		if cv, ok := c.Value.(*OperObject); ok {
			className = cv.ClassName
		}
	case *OperTemporary:
		if co, ok := c.Original.(*OperVariable); ok {
			if cv, ok := co.Value.(*OperObject); ok {
				className = cv.ClassName
			}
		}
	}
	funcName, err := GetOperName(op.Name)
	if err != nil {
		log.Fatalf("Error in GetStaticCallName: %v", err)
	}
	return className + "::" + funcName
}

func (op *OpExprStaticCall) Clone() Op {
	args := make([]Operand, len(op.Args))
	copy(args, op.Args)
	return &OpExprStaticCall{
		OpGeneral:  op.OpGeneral,
		Class:      op.Class,
		Name:       op.Name,
		Args:       args,
		CalledFunc: op.CalledFunc,
		Result:     op.Result,
	}
}

type OpExprStaticPropertyFetch struct {
	OpGeneral
	Class  Operand
	Name   Operand
	Result Operand
}

func NewOpExprStaticPropertyFetch(class, name Operand, pos *position.Position) *OpExprStaticPropertyFetch {
	Op := &OpExprStaticPropertyFetch{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Class:  class,
		Name:   name,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, class, name)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprStaticPropertyFetch) GetType() string {
	return "ExprStaticPropertyFetch"
}

func (op *OpExprStaticPropertyFetch) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Class":  op.Class,
		"Name":   op.Name,
		"Result": op.Result,
	}
}

func (op *OpExprStaticPropertyFetch) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Class":
		op.Class = vr
	case "Name":
		op.Name = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprStaticPropertyFetch) Clone() Op {
	return &OpExprStaticPropertyFetch{
		OpGeneral: op.OpGeneral,
		Class:     op.Class,
		Name:      op.Name,
		Result:    op.Result,
	}
}

type OpExprUnaryMinus struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprUnaryMinus(expr Operand, pos *position.Position) *OpExprUnaryMinus {
	Op := &OpExprUnaryMinus{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprUnaryMinus) GetType() string {
	return "ExprUnaryMinus"
}

func (op *OpExprUnaryMinus) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprUnaryMinus) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprUnaryMinus) Clone() Op {
	return &OpExprUnaryMinus{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprUnaryPlus struct {
	OpGeneral
	Expr   Operand
	Result Operand
}

func NewOpExprUnaryPlus(expr Operand, pos *position.Position) *OpExprUnaryPlus {
	Op := &OpExprUnaryPlus{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr:   expr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, expr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprUnaryPlus) GetType() string {
	return "ExprUnaryPlus"
}

func (op *OpExprUnaryPlus) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr":   op.Expr,
		"Result": op.Result,
	}
}

func (op *OpExprUnaryPlus) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprUnaryPlus) Clone() Op {
	return &OpExprUnaryPlus{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
		Result:    op.Result,
	}
}

type OpExprYield struct {
	OpGeneral
	Value  Operand
	Key    Operand
	Result Operand
}

func NewOpExprYield(value, key Operand, pos *position.Position) *OpExprYield {
	Op := &OpExprYield{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Value:  value,
		Key:    key,
		Result: NewOperTemporary(nil),
	}

	AddReadRefs(Op, value, key)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprYield) GetType() string {
	return "ExprYield"
}

func (op *OpExprYield) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Value":  op.Value,
		"Key":    op.Key,
		"Result": op.Result,
	}
}

func (op *OpExprYield) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Value":
		op.Value = vr
	case "Key":
		op.Key = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprYield) Clone() Op {
	return &OpExprYield{
		OpGeneral: op.OpGeneral,
		Value:     op.Value,
		Key:       op.Key,
		Result:    op.Result,
	}
}

type OpIterator interface {
	GetVar() Operand
}

type OpExprKey struct {
	OpGeneral
	Var    Operand
	Result Operand
}

func NewOpExprKey(vr Operand, pos *position.Position) *OpExprKey {
	Op := &OpExprKey{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var:    vr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, vr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprKey) GetVar() Operand {
	return op.Var
}

func (op *OpExprKey) GetType() string {
	return "ExprKey"
}

func (op *OpExprKey) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Result": op.Result,
	}
}

func (op *OpExprKey) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprKey) Clone() Op {
	return &OpExprKey{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
		Result:    op.Result,
	}
}

type OpExprValid struct {
	OpGeneral
	Var    Operand
	Result Operand
}

func NewOpExprValid(vr Operand, pos *position.Position) *OpExprValid {
	Op := &OpExprValid{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var:    vr,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, vr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprValid) GetVar() Operand {
	return op.Var
}

func (op *OpExprValid) GetType() string {
	return "ExprValid"
}

func (op *OpExprValid) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Result": op.Result,
	}
}

func (op *OpExprValid) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprValid) Clone() Op {
	return &OpExprValid{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
		Result:    op.Result,
	}
}

type OpExprValue struct {
	OpGeneral
	Var    Operand
	ByRef  bool
	Result Operand
}

func NewOpExprValue(vr Operand, byRef bool, pos *position.Position) *OpExprValue {
	Op := &OpExprValue{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var:    vr,
		ByRef:  byRef,
		Result: NewOperTemporary(nil),
	}

	AddReadRef(Op, vr)
	AddWriteRef(Op, Op.Result)

	return Op
}

func (op *OpExprValue) GetVar() Operand {
	return op.Var
}

func (op *OpExprValue) GetType() string {
	return "ExprValue"
}

func (op *OpExprValue) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":    op.Var,
		"Result": op.Result,
	}
}

func (op *OpExprValue) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "Result":
		op.Result = vr
	}
}

func (op *OpExprValue) Clone() Op {
	return &OpExprValue{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
		Result:    op.Result,
	}
}

type OpNext struct {
	OpGeneral
	Var Operand
}

func NewOpNext(vr Operand, pos *position.Position) *OpNext {
	Op := &OpNext{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var: vr,
	}

	AddReadRef(Op, vr)

	return Op
}

func (op *OpNext) GetVar() Operand {
	return op.Var
}

func (op *OpNext) GetType() string {
	return "Next"
}

func (op *OpNext) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var": op.Var,
	}
}

func (op *OpNext) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	}
}

func (op *OpNext) Clone() Op {
	return &OpNext{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
	}
}

type OpReset struct {
	OpGeneral
	Var Operand
}

func NewOpReset(vr Operand, pos *position.Position) *OpReset {
	Op := &OpReset{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var: vr,
	}

	AddReadRef(Op, vr)

	return Op
}

func (op *OpReset) GetVar() Operand {
	return op.Var
}

func (op *OpReset) GetType() string {
	return "Reset"
}

func (op *OpReset) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var": op.Var,
	}
}

func (op *OpReset) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	}
}

func (op *OpReset) Clone() Op {
	return &OpReset{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
	}
}

type ClassModifFlag int

const (
	CLASS_MODIF_PUBLIC    ClassModifFlag = 1
	CLASS_MODIF_PROTECTED ClassModifFlag = 1 << 1
	CLASS_MODIF_PRIVATE   ClassModifFlag = 1 << 2
	CLASS_MODIF_STATIC    ClassModifFlag = 1 << 3
	CLASS_MODIF_ABSTRACT  ClassModifFlag = 1 << 4
	CLASS_MODIF_FINAL     ClassModifFlag = 1 << 5
	CLASS_MODIF_READONLY  ClassModifFlag = 1 << 6
)

type OpStmtClass struct {
	OpGeneral
	Name       Operand
	Stmts      *Block
	Flags      ClassModifFlag // bitmask storing modifiers
	Extends    Operand
	Implements []Operand
	AttrGroups []*OpAttributeGroup
}

func NewOpStmtClass(name Operand, stmts *Block, flags ClassModifFlag, extends Operand, implements []Operand, attrGroups []*OpAttributeGroup, pos *position.Position) *OpStmtClass {
	Op := &OpStmtClass{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name:       name,
		Stmts:      stmts,
		Flags:      flags,
		Extends:    extends,
		Implements: implements,
		AttrGroups: attrGroups,
	}

	AddReadRef(Op, name)

	return Op
}

func (op *OpStmtClass) IsPublic() bool {
	return op.Flags&CLASS_MODIF_PUBLIC == 1
}

func (op *OpStmtClass) IsProtected() bool {
	return op.Flags&CLASS_MODIF_PROTECTED == 1
}

func (op *OpStmtClass) IsPrivate() bool {
	return op.Flags&CLASS_MODIF_PRIVATE == 1
}

func (op *OpStmtClass) IsStatic() bool {
	return op.Flags&CLASS_MODIF_STATIC == 1
}

func (op *OpStmtClass) IsAbstract() bool {
	return op.Flags&CLASS_MODIF_ABSTRACT == 1
}

func (op *OpStmtClass) IsFinal() bool {
	return op.Flags&CLASS_MODIF_FINAL == 1
}

func (op *OpStmtClass) IsReadonly() bool {
	return op.Flags&CLASS_MODIF_READONLY == 1
}

func (op *OpStmtClass) GetType() string {
	return "StmtClass"
}

func (op *OpStmtClass) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name":    op.Name,
		"Extends": op.Extends,
	}
}

func (op *OpStmtClass) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	case "Extends":
		op.Extends = vr
	}
}

func (op *OpStmtClass) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Implements": op.Implements,
	}
}

func (op *OpStmtClass) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Implements":
		op.Implements = vr
	}
}

func (op *OpStmtClass) Clone() Op {
	implements := make([]Operand, len(op.Implements))
	copy(implements, op.Implements)
	return &OpStmtClass{
		OpGeneral:  op.OpGeneral,
		Name:       op.Name,
		Stmts:      op.Stmts,
		Flags:      op.Flags,
		Extends:    op.Extends,
		Implements: implements,
		AttrGroups: op.AttrGroups,
	}
}

type OpStmtClassMethod struct {
	OpGeneral
	Func       *OpFunc
	AttrGroups []*OpAttributeGroup
	Visibility FuncModifFlag
	Static     bool
	Final      bool
	Abstract   bool
}

func NewOpStmtClassMethod(function *OpFunc, attrGroups []*OpAttributeGroup, visibility FuncModifFlag, static bool, final bool, abstract bool, pos *position.Position) *OpStmtClassMethod {
	Op := &OpStmtClassMethod{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Func:       function,
		AttrGroups: attrGroups,
		Visibility: visibility,
		Static:     static,
		Final:      final,
		Abstract:   abstract,
	}

	return Op
}

func (op *OpStmtClassMethod) GetType() string {
	return "StmtClassMethod"
}

func (op *OpStmtClassMethod) GetFunc() *OpFunc {
	return op.Func
}

func (op *OpStmtClassMethod) Clone() Op {
	attrGroups := make([]*OpAttributeGroup, len(op.AttrGroups))
	copy(attrGroups, op.AttrGroups)
	return &OpStmtClassMethod{
		OpGeneral:  op.OpGeneral,
		Func:       op.Func,
		AttrGroups: attrGroups,
		Visibility: op.Visibility,
		Static:     op.Static,
		Final:      op.Final,
		Abstract:   op.Abstract,
	}
}

type OpStmtFunc struct {
	OpGeneral
	Func       *OpFunc
	AttrGroups []*OpAttributeGroup
}

func NewOpStmtFunc(function *OpFunc, attrGroups []*OpAttributeGroup, pos *position.Position) *OpStmtFunc {
	Op := &OpStmtFunc{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Func:       function,
		AttrGroups: attrGroups,
	}

	return Op
}

func (op *OpStmtFunc) GetType() string {
	return "StmtFunc"
}

func (op *OpStmtFunc) GetFunc() *OpFunc {
	return op.Func
}

func (op *OpStmtFunc) Clone() Op {
	attrGroups := make([]*OpAttributeGroup, len(op.AttrGroups))
	copy(attrGroups, op.AttrGroups)
	return &OpStmtFunc{
		OpGeneral:  op.OpGeneral,
		Func:       op.Func,
		AttrGroups: attrGroups,
	}
}

type OpStmtInterface struct {
	OpGeneral
	Name    Operand
	Stmts   *Block
	Extends []Operand
}

func NewOpStmtInterface(name Operand, stmts *Block, extends []Operand, pos *position.Position) *OpStmtInterface {
	Op := &OpStmtInterface{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name:    name,
		Stmts:   stmts,
		Extends: extends,
	}

	AddReadRef(Op, name)

	return Op
}

func (op *OpStmtInterface) GetType() string {
	return "StmtInterface"
}

func (op *OpStmtInterface) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name": op.Name,
	}
}

func (op *OpStmtInterface) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	}
}

func (op *OpStmtInterface) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Extends": op.Extends,
	}
}

func (op *OpStmtInterface) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Extends":
		op.Extends = vr
	}
}

func (op *OpStmtInterface) Clone() Op {
	extends := make([]Operand, len(op.Extends))
	copy(extends, op.Extends)
	return &OpStmtInterface{
		OpGeneral: op.OpGeneral,
		Name:      op.Name,
		Stmts:     op.Stmts,
		Extends:   extends,
	}
}

type OpStmtJump struct {
	OpGeneral
	Target *Block
}

func NewOpStmtJump(target *Block, pos *position.Position) *OpStmtJump {
	Op := &OpStmtJump{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Target: target,
	}

	return Op
}

func (op *OpStmtJump) GetType() string {
	return "StmtJump"
}

func (op *OpStmtJump) Clone() Op {
	return &OpStmtJump{
		OpGeneral: op.OpGeneral,
		Target:    op.Target,
	}
}

type OpStmtJumpIf struct {
	OpGeneral
	Cond Operand
	If   *Block
	Else *Block
}

func NewOpStmtJumpIf(cond Operand, ifBlock *Block, elseBlock *Block, pos *position.Position) *OpStmtJumpIf {
	Op := &OpStmtJumpIf{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Cond: cond,
		If:   ifBlock,
		Else: elseBlock,
	}

	AddReadRef(Op, cond)

	return Op
}

func (op *OpStmtJumpIf) GetType() string {
	return "StmtJumpIf"
}

func (op *OpStmtJumpIf) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Cond": op.Cond,
	}
}

func (op *OpStmtJumpIf) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Cond":
		op.Cond = vr
	}
}

func (op *OpStmtJumpIf) Clone() Op {
	return &OpStmtJumpIf{
		OpGeneral: op.OpGeneral,
		Cond:      op.Cond,
		If:        op.If,
		Else:      op.Else,
	}
}

type OpStmtProperty struct {
	OpGeneral
	Name         Operand
	Visibility   ClassModifFlag
	Static       bool
	ReadOnly     bool
	AttrGroups   []*OpAttributeGroup
	DefaultVar   Operand
	DefaultBlock *Block
	DeclaredType OpType
}

func NewOpStmtProperty(name Operand, visibility ClassModifFlag, static bool, readOnly bool, attrGroups []*OpAttributeGroup,
	defaultVar Operand, defaultBlock *Block, declaredType OpType, pos *position.Position) *OpStmtProperty {
	Op := &OpStmtProperty{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name:         name,
		Visibility:   visibility,
		Static:       static,
		ReadOnly:     readOnly,
		AttrGroups:   attrGroups,
		DefaultVar:   defaultVar,
		DefaultBlock: defaultBlock,
		DeclaredType: declaredType,
	}

	return Op
}

func (op *OpStmtProperty) IsPublic() bool {
	return op.Visibility == CLASS_MODIF_PUBLIC
}

func (op *OpStmtProperty) IsProtected() bool {
	return op.Visibility == CLASS_MODIF_PROTECTED
}

func (op *OpStmtProperty) IsPrivate() bool {
	return op.Visibility == CLASS_MODIF_PRIVATE
}

func (op *OpStmtProperty) GetType() string {
	return "StmtProperty"
}

func (op *OpStmtProperty) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name":       op.Name,
		"DefaultVar": op.DefaultVar,
	}
}

func (op *OpStmtProperty) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	case "DefaultVar":
		op.DefaultVar = vr
	}
}

func (op *OpStmtProperty) Clone() Op {
	attrGroups := make([]*OpAttributeGroup, len(op.AttrGroups))
	copy(attrGroups, op.AttrGroups)
	return &OpStmtProperty{
		OpGeneral:    op.OpGeneral,
		Name:         op.Name,
		Visibility:   op.Visibility,
		Static:       op.Static,
		ReadOnly:     op.ReadOnly,
		AttrGroups:   attrGroups,
		DefaultVar:   op.DefaultVar,
		DefaultBlock: op.DefaultBlock,
		DeclaredType: op.DeclaredType,
	}
}

type OpStmtSwitch struct {
	OpGeneral
	Cond          Operand
	Cases         []Operand
	Targets       []*Block
	DefaultTarget *Block
}

func NewOpStmtSwitch(cond Operand, cases []Operand, targets []*Block, defaultTarget *Block, pos *position.Position) *OpStmtSwitch {
	Op := &OpStmtSwitch{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Cond:          cond,
		Cases:         cases,
		Targets:       targets,
		DefaultTarget: defaultTarget,
	}

	AddReadRef(Op, cond)
	AddReadRefs(Op, cases...)

	return Op
}

func (op *OpStmtSwitch) GetType() string {
	return "StmtSwitch"
}

func (op *OpStmtSwitch) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Cond": op.Cond,
	}
}

func (op *OpStmtSwitch) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Cond":
		op.Cond = vr
	}
}

func (op *OpStmtSwitch) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Cases": op.Cases,
	}
}

func (op *OpStmtSwitch) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Cases":
		op.Cases = vr
	}
}

func (op *OpStmtSwitch) Clone() Op {
	return &OpStmtSwitch{
		OpGeneral:     op.OpGeneral,
		Cond:          op.Cond,
		Cases:         op.Cases,
		Targets:       op.Targets,
		DefaultTarget: op.DefaultTarget,
	}
}

type OpStmtTrait struct {
	OpGeneral
	Name  Operand
	Stmts *Block
}

func NewOpStmtTrait(name Operand, stmts *Block, pos *position.Position) *OpStmtTrait {
	Op := &OpStmtTrait{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name:  name,
		Stmts: stmts,
	}

	AddReadRef(Op, name)

	return Op
}

func (op *OpStmtTrait) GetType() string {
	return "StmtTrait"
}

func (op *OpStmtTrait) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name": op.Name,
	}
}

func (op *OpStmtTrait) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	}
}

func (op *OpStmtTrait) Clone() Op {
	return &OpStmtTrait{
		OpGeneral: op.OpGeneral,
		Name:      op.Name,
		Stmts:     op.Stmts,
	}
}

type OpStmtTraitUse struct {
	OpGeneral
	Traits      []Operand
	Adaptations []Op
}

func NewOpStmtTraitUse(traits []Operand, adaptations []Op, pos *position.Position) *OpStmtTraitUse {
	Op := &OpStmtTraitUse{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Traits:      traits,
		Adaptations: adaptations,
	}

	return Op
}

func (op *OpStmtTraitUse) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Traits": op.Traits,
	}
}

func (op *OpStmtTraitUse) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Traits":
		op.Traits = vr
	}
}

func (op *OpStmtTraitUse) GetType() string {
	return "StmtTraitUse"
}

func (op *OpStmtTraitUse) Clone() Op {
	return &OpStmtTraitUse{
		OpGeneral:   op.OpGeneral,
		Traits:      op.Traits,
		Adaptations: op.Adaptations,
	}
}

type OpAlias struct {
	OpGeneral
	Trait       Operand
	Method      Operand
	NewName     Operand
	NewModifier ClassModifFlag
}

func NewOpAlias(trait, method, newName Operand, newModifier ClassModifFlag, pos *position.Position) *OpAlias {
	Op := &OpAlias{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Trait:       trait,
		Method:      method,
		NewName:     newName,
		NewModifier: newModifier,
	}

	return Op
}

func (op *OpAlias) GetType() string {
	return "Alias"
}

func (op *OpAlias) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Trait":   op.Trait,
		"Method":  op.Method,
		"NewName": op.NewName,
	}
}

func (op *OpAlias) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Trait":
		op.Trait = vr
	case "Method":
		op.Method = vr
	case "NewName":
		op.NewName = vr
	}
}

func (op *OpAlias) Clone() Op {
	return &OpAlias{
		OpGeneral:   op.OpGeneral,
		Trait:       op.Trait,
		Method:      op.Method,
		NewName:     op.NewName,
		NewModifier: op.NewModifier,
	}
}

type OpPrecedence struct {
	OpGeneral
	Trait     Operand
	Method    Operand
	InsteadOf []Operand
}

func NewOpPrecedence(trait, method Operand, insteadOf []Operand, pos *position.Position) *OpPrecedence {
	Op := &OpPrecedence{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Trait:     trait,
		Method:    method,
		InsteadOf: insteadOf,
	}

	return Op
}

func (op *OpPrecedence) GetType() string {
	return "Precedence"
}

func (op *OpPrecedence) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Trait":  op.Trait,
		"Method": op.Method,
	}
}

func (op *OpPrecedence) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Trait":
		op.Trait = vr
	case "Method":
		op.Method = vr
	}
}

func (op *OpPrecedence) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Insteadof": op.InsteadOf,
	}
}

func (op *OpPrecedence) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Insteadof":
		op.InsteadOf = vr
	}
}

func (op *OpPrecedence) Clone() Op {
	return &OpPrecedence{
		OpGeneral: op.OpGeneral,
		Trait:     op.Trait,
		Method:    op.Method,
		InsteadOf: op.InsteadOf,
	}
}

type OpConst struct {
	OpGeneral
	Name       Operand
	Value      Operand
	ValueBlock *Block
}

func NewOpConst(name, value Operand, block *Block, pos *position.Position) *OpConst {
	Op := &OpConst{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name:       name,
		Value:      value,
		ValueBlock: block,
	}

	AddReadRefs(Op, name, value)

	return Op
}

func (op *OpConst) GetType() string {
	return "Const"
}

func (op *OpConst) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Name":  op.Name,
		"Value": op.Value,
	}
}

func (op *OpConst) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Name":
		op.Name = vr
	case "Value":
		op.Value = vr
	}
}

func (op *OpConst) Clone() Op {
	return &OpConst{
		OpGeneral:  op.OpGeneral,
		Name:       op.Name,
		Value:      op.Value,
		ValueBlock: op.ValueBlock,
	}
}

type OpEcho struct {
	OpGeneral
	Expr Operand
}

func NewOpEcho(expr Operand, pos *position.Position) *OpEcho {
	Op := &OpEcho{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr: expr,
	}

	AddReadRef(Op, expr)

	return Op
}

func (op *OpEcho) GetType() string {
	return "Echo"
}

func (op *OpEcho) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr": op.Expr,
	}
}

func (op *OpEcho) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	}
}

func (op *OpEcho) Clone() Op {
	return &OpEcho{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
	}
}

type OpExit struct {
	OpGeneral
	Expr Operand
}

func NewOpExit(expr Operand, pos *position.Position) *OpExit {
	Op := &OpExit{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr: expr,
	}

	AddReadRef(Op, expr)

	return Op
}

func (op *OpExit) GetType() string {
	return "Exit"
}

func (op *OpExit) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr": op.Expr,
	}
}

func (op *OpExit) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	}
}

func (op *OpExit) Clone() Op {
	return &OpExit{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
	}
}

type OpGlobalVar struct {
	OpGeneral
	Var Operand
}

func NewOpGlobalVar(vr Operand, pos *position.Position) *OpGlobalVar {
	Op := &OpGlobalVar{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var: vr,
	}

	AddReadRef(Op, vr)

	return Op
}

func (op *OpGlobalVar) GetType() string {
	return "GlobalVar"
}

func (op *OpGlobalVar) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var": op.Var,
	}
}

func (op *OpGlobalVar) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	}
}

func (op *OpGlobalVar) Clone() Op {
	return &OpGlobalVar{
		OpGeneral: op.OpGeneral,
		Var:       op.Var,
	}
}

type OpReturn struct {
	OpGeneral
	Expr Operand
}

func NewOpReturn(expr Operand, pos *position.Position) *OpReturn {
	Op := &OpReturn{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr: expr,
	}

	AddReadRef(Op, expr)

	return Op
}

func (op *OpReturn) GetType() string {
	return "Return"
}

func (op *OpReturn) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr": op.Expr,
	}
}

func (op *OpReturn) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	}
}

func (op *OpReturn) Clone() Op {
	return &OpReturn{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
	}
}

type OpStaticVar struct {
	OpGeneral
	Var          Operand
	DefaultVar   Operand
	DefaultBlock *Block
}

func NewOpStaticVar(vr, defaultVr Operand, defaultBlock *Block, pos *position.Position) *OpStaticVar {
	Op := &OpStaticVar{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Var:          vr,
		DefaultVar:   defaultVr,
		DefaultBlock: defaultBlock,
	}

	AddReadRef(Op, defaultVr)
	AddWriteRef(Op, vr)

	return Op
}

func (op *OpStaticVar) GetType() string {
	return "StaticVar"
}

func (op *OpStaticVar) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Var":        op.Var,
		"DefaultVar": op.DefaultVar,
	}
}

func (op *OpStaticVar) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Var":
		op.Var = vr
	case "DefaultVar":
		op.DefaultVar = vr
	}
}

func (op *OpStaticVar) Clone() Op {
	return &OpStaticVar{
		OpGeneral:    op.OpGeneral,
		Var:          op.Var,
		DefaultVar:   op.DefaultVar,
		DefaultBlock: op.DefaultBlock,
	}
}

type OpThrow struct {
	OpGeneral
	Expr Operand
}

func NewOpThrow(expr Operand, pos *position.Position) *OpThrow {
	Op := &OpThrow{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Expr: expr,
	}

	AddReadRef(Op, expr)

	return Op
}

func (op *OpThrow) GetType() string {
	return "Throq"
}

func (op *OpThrow) GetOpVars() map[string]Operand {
	return map[string]Operand{
		"Expr": op.Expr,
	}
}

func (op *OpThrow) ChangeOpVar(vrName string, vr Operand) {
	switch vrName {
	case "Expr":
		op.Expr = vr
	}
}

func (op *OpThrow) Clone() Op {
	return &OpThrow{
		OpGeneral: op.OpGeneral,
		Expr:      op.Expr,
	}
}

type OpUnset struct {
	OpGeneral
	Exprs []Operand

	ExprsPos []*position.Position
}

func NewOpUnset(exprs []Operand, exprsPos []*position.Position, pos *position.Position) *OpUnset {
	Op := &OpUnset{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Exprs:    exprs,
		ExprsPos: exprsPos,
	}

	AddReadRefs(Op, exprs...)

	return Op
}

func (op *OpUnset) GetType() string {
	return "Unset"
}

func (op *OpUnset) GetOpListVars() map[string][]Operand {
	return map[string][]Operand{
		"Exprs": op.Exprs,
	}
}

func (op *OpUnset) ChangeOpListVar(vrName string, vr []Operand) {
	switch vrName {
	case "Exprs":
		op.Exprs = vr
	}
}

func (op *OpUnset) Clone() Op {
	return &OpUnset{
		OpGeneral: op.OpGeneral,
		Exprs:     op.Exprs,
	}
}

type VAR_TYPE int

const (
	TYPE_MIXED VAR_TYPE = iota
	TYPE_REFERENCE
	TYPE_VOID
	TYPE_LITERAL
	TYPE_UNION
)

// Type Union
// Type Literal
// Type Reference
// Type Void
// Type Mixed
type OpType interface {
	Kind() VAR_TYPE
	Nullable() bool
}

type OpTypeUnion struct {
	OpGeneral
	Types []OpType
}

func NewOpTypeUnion(types []OpType, pos *position.Position) *OpTypeUnion {
	return &OpTypeUnion{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Types: types,
	}
}

func (op *OpTypeUnion) Kind() VAR_TYPE {
	return TYPE_UNION
}

func (op *OpTypeUnion) Nullable() bool {
	return false
}

func (op *OpTypeUnion) GetType() string {
	return "TypeUnion"
}

type OpTypeLiteral struct {
	OpGeneral
	Name       string
	IsNullable bool
}

func NewOpTypeLiteral(name string, nullable bool, pos *position.Position) *OpTypeLiteral {
	return &OpTypeLiteral{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Name:       name,
		IsNullable: nullable,
	}
}

func (op *OpTypeLiteral) Kind() VAR_TYPE {
	return TYPE_LITERAL
}

func (op *OpTypeLiteral) Nullable() bool {
	return op.IsNullable
}

func (op *OpTypeLiteral) GetType() string {
	return "TypeLiteral"
}

type OpTypeReference struct {
	OpGeneral
	Declaration Operand
	IsNullable  bool
}

func NewOpTypeReference(name Operand, nullable bool, pos *position.Position) *OpTypeReference {
	return &OpTypeReference{
		OpGeneral: OpGeneral{
			Position: pos,
		},
		Declaration: name,
		IsNullable:  nullable,
	}
}

func (op *OpTypeReference) Kind() VAR_TYPE {
	return TYPE_REFERENCE
}

func (op *OpTypeReference) Nullable() bool {
	return op.IsNullable
}

func (op *OpTypeReference) GetType() string {
	return "TypeReference"
}

type OpTypeVoid struct {
	OpGeneral
}

func NewOpTypeVoid(pos *position.Position) *OpTypeVoid {
	return &OpTypeVoid{
		OpGeneral: OpGeneral{
			Position: pos,
		},
	}
}

func (op *OpTypeVoid) Kind() VAR_TYPE {
	return TYPE_VOID
}

func (op *OpTypeVoid) Nullable() bool {
	return false
}

func (op *OpTypeVoid) GetType() string {
	return "TypeVoid"
}

type OpTypeMixed struct {
	OpGeneral
}

func NewOpTypeMixed(pos *position.Position) *OpTypeMixed {
	return &OpTypeMixed{
		OpGeneral: OpGeneral{
			Position: pos,
		},
	}
}

func (op *OpTypeMixed) Kind() VAR_TYPE {
	return TYPE_MIXED
}

func (op *OpTypeMixed) Nullable() bool {
	return false
}

func (op *OpTypeMixed) GetType() string {
	return "TypeMixed"
}
