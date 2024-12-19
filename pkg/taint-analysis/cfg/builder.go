package cfg

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taint-analysis/asttraverser"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taint-analysis/asttraverser/loopresolver"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taint-analysis/asttraverser/mcresolver"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taint-analysis/asttraverser/nsresolver"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taint-analysis/astutil"
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/position"
	"github.com/VKCOM/php-parser/pkg/version"
)

type VAR_MODE int

const (
	MODE_NONE VAR_MODE = iota + 999
	MODE_READ
	MODE_WRITE
)

type Script struct {
	Funcs         map[string]*OpFunc
	OrderedFuncs  []*OpFunc
	IncludedFiles []string
	Main          *OpFunc
}

func (s *Script) AddFunc(fn *OpFunc) {
	name := fn.GetScopedName()
	s.Funcs[name] = fn
	s.OrderedFuncs = append(s.OrderedFuncs, fn)
}

type CfgBuilder struct {
	// AstTraverser  []any  // TODO: change type
	AutloadConfig map[string]string
	Ctx           FuncCtx
	CurrClass     *OperString
	CurrNamespace string // TODO: check type again
	Script        *Script
	AnonId        int     // id for naming anonymous thing like closure
	BlockCnt      BlockId // counter to generate block id
	Consts        map[string]Operand

	currBlock *Block
	currFunc  *OpFunc
}

func BuildCFG(src []byte, autoloadConfig map[string]string) *Script {
	cb := &CfgBuilder{
		AutloadConfig: autoloadConfig,
		Consts:        make(map[string]Operand),
		AnonId:        0,
	}

	// Error handler
	var parserErrors []*errors.Error
	errorHandler := func(e *errors.Error) {
		parserErrors = append(parserErrors, e)
	}

	// Parsing source code into AST
	rootNode, err := parser.Parse(src, conf.Config{
		Version:          &version.Version{Major: 8},
		ErrorHandlerFunc: errorHandler,
	})
	root, ok := rootNode.(*ast.Root)

	if err != nil || !ok {
		log.Fatal("Error:" + err.Error())
	}

	// Resolve name, loop, and magic constant for easier analysis
	nsRes := nsresolver.NewNamespaceResolver()
	lRes := loopresolver.NewLoopResolver()
	mcRes := mcresolver.NewMagicConstResolver()
	travs := asttraverser.NewTraverser()
	travs.AddNodeTraverser(nsRes, lRes, mcRes)
	travs.Traverse(root)

	cb.parseRoot(root)

	// TODO: recheck return value
	return cb.Script
}

func (cb *CfgBuilder) parseRoot(n *ast.Root) {
	// Create script instance
	entryBlock := NewBlock(cb.GetBlockId())
	mainFunc, err := NewFunc("{main}", FUNC_MODIF_PUBLIC, NewOpTypeVoid(nil), entryBlock, nil)
	if err != nil {
		log.Fatalf("Error in parseRoot: %v", err)
	}
	cb.Script = &Script{
		Funcs: make(map[string]*OpFunc),
		Main:  mainFunc,
	}

	cb.parseOpFunc(mainFunc, nil, n.Stmts)
}

// Function to parse OpFunc and build cfg for a function
func (cb *CfgBuilder) parseOpFunc(fn *OpFunc, params []ast.Vertex, stmts []ast.Vertex) {
	prevFunc := cb.currFunc
	cb.currFunc = fn
	// create new func context
	prevContext := cb.Ctx
	cb.Ctx = NewFuncCtx()

	// set func cfg as entryBlock block
	entryBlock := fn.Cfg

	// set entry block as builder's current block
	prevBlock := cb.currBlock
	cb.currBlock = entryBlock
	// parse parameter list
	for _, par := range params {
		var defaultBlock *Block = nil
		var defaultVar Operand = nil
		param := par.(*ast.Parameter)
		if param.DefaultValue != nil {
			// parameter has a default value,
			// create different block to define default value
			tmp := cb.currBlock
			defaultBlock = NewBlock(cb.GetBlockId())
			cb.currBlock = defaultBlock
			defaultVar = cb.parseExprNode(param.DefaultValue)
			cb.currBlock = tmp
		}

		paramName := cb.parseExprNode(param.Var.(*ast.ExprVariable).Name).(*OperString)
		byRef := param.AmpersandTkn != nil
		isVariadic := param.VariadicTkn != nil
		paramAttrGroups := cb.parseAttrGroups(param.AttrGroups)
		paramType := cb.parseTypeNode(param.Type)
		opParam := NewOpExprParam(
			paramName,
			defaultVar,
			defaultBlock,
			byRef,
			isVariadic,
			paramAttrGroups,
			paramType,
			param.Position,
		)

		// TODO: check again
		opParam.Result.(*OperTemporary).Original = NewOperVar(paramName, nil)
		opParam.Func = fn
		fn.Params = append(fn.Params, opParam)

		// write each param as variable
		// append OpParam into block's instructions (children)
		cb.writeVariableName(paramName.Val, opParam.Result, entryBlock)
		entryBlock.AddInstructions(opParam)
	}

	// if there are statements, parse each of statements
	// else, set cfg to nil and reset the currBlock
	// parse each statements inside function
	endBlock, err := cb.parseStmtNodes(stmts, entryBlock) // it can create some blocks
	if err != nil {
		log.Fatalf("Error in parseOpFunc: %v", err)
	}

	// reset current block to previous
	cb.currBlock = prevBlock

	// if the end block isn't dead yet,
	// add OpReturn as predecessor
	if !endBlock.Dead {
		endBlock.AddInstructions(NewOpReturn(nil, nil))
	}

	// if there are still unresolved gotos, create an error
	if len(cb.Ctx.UnresolvedGotos) != 0 {
		log.Fatal("Error: there are still unresolved gotos")
	}

	cb.Ctx.Complete = true
	// resolve all incomplete phis
	for block := range cb.Ctx.incompletePhis {
		for name, phi := range cb.Ctx.incompletePhis[block] {
			for _, pred := range block.Preds {
				if !pred.Dead {
					vr := cb.readVariableName(name, pred)
					phi.AddOperand(vr)
				}
			}
			// append complete phi to the list
			block.AddPhi(phi)
		}
	}

	// reset function block and context
	cb.currFunc = prevFunc
	cb.Ctx = prevContext
}

// TODO: Implement parse nodes function
func (cb *CfgBuilder) parseStmtNodes(nodes []ast.Vertex, block *Block) (*Block, error) {
	if block == nil {
		return nil, fmt.Errorf("cannot parse nodes in nil block")
	}

	tmp := cb.currBlock
	cb.currBlock = block

	for _, node := range nodes {
		cb.parseStmtNode(node)
	}

	end := cb.currBlock
	cb.currBlock = tmp

	if end == nil {
		log.Fatalf("Error in parseStmtNodes: got nil as end block")
	}

	return end, nil
}

func (cb *CfgBuilder) parseStmtNode(node ast.Vertex) {
	switch n := node.(type) {
	case *ast.StmtExpression:
		cb.parseExprNode(n.Expr)
	case *ast.StmtConstList:
		// parse each constant
		cb.parseStmtConstList(n)
	case *ast.StmtConstant:
		cb.parseStmtConst(n)
	case *ast.StmtDeclare:
		cb.parseStmtDeclare(n)
	case *ast.StmtEcho:
		cb.parseStmtEcho(n)
	case *ast.StmtReturn:
		cb.parseStmtReturn(n)
	case *ast.StmtGlobal:
		cb.parseStmtGlobal(n)
	case *ast.StmtHaltCompiler:
		// do nothing
	case *ast.StmtInlineHtml:
		// do nothing
	case *ast.StmtNop:
		// do nothing
	case *ast.StmtStmtList:
		// Do nothing, handled in parseFunc()
	case *ast.StmtUnset:
		cb.parseStmtUnset(n)
	case *ast.StmtThrow:
		cb.parseStmtThrow(n)
	case *ast.StmtTry:
		cb.parseStmtTry(n)
	case *ast.StmtCatch:
		cb.parseStmtCatch(n)
	case *ast.StmtFinally:
		cb.parseStmtFinally(n)
	case *ast.StmtSwitch:
		cb.parseStmtSwitch(n)
	case *ast.StmtCase:
		// Do nothing, handled in parseStmtSwitch()
	case *ast.StmtDefault:
		// Do nothing, handled in parseStmtSwitch()
	case *ast.StmtIf:
		cb.parseStmtIf(n)
	case *ast.StmtElse:
		// Do nothing, handled in parseStmtIf()
	case *ast.StmtElseIf:
		// Do nothing, handled in parseStmtIf()
	case *ast.StmtContinue:
		// Do nothing, handled in loop resolver
	case *ast.StmtBreak:
		// Do nothing, handled in loop resolver
	case *ast.StmtGoto:
		cb.parseStmtGoto(n)
	case *ast.StmtLabel:
		cb.parseStmtLabel(n)
	case *ast.StmtDo:
		cb.parseStmtDo(n)
	case *ast.StmtFor:
		cb.parseStmtFor(n)
	case *ast.StmtForeach:
		cb.parseStmtForeach(n)
	case *ast.StmtWhile:
		cb.parseStmtWhile(n)
	case *ast.StmtEnum:
		// TODO: currently not needed
	case *ast.StmtClass:
		cb.parseStmtClass(n)
	case *ast.StmtClassConstList:
		cb.parseStmtClassConstList(n)
	case *ast.StmtClassMethod:
		cb.parseStmtClassMethod(n)
	case *ast.StmtInterface:
		cb.parseStmtInterface(n)
	case *ast.StmtPropertyList:
		cb.parseStmtPropertyList(n)
	case *ast.StmtProperty:
		// do nothing, handleed in parseStmtPropertyList()
	case *ast.StmtStatic:
		cb.parseStmtStatic(n)
	case *ast.StmtStaticVar:
		cb.parseStmtStaticVar(n)
	case *ast.StmtTrait:
		cb.parseStmtTrait(n)
	case *ast.StmtTraitUse:
		cb.parseStmtTraitUse(n)
	case *ast.StmtTraitUseAlias:
		// do nothing, handled in parseStmtTraitUse()
	case *ast.StmtTraitUsePrecedence:
		// do nothing, handled in parseStmtTraitUse()
	case *ast.StmtFunction:
		cb.parseStmtFunction(n)
	case *ast.StmtNamespace:
		cb.parseStmtNamespace(n)
	case *ast.StmtGroupUseList:
		// do nothing, have been handled by namespace resolver
	case *ast.StmtUseList:
		// do nothing, have been handled by namespace resolver
	case *ast.StmtUse:
		// do nothing, have been handled by namespace resolver
	default:
		log.Fatal("Error: Invalid statement node type")
	}
}

func (cb *CfgBuilder) parseStmtConstList(stmt *ast.StmtConstList) {
	for _, c := range stmt.Consts {
		cb.parseStmtNode(c)
	}
}

func (cb *CfgBuilder) parseStmtConst(stmt *ast.StmtConstant) {
	// create a new block for defining const
	tmp := cb.currBlock
	valBlock := NewBlock(cb.GetBlockId())
	cb.currBlock = valBlock
	val := cb.parseExprNode(stmt.Expr)
	cb.currBlock = tmp

	name := cb.readVariable(cb.parseExprNode(stmt.Name))
	opConst := NewOpConst(name, val, valBlock, stmt.Position)
	cb.currBlock.AddInstructions(opConst)

	// define the constant in this block
	nameStr := GetOperName(name)
	if cb.currFunc == cb.Script.Main {
		cb.Consts[nameStr] = val
	}
}

func (cb *CfgBuilder) parseStmtDeclare(stmt *ast.StmtDeclare) {
	// TODO: right now, it isn't important
}

func (cb *CfgBuilder) parseStmtEcho(stmt *ast.StmtEcho) {
	for _, expr := range stmt.Exprs {
		exprOper := cb.readVariable(cb.parseExprNode(expr))
		echoOp := NewOpEcho(exprOper, expr.GetPosition())
		cb.currBlock.AddInstructions(echoOp)
	}
}

func (cb *CfgBuilder) parseStmtReturn(stmt *ast.StmtReturn) {
	expr := Operand(nil)
	if stmt.Expr != nil {
		expr = cb.readVariable(cb.parseExprNode(stmt.Expr))
	}

	returnOp := NewOpReturn(expr, stmt.Position)
	cb.currBlock.AddInstructions(returnOp)

	// TODO: check again
	cb.currBlock = NewBlock(cb.GetBlockId())
	cb.currBlock.Dead = true
}

func (cb *CfgBuilder) parseStmtGlobal(stmt *ast.StmtGlobal) {
	for _, vr := range stmt.Vars {
		vrOper := cb.writeVariable(cb.parseExprNode(vr))
		op := NewOpGlobalVar(vrOper, vr.GetPosition())
		cb.currBlock.AddInstructions(op)
	}
}

func (cb *CfgBuilder) parseStmtUnset(stmt *ast.StmtUnset) {
	exprs := cb.parseExprList(stmt.Vars, MODE_WRITE)
	op := NewOpUnset(exprs, stmt.Position)
	cb.currBlock.AddInstructions(op)
}

func (cb *CfgBuilder) parseStmtThrow(stmt *ast.StmtThrow) {
	expr := cb.readVariable(cb.parseExprNode(stmt.Expr))
	op := NewOpThrow(expr, stmt.Position)
	cb.currBlock.AddInstructions(op)
	// script after throw will be a dead code
	cb.currBlock = NewBlock(cb.GetBlockId())
	cb.currBlock.Dead = true
}

func (cb *CfgBuilder) parseStmtTry(stmt *ast.StmtTry) {
	// TODO
}

func (cb *CfgBuilder) parseStmtCatch(stmt *ast.StmtCatch) {
	// TODO
}

func (cb *CfgBuilder) parseStmtFinally(stmt *ast.StmtFinally) {
	// TODO
}

func (cb *CfgBuilder) parseStmtSwitch(stmt *ast.StmtSwitch) {
	var err error
	if isJumpTableSwitch(stmt) {
		// build jump table switch
		cond := cb.readVariable(cb.parseExprNode(stmt.Cond))
		cases := make([]Operand, 0)
		targets := make([]*Block, 0)
		endBlock := NewBlock(cb.GetBlockId())
		defaultBlock := endBlock
		prevBlock := (*Block)(nil)

		for _, caseNode := range stmt.Cases {
			caseBlock := NewBlock(cb.GetBlockId())
			caseBlock.AddPredecessor(cb.currBlock)

			// case will be fallthrough if no break (prevBlock dead)
			if prevBlock != nil && !prevBlock.Dead {
				jmp := NewOpStmtJump(caseBlock, caseNode.GetPosition())
				prevBlock.AddInstructions(jmp)
				caseBlock.AddPredecessor(prevBlock)
			}

			switch cn := caseNode.(type) {
			case *ast.StmtCase:
				targets = append(targets, caseBlock)
				cases = append(cases, cb.parseExprNode(cn.Cond))
				prevBlock, err = cb.parseStmtNodes(cn.Stmts, caseBlock)
				if err != nil {
					log.Fatalf("Error in parseOpFunc: %v", err)
				}
			case *ast.StmtDefault:
				defaultBlock = caseBlock
				prevBlock, err = cb.parseStmtNodes(cn.Stmts, caseBlock)
				if err != nil {
					log.Fatalf("Error in parseOpFunc: %v", err)
				}
			default:
				log.Fatal("Error: Invalid case node type")
			}
		}

		switchOp := NewOpStmtSwitch(cond, cases, targets, defaultBlock, stmt.Position)
		cb.currBlock.AddInstructions(switchOp)

		if prevBlock != nil && !prevBlock.Dead {
			jmp := NewOpStmtJump(endBlock, stmt.Position)
			prevBlock.AddInstructions(jmp)
			endBlock.AddPredecessor(prevBlock)
		}

		cb.currBlock = endBlock
	} else {
		// build sequence of compare-and-jump
		cond := cb.parseExprNode(stmt.Cond)
		endBlock := NewBlock(cb.GetBlockId())
		defaultBlock := endBlock
		prevBlock := (*Block)(nil)

		for _, caseNode := range stmt.Cases {
			ifBlock := NewBlock(cb.GetBlockId())
			if prevBlock != nil && !prevBlock.Dead {
				jmp := NewOpStmtJump(ifBlock, caseNode.GetPosition())
				prevBlock.AddInstructions(jmp)
				ifBlock.AddPredecessor(prevBlock)
			}

			switch cn := caseNode.(type) {
			case *ast.StmtCase:
				caseExpr := cb.parseExprNode(cn.Cond)
				left := cb.readVariable(cond)
				right := cb.readVariable(caseExpr)
				opEqual := NewOpExprBinaryEqual(left, right, cn.Position)
				cb.currBlock.AddInstructions(opEqual)

				elseBlock := NewBlock(cb.GetBlockId())
				opJmpIf := NewOpStmtJumpIf(opEqual.Result, ifBlock, elseBlock, cn.Position)
				cb.currBlock.AddInstructions(opJmpIf)
				cb.currBlock.IsConditional = true
				ifBlock.AddPredecessor(cb.currBlock)
				elseBlock.AddPredecessor(cb.currBlock)
				cb.currBlock = elseBlock
				prevBlock, err = cb.parseStmtNodes(cn.Stmts, ifBlock)
				if err != nil {
					log.Fatalf("Error in parseStmtSwitch: %v", err)
				}
			case *ast.StmtDefault:
				defaultBlock = ifBlock
				prevBlock, err = cb.parseStmtNodes(cn.Stmts, ifBlock)
				if err != nil {
					log.Fatalf("Error in parseStmtSwitch: %v", err)
				}
			}
		}

		if prevBlock != nil && !prevBlock.Dead {
			jmp := NewOpStmtJump(endBlock, stmt.Position)
			prevBlock.AddInstructions(jmp)
			endBlock.AddPredecessor(prevBlock)
		}

		cb.currBlock.AddInstructions(NewOpStmtJump(defaultBlock, stmt.Position))
		defaultBlock.AddPredecessor(cb.currBlock)
		cb.currBlock = endBlock
	}
}

func isJumpTableSwitch(stmt *ast.StmtSwitch) bool {
	for _, cs := range stmt.Cases {
		// all case must be a scalar
		if !astutil.IsScalarNode(cs.(*ast.StmtCase).Cond) {
			return false
		}
	}
	return true
}

func (cb *CfgBuilder) parseStmtIf(stmt *ast.StmtIf) {
	endBlock := NewBlock(cb.GetBlockId())
	cb.parseIf(stmt, endBlock)
	cb.currBlock = endBlock
}

func (cb *CfgBuilder) parseIf(stmt ast.Vertex, endBlock *Block) {
	condPosition := &position.Position{}
	cond := Operand(nil)
	var stmts []ast.Vertex
	switch n := stmt.(type) {
	case *ast.StmtIf:
		condPosition = n.Cond.GetPosition()
		cond = cb.readVariable(cb.parseExprNode(n.Cond))
		if cond == nil {
			log.Fatal("afjsan")
		}
		stmts = n.Stmt.(*ast.StmtStmtList).Stmts
	case *ast.StmtElseIf:
		condPosition = n.Cond.GetPosition()
		cond = cb.readVariable(cb.parseExprNode(n.Cond))
		stmts = n.Stmt.(*ast.StmtStmtList).Stmts
	default:
		log.Fatal("Error: Invalid if node in parseIf()")
	}
	// create block for if and else
	ifBlock := NewBlock(cb.GetBlockId())
	ifBlock.AddPredecessor(cb.currBlock)
	elseBlock := NewBlock(cb.GetBlockId())
	elseBlock.AddPredecessor(cb.currBlock)

	jmpIf := NewOpStmtJumpIf(cond, ifBlock, elseBlock, condPosition)
	cb.currBlock.AddInstructions(jmpIf)
	cb.currBlock.IsConditional = true
	if cond == nil {
		log.Fatal("cond nil")
	}
	cb.processAssertion(cond, ifBlock, elseBlock)

	var err error
	cb.currBlock, err = cb.parseStmtNodes(stmts, ifBlock)
	if err != nil {
		log.Fatalf("Error in parseIf: %v", err)
	}

	jmp := NewOpStmtJump(endBlock, stmt.GetPosition())
	cb.currBlock.AddInstructions(jmp)
	endBlock.AddPredecessor(cb.currBlock)
	cb.currBlock = elseBlock

	if ifNode, ok := stmt.(*ast.StmtIf); ok {
		for _, elseIfNode := range ifNode.ElseIf {
			cb.parseIf(elseIfNode, endBlock)
		}
		if ifNode.Else != nil {
			cb.currBlock, err = cb.parseStmtNodes(ifNode.Else.(*ast.StmtElse).Stmt.(*ast.StmtStmtList).Stmts, cb.currBlock)
			if err != nil {
				log.Fatalf("Error in parseIf: %v", err)
			}
		}
		jmp := NewOpStmtJump(endBlock, ifNode.Position)
		cb.currBlock.AddInstructions(jmp)
		endBlock.AddPredecessor(cb.currBlock)
	}
}

func (cb *CfgBuilder) parseStmtGoto(stmt *ast.StmtGoto) {
	labelName, err := astutil.GetNameString(stmt.Label.(*ast.StmtLabel).Name)
	if err != nil {
		log.Fatalf("Error in StmtGoto: %v", err)
	}

	if labelBlock, ok := cb.Ctx.getLabel(labelName); ok {
		cb.currBlock.AddInstructions(NewOpStmtJump(labelBlock, stmt.Position))
		labelBlock.AddPredecessor(cb.currBlock)
	} else {
		cb.Ctx.addUnresolvedGoto(labelName, cb.currBlock)
	}
	cb.currBlock = NewBlock(cb.GetBlockId())
	cb.currBlock.Dead = true
}

func (cb *CfgBuilder) parseStmtLabel(stmt *ast.StmtLabel) {
	labelName, err := astutil.GetNameString(stmt.Name)
	if err != nil {
		log.Fatal("Error label name in StmtLabel")
	}
	if _, ok := cb.Ctx.getLabel(labelName); ok {
		log.Fatal("Error: label '", labelName, "' have been defined")
	}

	labelBlock := NewBlock(cb.GetBlockId())
	jmp := NewOpStmtJump(labelBlock, stmt.Position)
	cb.currBlock.AddInstructions(jmp)
	labelBlock.AddPredecessor(cb.currBlock)

	if unresolvedGotos, ok := cb.Ctx.getUnresolvedGotos(labelName); ok {
		for _, unresolvedGoto := range unresolvedGotos {
			jmp = NewOpStmtJump(labelBlock, nil)
			unresolvedGoto.AddInstructions(jmp)
			labelBlock.AddPredecessor(unresolvedGoto)
		}
		cb.Ctx.resolveGoto(labelName)
	}

	cb.Ctx.Labels[labelName] = labelBlock
	cb.currBlock = labelBlock
}

func (cb *CfgBuilder) parseStmtDo(stmt *ast.StmtDo) {
	var err error

	bodyBlock := NewBlock(cb.GetBlockId())
	bodyBlock.AddPredecessor(cb.currBlock)
	endBlock := NewBlock(cb.GetBlockId())
	cb.currBlock.AddInstructions(NewOpStmtJump(bodyBlock, stmt.Position))

	// parse statements in the loop body
	cb.currBlock = bodyBlock
	cb.currBlock, err = cb.parseStmtNodes(stmt.Stmt.(*ast.StmtStmtList).Stmts, bodyBlock)
	if err != nil {
		log.Fatalf("Error in parseStmtNodes: %v", err)
	}
	cond := cb.readVariable(cb.parseExprNode(stmt.Cond))
	cb.currBlock.AddInstructions(NewOpStmtJumpIf(cond, bodyBlock, endBlock, stmt.Cond.GetPosition()))
	cb.currBlock.IsConditional = true
	cb.processAssertion(cond, bodyBlock, endBlock)
	bodyBlock.AddPredecessor(cb.currBlock)
	endBlock.AddPredecessor(cb.currBlock)

	cb.currBlock = endBlock
}

func (cb *CfgBuilder) parseStmtFor(stmt *ast.StmtFor) {
	var err error

	cb.parseExprList(stmt.Init, MODE_READ)
	initBlock := NewBlock(cb.GetBlockId())
	bodyBlock := NewBlock(cb.GetBlockId())
	endBlock := NewBlock(cb.GetBlockId())

	// go to init block
	cb.currBlock.AddInstructions(NewOpStmtJump(initBlock, stmt.Position))
	initBlock.AddPredecessor(cb.currBlock)
	cb.currBlock = initBlock

	// check the condition
	cond := Operand(nil)
	if len(stmt.Cond) != 0 {
		vr := cb.parseExprList(stmt.Cond, MODE_NONE)
		cond = cb.readVariable(vr[len(vr)-1])
	} else {
		cond = NewOperBool(true)
	}
	cb.currBlock.AddInstructions(NewOpStmtJumpIf(cond, bodyBlock, endBlock, stmt.Position))
	cb.currBlock.IsConditional = true
	cb.processAssertion(cond, bodyBlock, endBlock)
	bodyBlock.AddPredecessor(cb.currBlock)
	endBlock.AddPredecessor(cb.currBlock)

	// parse statements inside loop body
	cb.currBlock, err = cb.parseStmtNodes(stmt.Stmt.(*ast.StmtStmtList).Stmts, bodyBlock)
	if err != nil {
		log.Fatalf("Error in parseStmtFor: %v", err)
	}
	cb.parseExprList(stmt.Loop, MODE_READ)
	// go back to init block
	cb.currBlock.AddInstructions(NewOpStmtJump(initBlock, stmt.Position))
	initBlock.AddPredecessor(cb.currBlock)
	cb.currBlock = endBlock
}

func (cb *CfgBuilder) parseStmtForeach(stmt *ast.StmtForeach) {
	var err error
	iterable := cb.readVariable(cb.parseExprNode(stmt.Expr))
	cb.currBlock.AddInstructions(NewOpReset(iterable, stmt.Expr.GetPosition()))

	initBlock := NewBlock(cb.GetBlockId())
	bodyBlock := NewBlock(cb.GetBlockId())
	endBlock := NewBlock(cb.GetBlockId())

	// go to init block
	cb.currBlock.AddInstructions(NewOpStmtJump(initBlock, stmt.Position))
	initBlock.AddPredecessor(cb.currBlock)

	// create valid iterator
	validOp := NewOpExprValid(iterable, nil)
	initBlock.AddInstructions(validOp)

	// go to body block
	cb.currBlock.AddInstructions(NewOpStmtJumpIf(validOp.Result, bodyBlock, endBlock, stmt.Position))
	cb.currBlock.IsConditional = true
	cb.processAssertion(validOp.Result, bodyBlock, endBlock)
	bodyBlock.AddPredecessor(cb.currBlock)
	endBlock.AddPredecessor(cb.currBlock)

	// parse body
	cb.currBlock = bodyBlock
	if stmt.Key != nil {
		keyOp := NewOpExprKey(iterable, nil)
		keyVar := cb.readVariable(cb.parseExprNode(stmt.Key))
		cb.currBlock.AddInstructions(keyOp)
		assignOp := NewOpExprAssign(keyVar, keyOp.Result, nil)
		cb.currBlock.AddInstructions(assignOp)
	}
	isRef := stmt.AmpersandTkn != nil
	valueOp := NewOpExprValue(iterable, isRef, nil)

	// assign each item to variable
	switch v := stmt.Var.(type) {
	case *ast.ExprList:
		cb.parseAssignList(v.Items, valueOp.Result, nil)
	case *ast.ExprArray:
		cb.parseAssignList(v.Items, valueOp.Result, nil)
	default:
		vr := cb.readVariable(cb.parseExprNode(stmt.Var))
		if isRef {
			cb.currBlock.AddInstructions(NewOpExprAssignRef(vr, valueOp.Result, nil))
		} else {
			cb.currBlock.AddInstructions(NewOpExprAssign(vr, valueOp.Result, nil))
		}
	}

	// parse statements inside loop body
	cb.currBlock, err = cb.parseStmtNodes(stmt.Stmt.(*ast.StmtStmtList).Stmts, cb.currBlock)
	if err != nil {
		log.Fatalf("Error in parseStmtForEach: %v", err)
	}
	cb.currBlock.AddInstructions(NewOpStmtJump(initBlock, stmt.Position))
	initBlock.AddPredecessor(cb.currBlock)

	cb.currBlock = endBlock
}

func (cb *CfgBuilder) parseStmtWhile(stmt *ast.StmtWhile) {
	var err error
	// initialize 3 block in while loop
	initBlock := NewBlock(cb.GetBlockId())
	bodyBlock := NewBlock(cb.GetBlockId())
	endBlock := NewBlock(cb.GetBlockId())

	// go to init block
	cb.currBlock.AddInstructions(NewOpStmtJump(initBlock, stmt.Position))
	initBlock.AddPredecessor(cb.currBlock)
	cb.currBlock = initBlock

	// create branch to body and end block
	cond := cb.readVariable(cb.parseExprNode(stmt.Cond))
	cb.currBlock.AddInstructions(NewOpStmtJumpIf(cond, bodyBlock, endBlock, stmt.Cond.GetPosition()))
	cb.currBlock.IsConditional = true
	bodyBlock.AddPredecessor(cb.currBlock)
	endBlock.AddPredecessor(cb.currBlock)

	// parse statements inside body loop
	cb.currBlock, err = cb.parseStmtNodes(stmt.Stmt.(*ast.StmtStmtList).Stmts, bodyBlock)
	if err != nil {
		log.Fatalf("Error in parseStmtWhile: %v", err)
	}

	// go back to init block
	cb.currBlock.AddInstructions(NewOpStmtJump(initBlock, stmt.Position))
	initBlock.AddPredecessor(cb.currBlock)

	cb.currBlock = endBlock
}

func (cb *CfgBuilder) parseStmtClass(stmt *ast.StmtClass) {
	name := cb.parseExprNode(stmt.Name)
	prevClass := cb.CurrClass
	cb.CurrClass = name.(*OperString)
	attrGroups := cb.parseAttrGroups(stmt.AttrGroups)
	stmts, err := cb.parseStmtNodes(stmt.Stmts, NewBlock(cb.GetBlockId()))
	if err != nil {
		log.Fatalf("Error in parseStmtClass: %v", err)
	}
	modifFlags := cb.parseClassModifier(stmt.Modifiers)
	extends := cb.parseExprNode(stmt.Extends)
	implements := cb.parseExprList(stmt.Implements, MODE_NONE)

	op := NewOpStmtClass(name, stmts, modifFlags, extends, implements, attrGroups, stmt.Position)
	cb.currBlock.AddInstructions(op)

	cb.CurrClass = prevClass
}

func (cb *CfgBuilder) parseStmtClassConstList(stmt *ast.StmtClassConstList) {
	if cb.CurrClass == nil {
		log.Fatal("Error: Unknown current class for a constants")
	}

	for _, c := range stmt.Consts {
		cb.parseStmtNode(c)
	}
}

func (cb *CfgBuilder) parseStmtClassMethod(stmt *ast.StmtClassMethod) {
	if cb.CurrClass == nil {
		log.Fatal("Error: Unknown current class for a method")
	}

	name, err := astutil.GetNameString(stmt.Name)
	if err != nil {
		log.Fatal("Error method name in StmtClassMethod")
	}
	flags := cb.parseFuncModifier(stmt.Modifiers, stmt.AmpersandTkn != nil)
	returnType := cb.parseTypeNode(stmt.ReturnType)
	entryBlock := NewBlock(cb.GetBlockId())
	fn, err := NewClassFunc(name, flags, returnType, *cb.CurrClass, entryBlock, stmt.Position)
	if err != nil {
		log.Fatalf("Error in parseStmtClassMethod: %v", err)
	}
	cb.Script.AddFunc(fn)

	// parse function
	cb.parseOpFunc(fn, stmt.Params, stmt.Stmt.(*ast.StmtStmtList).Stmts)

	// create method op
	visibility := fn.GetVisibility()
	static := fn.IsStatic()
	final := fn.IsFinal()
	abstract := fn.IsAbstract()
	attrs := cb.parseAttrGroups(stmt.AttrGroups)
	op := NewOpStmtClassMethod(fn, attrs, visibility, static, final, abstract, stmt.Position)
	cb.currBlock.AddInstructions(op)
	fn.CallableOp = op
}

func (cb *CfgBuilder) parseStmtInterface(stmt *ast.StmtInterface) {
	name := cb.readVariable(cb.parseExprNode(stmt.Name))
	tmpClass := cb.CurrClass
	cb.CurrClass = name.(*OperString)

	extends := cb.parseExprList(stmt.Extends, MODE_NONE)
	block, err := cb.parseStmtNodes(stmt.Stmts, cb.currBlock)
	if err != nil {
		log.Fatalf("Error in parseStmtInterface: %v", err)
	}
	op := NewOpStmtInterface(name, block, extends, stmt.Position)
	cb.currBlock.AddInstructions(op)

	cb.CurrClass = tmpClass
}

func (cb *CfgBuilder) parseStmtPropertyList(stmt *ast.StmtPropertyList) {
	attrGroups := cb.parseAttrGroups(stmt.AttrGroups)
	declaredType := cb.parseTypeNode(stmt.Type)
	visibility := ClassModifFlag(CLASS_MODIF_PUBLIC)
	static := false
	readonly := false
	// parse modifiers
	for _, modifier := range stmt.Modifiers {
		modifierStr, err := astutil.GetNameString(modifier)
		if err != nil {
			log.Fatal("Error modifier name in StmtPropertyList")
		}
		switch strings.ToLower(modifierStr) {
		case "public":
			visibility = CLASS_MODIF_PUBLIC
		case "protected":
			visibility = CLASS_MODIF_PROTECTED
		case "private":
			visibility = CLASS_MODIF_PRIVATE
		case "static":
			static = true
		case "readonly":
			readonly = true
		}
	}

	// parse each property
	for _, prop := range stmt.Props {
		defaultVar := Operand(nil)
		defaultBlock := (*Block)(nil)
		if prop.(*ast.StmtProperty).Expr != nil {
			defaultBlock = NewBlock(cb.GetBlockId())
			tmp := cb.currBlock
			cb.currBlock = defaultBlock
			defaultVar = cb.parseExprNode(prop.(*ast.StmtProperty).Expr)
			cb.currBlock = tmp
			if defaultVar.IsTainted() {
				cb.currFunc.ContaintTainted = true
				cb.currBlock.ContaintTainted = true
			}
		}

		name := cb.parseExprNode(prop.(*ast.StmtProperty).Var)
		op := NewOpStmtProperty(name, visibility, static, readonly, attrGroups, defaultVar, defaultBlock, declaredType, prop.GetPosition())
		cb.currBlock.AddInstructions(op)
	}
}

func (cb *CfgBuilder) parseStmtStatic(stmt *ast.StmtStatic) {
	for _, vr := range stmt.Vars {
		cb.parseStmtNode(vr)
	}
}

func (cb *CfgBuilder) parseStmtStaticVar(stmt *ast.StmtStaticVar) {
	defaultVar := Operand(nil)
	defaultBlock := (*Block)(nil)
	if stmt.Expr != nil {
		tmp := cb.currBlock
		defaultBlock = NewBlock(cb.GetBlockId())
		cb.currBlock = defaultBlock
		defaultVar = cb.parseExprNode(stmt.Expr)
		cb.currBlock = tmp
	}

	// TODO: add value to bound variable
	vr := cb.writeVariable(NewOperBoundVar(cb.parseExprNode(stmt.Var), NewOperNull(), true, VAR_SCOPE_FUNCTION, nil))
	cb.currBlock.AddInstructions(NewOpStaticVar(vr, defaultVar, defaultBlock, stmt.Position))
}

func (cb *CfgBuilder) parseStmtTrait(stmt *ast.StmtTrait) {
	name := cb.parseExprNode(stmt.Name)
	prevClass := cb.CurrClass
	cb.CurrClass = name.(*OperString)
	stmts, err := cb.parseStmtNodes(stmt.Stmts, NewBlock(cb.GetBlockId()))
	if err != nil {
		log.Fatalf("Error in parseStmtTrait: %v", err)
	}
	cb.currBlock.AddInstructions(NewOpStmtTrait(name, stmts, stmt.Position))
	cb.CurrClass = prevClass
}

func (cb *CfgBuilder) parseStmtTraitUse(stmt *ast.StmtTraitUse) {
	traits := make([]Operand, 0, len(stmt.Traits))
	adaptations := make([]Op, 0, len(stmt.Adaptations))

	for _, trait := range stmt.Traits {
		traitName, err := astutil.GetNameString(trait)
		if err != nil {
			log.Fatal("Error trait name in StmtTraitUse")
		}
		traits = append(traits, NewOperString(traitName))
	}

	for _, adaptation := range stmt.Adaptations {
		switch a := adaptation.(type) {
		case *ast.StmtTraitUseAlias:
			trait := Operand(nil)
			methodStr, err := astutil.GetNameString(a.Method)
			if err != nil {
				log.Fatal("Error method name in StmtTraitUse")
			}
			method := NewOperString(methodStr)
			newName := Operand(nil)
			newModifier := cb.parseClassModifier([]ast.Vertex{a.Modifier})
			if a.Trait != nil {
				traitStr, err := astutil.GetNameString(a.Trait)
				if err != nil {
					log.Fatal("Error trait name in StmtTraitUse")
				}
				trait = NewOperString(traitStr)
			}
			if a.Alias != nil {
				aliasStr, err := astutil.GetNameString(a.Alias)
				if err != nil {
					log.Fatal("Error alias name in StmtTraitUse")
				}
				newName = NewOperString(aliasStr)
			}

			aliasOp := NewOpAlias(trait, method, newName, newModifier, a.Position)
			adaptations = append(adaptations, aliasOp)
		case *ast.StmtTraitUsePrecedence:
			insteadOfs := make([]Operand, 0, len(a.Insteadof))
			trait := Operand(nil)
			methodStr, err := astutil.GetNameString(a.Method)
			if err != nil {
				log.Fatal("Error method name in StmtTraitUsePrecedence")
			}
			method := NewOperString(methodStr)
			if a.Trait != nil {
				traitStr, err := astutil.GetNameString(a.Trait)
				if err != nil {
					log.Fatal("Error trait name in StmtTraitUsePrecedence")
				}
				trait = NewOperString(traitStr)
			}
			for _, insteadOf := range a.Insteadof {
				insteadOfStr, err := astutil.GetNameString(insteadOf)
				if err != nil {
					log.Fatalf("Error insteadof in StmtTraitUsePrecedence")
				}
				insteadOfName := NewOperString(insteadOfStr)
				insteadOfs = append(insteadOfs, insteadOfName)
			}

			precedenceOp := NewOpPrecedence(trait, method, insteadOfs, a.Position)
			adaptations = append(adaptations, precedenceOp)
		}
	}
	traitUseOp := NewOpStmtTraitUse(traits, adaptations, stmt.Position)
	cb.currBlock.AddInstructions(traitUseOp)
}

func (cb *CfgBuilder) parseStmtFunction(stmt *ast.StmtFunction) {
	// create OpFunc instance and append to script object
	name, err := astutil.GetNameString(stmt.Name)
	if err != nil {
		log.Fatal("Error func name in StmtFunction")
	}
	flags := FuncModifFlag(0)
	returnType := cb.parseTypeNode(stmt.ReturnType)
	if stmt.AmpersandTkn != nil {
		flags |= FUNC_MODIF_RETURNS_REF
	}
	entryBlock := NewBlock(cb.GetBlockId())
	fn, err := NewFunc(name, flags, returnType, entryBlock, stmt.Position)
	if err != nil {
		log.Fatalf("Error in parseStmtFunction: %v", err)
	}
	cb.Script.AddFunc(fn)

	// parse function
	cb.parseOpFunc(fn, stmt.Params, stmt.Stmts)
	attrGroups := cb.parseAttrGroups(stmt.AttrGroups)
	opStmtFunc := NewOpStmtFunc(fn, attrGroups, stmt.Position)
	cb.currBlock.AddInstructions(opStmtFunc)
	fn.CallableOp = opStmtFunc
}

func (cb *CfgBuilder) parseStmtNamespace(stmt *ast.StmtNamespace) {
	nameSpace, err := astutil.GetNameString(stmt.Name)
	if err != nil {
		log.Fatal("Error namespace in StmtNameSpace")
	}
	cb.CurrNamespace = nameSpace
	cb.currBlock, err = cb.parseStmtNodes(stmt.Stmts, cb.currBlock)
	if err != nil {
		log.Fatalf("Error in StmtNameSpace: %v", err)
	}
}

func (cb *CfgBuilder) parseClassModifier(modifiers []ast.Vertex) ClassModifFlag {
	flags := ClassModifFlag(0)
	for _, modifier := range modifiers {
		switch strings.ToLower(string(modifier.(*ast.Identifier).Value)) {
		case "public":
			flags |= CLASS_MODIF_PUBLIC
		case "protected":
			flags |= CLASS_MODIF_PROTECTED
		case "private":
			flags |= CLASS_MODIF_PRIVATE
		case "static":
			flags |= CLASS_MODIF_STATIC
		case "abstract":
			flags |= CLASS_MODIF_ABSTRACT
		case "final":
			flags |= CLASS_MODIF_FINAL
		case "readonly":
			flags |= CLASS_MODIF_READONLY
		default:
			log.Fatal("Error: Unknown Identifier '", string(modifier.(*ast.Identifier).Value), "'")
		}
	}

	return flags
}

func (cb *CfgBuilder) parseFuncModifier(modifiers []ast.Vertex, isRef bool) FuncModifFlag {
	flags := FuncModifFlag(0)

	if isRef {
		flags |= FUNC_MODIF_RETURNS_REF
	}

	for _, modifier := range modifiers {
		switch strings.ToLower(string(modifier.(*ast.Identifier).Value)) {
		case "public":
			flags |= FUNC_MODIF_PUBLIC
		case "protected":
			flags |= FUNC_MODIF_PROTECTED
		case "private":
			flags |= FUNC_MODIF_PRIVATE
		case "static":
			flags |= FUNC_MODIF_STATIC
		case "abstract":
			flags |= FUNC_MODIF_ABSTRACT
		case "final":
			flags |= FUNC_MODIF_FINAL
		default:
			log.Fatal("Error: Unknown Identifier '", string(modifier.(*ast.Identifier).Value), "'")
		}
	}

	return flags
}

func (cb *CfgBuilder) parseAttrGroup(attrGroup *ast.AttributeGroup) *OpAttributeGroup {
	// parse each attribute
	attrs := make([]*OpAttribute, 0)
	for _, attrNode := range attrGroup.Attrs {
		args := make([]Operand, 0)
		for _, argNode := range attrNode.(*ast.Attribute).Args {
			arg := cb.readVariable(cb.parseExprNode(argNode.(*ast.Argument).Expr))
			args = append(args, arg)
		}
		attrName := cb.readVariable(cb.parseExprNode(attrNode.(*ast.Attribute).Name))
		attr := NewOpAttribute(attrName, args, attrNode.GetPosition())
		attrs = append(attrs, attr)
	}

	return NewOpAttributeGroup(attrs, attrGroup.Position)
}

func (cb *CfgBuilder) parseAttrGroups(attrGroupNodes []ast.Vertex) []*OpAttributeGroup {
	attrGroups := make([]*OpAttributeGroup, 0)
	for _, attrGroupNode := range attrGroupNodes {
		attrGroup := cb.parseAttrGroup(attrGroupNode.(*ast.AttributeGroup))
		attrGroups = append(attrGroups, attrGroup)
	}

	return attrGroups
}

func (cb *CfgBuilder) parseExprNode(expr ast.Vertex) Operand {
	if expr == nil {
		return nil
	}

	switch exprT := expr.(type) {
	case *ast.ExprVariable:
		nameStr, err := astutil.GetNameString(exprT.Name)
		if err != nil && nameStr == "this" {
			// TODO: add value to bound variable
			return NewOperBoundVar(
				cb.parseExprNode(exprT.Name),
				NewOperNull(),
				false,
				VAR_SCOPE_OBJECT,
				cb.CurrClass,
			)
		}
		name := cb.parseExprNode(exprT.Name)
		return NewOperVar(name, nil)
	case *ast.Name, *ast.NameFullyQualified, *ast.NameRelative, *ast.Identifier:
		nameStr, _ := astutil.GetNameString(exprT)
		return NewOperString(nameStr)
	case *ast.ScalarDnumber:
		num, err := strconv.Atoi(string(exprT.Value))
		if err != nil {
			log.Fatal(err)
		}
		return NewOperNumber(float64(num))
	case *ast.ScalarLnumber:
		num, err := strconv.ParseFloat(string(exprT.Value), 64)
		if err != nil {
			log.Fatal(err)
		}
		return NewOperNumber(num)
	case *ast.ScalarString:
		str := string(exprT.Value)
		return NewOperString(str)
	case *ast.ScalarEncapsed:
		parts := cb.parseExprList(exprT.Parts, MODE_READ)
		op := NewOpExprConcatList(parts, exprT.Position)
		cb.currBlock.Instructions = append(cb.currBlock.Instructions, op)
		return op.Result
	case *ast.ScalarEncapsedStringBrackets:
		return cb.parseExprNode(exprT.Var)
	case *ast.ScalarEncapsedStringPart:
		str := string(exprT.Value)
		return NewOperString(str)
	case *ast.Argument:
		vr := cb.readVariable(cb.parseExprNode(exprT.Expr))
		return vr
	case *ast.ExprAssign:
		return cb.parseExprAssign(exprT)
	case *ast.ExprAssignBitwiseAnd, *ast.ExprAssignBitwiseOr, *ast.ExprAssignBitwiseXor,
		*ast.ExprAssignCoalesce, *ast.ExprAssignConcat, *ast.ExprAssignDiv, *ast.ExprAssignMinus,
		*ast.ExprAssignMod, *ast.ExprAssignMul, *ast.ExprAssignPlus, *ast.ExprAssignPow,
		*ast.ExprAssignShiftLeft, *ast.ExprAssignShiftRight:
		return cb.parseExprAssignOp(exprT)
	case *ast.ExprAssignReference:
		return cb.parseExprAssignRef(exprT)
	// Not use short circuiting to simplify the taint analysis
	// case *ast.ExprBinaryBooleanAnd, *ast.ExprBinaryLogicalAnd:
	// 	// TODO: finish this
	// 	return cb.parseShortCircuiting(exprT, false)
	// case *ast.ExprBinaryBooleanOr, *ast.ExprBinaryLogicalOr:
	// 	// TODO: finish this
	// 	return cb.parseShortCircuiting(exprT, true)
	case *ast.ExprBinaryBitwiseAnd, *ast.ExprBinaryBitwiseOr, *ast.ExprBinaryBitwiseXor, *ast.ExprBinaryNotEqual,
		*ast.ExprBinaryCoalesce, *ast.ExprBinaryConcat, *ast.ExprBinaryDiv, *ast.ExprBinaryEqual, *ast.ExprBinaryGreater,
		*ast.ExprBinaryGreaterOrEqual, *ast.ExprBinaryIdentical, *ast.ExprBinaryLogicalXor, *ast.ExprBinaryMinus,
		*ast.ExprBinaryMod, *ast.ExprBinaryMul, *ast.ExprBinaryNotIdentical, *ast.ExprBinaryPlus, *ast.ExprBinaryPow,
		*ast.ExprBinaryShiftLeft, *ast.ExprBinaryShiftRight, *ast.ExprBinarySmaller, *ast.ExprBinarySmallerOrEqual,
		*ast.ExprBinarySpaceship, *ast.ExprBinaryBooleanAnd, *ast.ExprBinaryLogicalAnd, *ast.ExprBinaryBooleanOr, *ast.ExprBinaryLogicalOr:
		return cb.parseBinaryExprNode(exprT)
	case *ast.ExprCastArray:
		vr := cb.parseExprNode(exprT.Expr)
		e := cb.readVariable(vr)
		op := NewOpExprCastArray(e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastBool:
		vr := cb.parseExprNode(exprT.Expr)
		e := cb.readVariable(vr)
		op := NewOpExprCastBool(e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastDouble:
		vr := cb.parseExprNode(exprT.Expr)
		e := cb.readVariable(vr)
		op := NewOpExprCastDouble(e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastInt:
		vr := cb.parseExprNode(exprT.Expr)
		e := cb.readVariable(vr)
		op := NewOpExprCastInt(e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastObject:
		vr := cb.parseExprNode(exprT.Expr)
		e := cb.readVariable(vr)
		op := NewOpExprCastObject(e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastString:
		vr := cb.parseExprNode(exprT.Expr)
		e := cb.readVariable(vr)
		op := NewOpExprCastString(e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastUnset:
		vr := cb.parseExprNode(exprT.Expr)
		e := cb.readVariable(vr)
		op := NewOpExprCastUnset(e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprUnaryMinus:
		vr := cb.parseExprNode(exprT.Expr)
		val := cb.readVariable(vr)
		op := NewOpExprUnaryMinus(val, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprUnaryPlus:
		vr := cb.parseExprNode(exprT.Expr)
		val := cb.readVariable(vr)
		op := NewOpExprUnaryPlus(val, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprArray:
		return cb.parseExprArray(exprT)
	case *ast.ExprArrayDimFetch:
		return cb.parseExprArrayDimFetch(exprT)
	case *ast.ExprBitwiseNot:
		oper := cb.readVariable(cb.parseExprNode(exprT.Expr))
		op := NewOpExprBitwiseNot(oper, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBooleanNot:
		cond := cb.readVariable(cb.parseExprNode(exprT.Expr))
		op := NewOpExprBooleanNot(cond, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprClosure:
		return cb.parseExprClosure(exprT)
	case *ast.ExprClassConstFetch:
		class := cb.readVariable(cb.parseExprNode(exprT.Class))
		name := cb.readVariable(cb.parseExprNode(exprT.Class))
		op := NewOpExprClassConstFetch(class, name, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprClone:
		clone := cb.readVariable(cb.parseExprNode(exprT.Expr))
		op := NewOpExprClone(clone, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprConstFetch:
		return cb.parseExprConstFetch(exprT)
	case *ast.ExprEmpty:
		empty := cb.readVariable(cb.parseExprNode(exprT.Expr))
		op := NewOpExprEmpty(empty, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprErrorSuppress:
		return cb.parseExprErrorSuppress(exprT)
	case *ast.ExprEval:
		eval := cb.readVariable(cb.parseExprNode(exprT.Expr))
		op := NewOpExprEval(eval, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprExit:
		return cb.parseExprExit(exprT)
	case *ast.ExprFunctionCall:
		return cb.parseExprFuncCall(exprT)
	case *ast.ExprInclude:
		include := cb.readVariable(cb.parseExprNode(exprT.Expr))
		// add to include file
		if includeStr, ok := include.(*OperString); ok {
			cb.Script.IncludedFiles = append(cb.Script.IncludedFiles, includeStr.Val)
		}
		op := NewOpExprInclude(include, TYPE_INCLUDE, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprIncludeOnce:
		include := cb.readVariable(cb.parseExprNode(exprT.Expr))
		// add to include file
		if includeStr, ok := include.(*OperString); ok {
			cb.Script.IncludedFiles = append(cb.Script.IncludedFiles, includeStr.Val)
		}
		op := NewOpExprInclude(include, TYPE_INCLUDE_ONCE, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprRequire:
		include := cb.readVariable(cb.parseExprNode(exprT.Expr))
		// add to include file
		if includeStr, ok := include.(*OperString); ok {
			cb.Script.IncludedFiles = append(cb.Script.IncludedFiles, includeStr.Val)
		}
		op := NewOpExprInclude(include, TYPE_REQUIRE, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprRequireOnce:
		include := cb.readVariable(cb.parseExprNode(exprT.Expr))
		// add to include file
		if includeStr, ok := include.(*OperString); ok {
			cb.Script.IncludedFiles = append(cb.Script.IncludedFiles, includeStr.Val)
		}
		op := NewOpExprInclude(include, TYPE_REQUIRE_ONCE, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprInstanceOf:
		vr := cb.readVariable(cb.parseExprNode(exprT.Expr))
		class := cb.readVariable(cb.parseExprNode(exprT.Class))
		op := NewOpExprInstanceOf(vr, class, exprT.Position)
		op.Result.AddAssertion(vr, NewTypeAssertion(class, false), ASSERT_MODE_INTERSECT)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprIsset:
		isset := cb.parseExprList(exprT.Vars, MODE_READ)
		op := NewOpExprIsset(isset, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprMethodCall:
		vr := cb.readVariable(cb.parseExprNode(exprT.Var))
		name := cb.readVariable(cb.parseExprNode(exprT.Method))
		args := cb.parseExprList(exprT.Args, MODE_READ)
		op := NewOpExprMethodCall(vr, name, args, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprNullsafeMethodCall:
		vr := cb.readVariable(cb.parseExprNode(exprT.Var))
		name := cb.readVariable(cb.parseExprNode(exprT.Method))
		args := cb.parseExprList(exprT.Args, MODE_READ)
		op := NewOpExprNullSafeMethodCall(vr, name, args, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprNew:
		return cb.parseExprNew(exprT)
	case *ast.ExprPostDec:
		vr := cb.parseExprNode(exprT.Var)
		read := cb.readVariable(vr)
		write := cb.writeVariable(vr)
		opMinus := NewOpExprBinaryMinus(read, NewOperNumber(1), exprT.Position)
		opAssign := NewOpExprAssign(write, opMinus.Result, exprT.Position)
		cb.currBlock.AddInstructions(opMinus)
		cb.currBlock.AddInstructions(opAssign)
		return read
	case *ast.ExprPostInc:
		vr := cb.parseExprNode(exprT.Var)
		read := cb.readVariable(vr)
		write := cb.writeVariable(vr)
		opPlus := NewOpExprBinaryPlus(read, NewOperNumber(1), exprT.Position)
		opAssign := NewOpExprAssign(write, opPlus.Result, exprT.Position)
		cb.currBlock.AddInstructions(opPlus)
		cb.currBlock.AddInstructions(opAssign)
		return read
	case *ast.ExprPreDec:
		vr := cb.parseExprNode(exprT.Var)
		read := cb.readVariable(vr)
		write := cb.writeVariable(vr)
		opMinus := NewOpExprBinaryMinus(read, NewOperNumber(1), exprT.Position)
		opAssign := NewOpExprAssign(write, opMinus.Result, exprT.Position)
		cb.currBlock.AddInstructions(opMinus)
		cb.currBlock.AddInstructions(opAssign)
		return opMinus.Result
	case *ast.ExprPreInc:
		vr := cb.parseExprNode(exprT.Var)
		read := cb.readVariable(vr)
		write := cb.writeVariable(vr)
		opPlus := NewOpExprBinaryPlus(read, NewOperNumber(1), exprT.Position)
		opAssign := NewOpExprAssign(write, opPlus.Result, exprT.Position)
		cb.currBlock.AddInstructions(opPlus)
		cb.currBlock.AddInstructions(opAssign)
		return opPlus.Result
	case *ast.ExprPrint:
		print := cb.readVariable(cb.parseExprNode(exprT.Expr))
		op := NewOpExprPrint(print, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprPropertyFetch:
		vr := cb.readVariable(cb.parseExprNode(exprT.Var))
		name := cb.readVariable(cb.parseExprNode(exprT.Prop))
		op := NewOpExprPropertyFetch(vr, name, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprNullsafePropertyFetch:
		vr := cb.readVariable(cb.parseExprNode(exprT.Var))
		name := cb.readVariable(cb.parseExprNode(exprT.Prop))
		op := NewOpExprPropertyFetch(vr, name, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprStaticCall:
		class := cb.readVariable(cb.parseExprNode(exprT.Class))
		name := cb.readVariable(cb.parseExprNode(exprT.Call))
		args := cb.parseExprList(exprT.Args, MODE_READ)
		op := NewOpExprStaticCall(class, name, args, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprStaticPropertyFetch:
		class := cb.readVariable(cb.parseExprNode(exprT.Class))
		name := cb.readVariable(cb.parseExprNode(exprT.Prop))
		op := NewOpExprStaticPropertyFetch(class, name, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprTernary:
		return cb.parseExprTernary(exprT)
	case *ast.ExprYield:
		return cb.parseExprYield(exprT)
	case *ast.ExprShellExec:
		args := cb.parseExprList(exprT.Parts, MODE_READ)
		argOp := NewOpExprConcatList(args, exprT.Position)
		cb.currBlock.AddInstructions(argOp)
		funcCallOp := NewOpExprFunctionCall(NewOperString("shell_exec"), []Operand{argOp.Result}, exprT.Position)
		cb.currBlock.AddInstructions(funcCallOp)
		return argOp.Result
	case *ast.ExprBrackets:
		return cb.parseExprNode(exprT.Expr)
	case *ast.ExprClosureUse:
		// parsed in parseExprClosure function
		log.Fatal("Error: ExprClosureUse parsed in parseExprClosure function")
	case *ast.ExprList:
		log.Fatal("Error: ExprList parsed in parseExprAssign function")
	case *ast.ExprArrayItem:
		log.Fatal("Error: ExprArrayItem parsed in parseExprArray function")
	case *ast.ExprArrowFunction, *ast.ExprMatch, *ast.ExprYieldFrom, *ast.ExprThrow:
		// TODO
		log.Fatal("Error: Cannot parse expression node, wrong type '", reflect.TypeOf(expr), "'")
	default:
		fmt.Println(expr.(*ast.Attribute))
		log.Fatal("Error: Cannot parse expression node, wrong type '", reflect.TypeOf(expr), "'")
	}

	return nil
}

// function to parse ast.ExprAssign into OpExprAssign
// return right side of assignment operation
func (cb *CfgBuilder) parseExprAssign(expr *ast.ExprAssign) Operand {
	var right Operand = cb.readVariable(cb.parseExprNode(expr.Expr))

	// if var a list or array, do list assignment
	switch e := expr.Var.(type) {
	case *ast.ExprList:
		cb.parseAssignList(e.Items, right, e.Position)
		return right
	case *ast.ExprArray:
		cb.parseAssignList(e.Items, right, e.Position)
		return right
	}

	left := cb.writeVariable(cb.parseExprNode(expr.Var))
	op := NewOpExprAssign(left, right, expr.Position)
	cb.currBlock.AddInstructions(op)

	// if right expr is a literal or object
	switch rv := GetOperVal(right).(type) {
	case *OperBool, *OperObject, *OperString, *OperSymbolic, *OperNumber:
		SetOperVal(op.Result, rv)
		SetOperVal(left, rv)
	}

	return op.Result
}

func (cb *CfgBuilder) parseAssignList(items []ast.Vertex, arrVar Operand, pos *position.Position) {
	cnt := 0
	for _, item := range items {
		if item == nil {
			continue
		}
		var key Operand = nil
		arrItem := item.(*ast.ExprArrayItem)

		// if no key, set key to cnt (considered as array)
		if arrItem.Key == nil {
			key = NewOperNumber(float64(cnt))
			cnt += 1
		} else {
			key = cb.readVariable(cb.parseExprNode(arrItem.Key))
		}

		// set array's item value
		vr := arrItem.Val
		fetch := NewOpExprArrayDimFetch(arrVar, key, pos)
		cb.currBlock.AddInstructions(fetch)

		// assign recursively
		switch e := vr.(type) {
		case *ast.ExprList:
			cb.parseAssignList(e.Items, fetch.Result, e.Position)
			continue
		case *ast.ExprArray:
			cb.parseAssignList(e.Items, fetch.Result, e.Position)
			continue
		}

		// assign item with corresponding value
		left := cb.writeVariable(cb.parseExprNode(vr))
		assign := NewOpExprAssign(left, fetch.Result, pos)
		cb.currBlock.AddInstructions(assign)
	}
}

// function to parse ast.ExprAssignReference into OpExprAssignRef
// return right side of assignment operation
func (cb *CfgBuilder) parseExprAssignRef(expr *ast.ExprAssignReference) Operand {
	left := cb.writeVariable(cb.parseExprNode(expr.Var))
	right := cb.readVariable(cb.parseExprNode(expr.Expr))

	assign := NewOpExprAssignRef(left, right, expr.Position)
	return assign.Result
}

// function to parse other ast.ExprAssign... into OpExprAssign...
// return right side of assignment operation
func (cb *CfgBuilder) parseExprAssignOp(expr ast.Vertex) Operand {
	var vr, e Operand
	var read, write Operand
	switch exprT := expr.(type) {
	case *ast.ExprAssignBitwiseAnd:
		vr = cb.parseExprNode(exprT.Var)
		read = cb.readVariable(vr)
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryBitwiseAnd(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Position)
		cb.currBlock.AddInstructions(assign)
	case *ast.ExprAssignBitwiseOr:
		vr = cb.parseExprNode(exprT.Var)
		read = cb.readVariable(vr)
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryBitwiseOr(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Position)
		cb.currBlock.AddInstructions(assign)
	case *ast.ExprAssignBitwiseXor:
		vr = cb.parseExprNode(exprT.Var)
		read = cb.readVariable(vr)
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryBitwiseXor(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Position)
		cb.currBlock.AddInstructions(assign)
	case *ast.ExprAssignConcat:
		vr = cb.parseExprNode(exprT.Var)
		read = cb.readVariable(vr)
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryConcat(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Position)
		cb.currBlock.AddInstructions(assign)
	case *ast.ExprAssignCoalesce:
		vr = cb.parseExprNode(exprT.Var)
		read = cb.readVariable(vr)
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryCoalesce(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Position)
		cb.currBlock.AddInstructions(assign)
	case *ast.ExprAssignDiv:
		vr = cb.parseExprNode(exprT.Var)
		read = cb.readVariable(vr)
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryDiv(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Position)
		cb.currBlock.AddInstructions(assign)
	case *ast.ExprAssignMinus:
		vr = cb.parseExprNode(exprT.Var)
		read = cb.readVariable(vr)
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryMinus(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Position)
		cb.currBlock.AddInstructions(assign)
	case *ast.ExprAssignMod:
		vr = cb.parseExprNode(exprT.Var)
		read = cb.readVariable(vr)
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryMod(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Position)
		cb.currBlock.AddInstructions(assign)
	case *ast.ExprAssignPlus:
		vr = cb.parseExprNode(exprT.Var)
		read = cb.readVariable(vr)
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryPlus(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Position)
		cb.currBlock.AddInstructions(assign)
	case *ast.ExprAssignMul:
		vr = cb.parseExprNode(exprT.Var)
		read = cb.readVariable(vr)
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryMul(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Position)
		cb.currBlock.AddInstructions(assign)
	case *ast.ExprAssignPow:
		vr = cb.parseExprNode(exprT.Var)
		read = cb.readVariable(vr)
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryPow(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Position)
		cb.currBlock.AddInstructions(assign)
	case *ast.ExprAssignShiftLeft:
		vr = cb.parseExprNode(exprT.Var)
		read = cb.readVariable(vr)
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryShiftLeft(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Position)
		cb.currBlock.AddInstructions(assign)
	case *ast.ExprAssignShiftRight:
		vr = cb.parseExprNode(exprT.Var)
		read = cb.readVariable(vr)
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryShiftRight(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Position)
		cb.currBlock.AddInstructions(assign)
	default:
		log.Fatal("Error: invalid assignment operator expression")
	}

	return nil
}

func (cb *CfgBuilder) parseExprClosure(expr *ast.ExprClosure) Operand {
	// get each closure's uses
	uses := make([]Operand, len(expr.Uses))
	for i, exprUse := range expr.Uses {
		eu := exprUse.(*ast.ExprClosureUse)
		nameVar := cb.readVariable(cb.parseExprNode(eu.Var))
		useByRef := eu.AmpersandTkn != nil
		// TODO: add value to bound variable
		uses[i] = NewOperBoundVar(nameVar, NewOperNull(), useByRef, VAR_SCOPE_LOCAL, nil)
	}

	// Create opFunction
	byRef := expr.AmpersandTkn != nil
	isStatic := expr.StaticTkn != nil
	name := fmt.Sprintf("{anonymous}#%d", cb.getAnonId())
	types := cb.parseTypeNode(expr.ReturnType)
	entryBlock := NewBlock(cb.GetBlockId())
	opFunc, err := NewFunc(name, FUNC_MODIF_CLOSURE, types, entryBlock, expr.Position)
	if err != nil {
		log.Fatalf("Error in parseExprClosure: %v", err)
	}
	if byRef {
		opFunc.AddModifier(FUNC_MODIF_RETURNS_REF)
	}
	if isStatic {
		opFunc.AddModifier(FUNC_MODIF_STATIC)
	}
	cb.currBlock.AddInstructions(opFunc)

	// build cfg for the closure
	cb.parseOpFunc(opFunc, expr.Params, expr.Stmts)

	// create op closure
	closure := NewOpExprClosure(opFunc, uses, expr.Position)
	opFunc.CallableOp = closure

	cb.currBlock.AddInstructions(closure)
	return closure.Result
}

func (cb *CfgBuilder) parseExprConstFetch(expr *ast.ExprConstFetch) Operand {
	nameStr, err := astutil.GetNameString(expr.Const)
	if err != nil {
		log.Fatal("Error const name in ExprConstFetch")
	}
	lowerName := strings.ToLower(nameStr)
	switch lowerName {
	case "null":
		return NewOperNull()
	case "true":
		return NewOperBool(true)
	case "false":
		return NewOperBool(false)
	}

	// TODO: Check again
	name := cb.parseExprNode(expr.Const)
	op := NewOpExprConstFetch(name, expr.Position)
	cb.currBlock.AddInstructions(op)

	// find the constant definition
	if val, ok := cb.Consts[nameStr]; ok {
		op.Result = val
	}

	return op.Result
}

// function to parse ast.ExprBinary... into OpExprBinary
func (cb *CfgBuilder) parseBinaryExprNode(expr ast.Vertex) Operand {
	switch e := expr.(type) {
	case *ast.ExprBinaryBitwiseAnd:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryBitwiseAnd(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryBitwiseOr:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryBitwiseOr(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryBitwiseXor:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryBitwiseXor(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryNotEqual:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryNotEqual(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryCoalesce:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryCoalesce(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryConcat:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryConcat(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryDiv:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryDiv(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryEqual:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryEqual(left, right, e.Position)

		// handle type assertion using gettype()
		if left.IsWritten() {
			leftOp, isLeftFuncCall := left.GetWriteOp()[0].(*OpExprFunctionCall)
			rightStr, isRightString := GetStringOper(right)
			// left must be function call with name gettype
			// right must be a string
			if isLeftFuncCall && GetOperName(leftOp.Name) == "gettype" && isRightString {
				switch strings.Trim(rightStr, "\"") {
				case "integer":
					assert := NewTypeAssertion(NewOperString("int"), false)
					op.Result.AddAssertion(leftOp.Args[0], assert, ASSERT_MODE_INTERSECT)
				case "float":
					assert := NewTypeAssertion(NewOperString("float"), false)
					op.Result.AddAssertion(leftOp.Args[0], assert, ASSERT_MODE_INTERSECT)
				case "double":
					assert := NewTypeAssertion(NewOperString("float"), false)
					op.Result.AddAssertion(leftOp.Args[0], assert, ASSERT_MODE_INTERSECT)
				case "boolean":
					assert := NewTypeAssertion(NewOperString("bool"), false)
					op.Result.AddAssertion(leftOp.Args[0], assert, ASSERT_MODE_INTERSECT)
				case "NULL":
					assert := NewTypeAssertion(NewOperString("null"), false)
					op.Result.AddAssertion(leftOp.Args[0], assert, ASSERT_MODE_INTERSECT)
				}
			}
		}

		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryGreater:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryGreater(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryGreaterOrEqual:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryGreaterOrEqual(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryIdentical:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryIdentical(left, right, e.Position)

		// handle type assertion using gettype()
		if left.IsWritten() {
			leftOp, isLeftFuncCall := left.GetWriteOp()[0].(*OpExprFunctionCall)
			rightStr, isRightString := GetStringOper(right)
			// left must be function call with name gettype
			// right must be a string
			if isLeftFuncCall && GetOperName(leftOp.Name) == "gettype" && isRightString {
				switch strings.Trim(rightStr, "\"") {
				case "integer":
					assert := NewTypeAssertion(NewOperString("int"), false)
					op.Result.AddAssertion(leftOp.Args[0], assert, ASSERT_MODE_INTERSECT)
				case "float":
					assert := NewTypeAssertion(NewOperString("float"), false)
					op.Result.AddAssertion(leftOp.Args[0], assert, ASSERT_MODE_INTERSECT)
				case "double":
					assert := NewTypeAssertion(NewOperString("float"), false)
					op.Result.AddAssertion(leftOp.Args[0], assert, ASSERT_MODE_INTERSECT)
				case "boolean":
					assert := NewTypeAssertion(NewOperString("bool"), false)
					op.Result.AddAssertion(leftOp.Args[0], assert, ASSERT_MODE_INTERSECT)
				case "NULL":
					assert := NewTypeAssertion(NewOperString("null"), false)
					op.Result.AddAssertion(leftOp.Args[0], assert, ASSERT_MODE_INTERSECT)
				}
			}
		}

		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryLogicalOr:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryLogicalOr(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryBooleanOr:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryLogicalOr(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryLogicalAnd:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryLogicalAnd(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryBooleanAnd:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryLogicalAnd(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryLogicalXor:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryLogicalXor(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryMinus:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryMinus(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryMod:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryMod(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryMul:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryMul(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryNotIdentical:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryNotIdentical(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryPlus:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryPlus(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryPow:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryPow(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryShiftLeft:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryShiftLeft(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryShiftRight:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinaryShiftRight(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinarySmaller:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinarySmaller(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinarySmallerOrEqual:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinarySmallerOrEqual(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinarySpaceship:
		left := cb.readVariable(cb.parseExprNode(e.Left))
		right := cb.readVariable(cb.parseExprNode(e.Right))
		op := NewOpExprBinarySpaceship(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	default:
		log.Fatalf("Error: Invalid binary op node '%v'", reflect.TypeOf(e))
	}
	return nil
}

// function to parse ast.ExprArray into OpExprArray
func (cb *CfgBuilder) parseExprArray(expr *ast.ExprArray) Operand {
	keys := make([]Operand, 0)
	vals := make([]Operand, 0)
	byRefs := make([]bool, 0)

	if expr.Items != nil {
		for _, arrItem := range expr.Items {
			item := arrItem.(*ast.ExprArrayItem)
			if item.Key != nil {
				key := cb.readVariable(cb.parseExprNode(item.Key))
				keys = append(keys, key)
			} else {
				keys = append(keys, NewOperNull())
			}

			val := cb.readVariable(cb.parseExprNode(item.Val))
			vals = append(vals, val)

			if item.AmpersandTkn != nil {
				byRefs = append(byRefs, true)
			} else {
				byRefs = append(byRefs, false)
			}
		}
	}

	op := NewOpExprArray(keys, vals, byRefs, expr.Position)
	cb.currBlock.AddInstructions(op)

	return op.Result
}

// function to parse ast.ExprArrayDimFetch into OpExprArrayDimFetch
func (cb *CfgBuilder) parseExprArrayDimFetch(expr *ast.ExprArrayDimFetch) Operand {
	vr := cb.readVariable(cb.parseExprNode(expr.Var))
	var dim Operand
	if expr.Dim != nil {
		dim = cb.readVariable(cb.parseExprNode(expr.Dim))
	} else {
		dim = NewOperNull()
	}

	op := NewOpExprArrayDimFetch(vr, dim, expr.Position)
	cb.currBlock.AddInstructions(op)

	// symbolic interpreter
	if vs, ok := vr.(*OperSymbolic); ok && vs.Val != "undefined" {
		op.Result = NewOperSymbolic(vs.Val, true)
	}

	// if op.Result.IsTainted() {
	// 	log.Fatal("")
	// }

	return op.Result
}

func (cb *CfgBuilder) parseExprErrorSuppress(expr *ast.ExprErrorSuppress) Operand {
	// create new error supress block
	errSupressBlock := NewBlock(cb.GetBlockId())
	// add instruction to jump into error supress block
	jmp := NewOpStmtJump(errSupressBlock, expr.Position)
	cb.currBlock.AddInstructions(jmp)
	errSupressBlock.AddPredecessor(cb.currBlock)
	cb.currBlock = errSupressBlock

	// parse expression
	result := cb.parseExprNode(expr.Expr)
	// create new block as end block
	endBlock := NewBlock(cb.GetBlockId())
	jmp = NewOpStmtJump(endBlock, expr.Position)
	cb.currBlock.AddInstructions(jmp)
	endBlock.AddPredecessor(cb.currBlock)
	cb.currBlock = endBlock

	return result
}

func (cb *CfgBuilder) parseExprExit(expr *ast.ExprExit) Operand {
	var e Operand = nil
	if expr.Expr != nil {
		e = cb.readVariable(cb.parseExprNode(expr.Expr))
	}

	// create exit op
	exitOp := NewOpExit(e, expr.Position)
	cb.currBlock.AddInstructions(exitOp)
	// ignore all code after exit
	cb.currBlock = NewBlock(cb.GetBlockId())
	cb.currBlock.Dead = true

	// TODO: Check again
	return NewOperNumber(1)
}

func (cb *CfgBuilder) parseExprFuncCall(expr *ast.ExprFunctionCall) Operand {
	args := cb.parseExprList(expr.Args, MODE_READ)
	name := cb.readVariable(cb.parseExprNode(expr.Function))
	// TODO: check if need NsFuncCall
	opFuncCall := NewOpExprFunctionCall(name, args, expr.Position)

	if nameStr, ok := name.(*OperString); ok {
		if tp, ok := GetTypeAssertFunc(nameStr.Val); ok {
			assert := NewTypeAssertion(NewOperString(tp), false)
			opFuncCall.Result.AddAssertion(args[0], assert, ASSERT_MODE_INTERSECT)
		}
	}

	cb.currBlock.AddInstructions(opFuncCall)

	return opFuncCall.Result
}

func (cb *CfgBuilder) parseExprNew(expr *ast.ExprNew) Operand {
	var className Operand
	switch ec := expr.Class.(type) {
	case *ast.StmtClass:
		// anonymous class
		className = cb.parseExprNode(ec.Name)
	default:
		className = cb.parseExprNode(ec)
	}

	args := cb.parseExprList(expr.Args, MODE_READ)
	opNew := NewOpExprNew(className, args, expr.Position)
	cb.currBlock.AddInstructions(opNew)

	// set result type to object operand
	opNew.Result = NewOperObject(className.(*OperString).Val)

	return opNew.Result
}

func (cb *CfgBuilder) parseExprTernary(expr *ast.ExprTernary) Operand {
	cond := cb.readVariable(cb.parseExprNode(expr.Cond))
	ifBlock := NewBlock(cb.GetBlockId())
	elseBlock := NewBlock(cb.GetBlockId())
	endBlock := NewBlock(cb.GetBlockId())

	jmpIf := NewOpStmtJumpIf(cond, ifBlock, elseBlock, expr.Position)
	cb.currBlock.AddInstructions(jmpIf)
	cb.currBlock.IsConditional = true
	cb.processAssertion(cond, ifBlock, elseBlock)
	ifBlock.AddPredecessor(cb.currBlock)
	elseBlock.AddPredecessor(cb.currBlock)

	// build ifTrue block
	cb.currBlock = ifBlock
	ifVar := NewOperTemporary(nil)
	var ifAssignOp *OpExprAssign
	// if there is ifTrue value, assign ifVar with it
	// else, assign with 1
	if expr.IfTrue != nil {
		ifVal := cb.readVariable(cb.parseExprNode(expr.IfTrue))
		ifAssignOp = NewOpExprAssign(ifVar, ifVal, expr.Position)
	} else {
		ifAssignOp = NewOpExprAssign(ifVar, NewOperNumber(1), expr.Position)
	}
	cb.currBlock.AddInstructions(ifAssignOp)
	// add jump op to end block
	jmp := NewOpStmtJump(endBlock, expr.Position)
	cb.currBlock.AddInstructions(jmp)

	// build ifFalse block
	cb.currBlock = elseBlock
	elseVar := NewOperTemporary(nil)
	elseVal := cb.readVariable(cb.parseExprNode(expr.IfFalse))
	elseAssignOp := NewOpExprAssign(elseVar, elseVal, expr.Position)
	cb.currBlock.AddInstructions(elseAssignOp)
	// add jump to end block
	jmp = NewOpStmtJump(endBlock, expr.Position)
	cb.currBlock.AddInstructions(jmp)

	// build end block
	cb.currBlock = endBlock
	result := NewOperTemporary(nil)
	phi := NewOpPhi(result, cb.currBlock, nil)
	phi.AddOperand(ifVar)
	phi.AddOperand(elseVar)
	cb.currBlock.AddPhi(phi)

	// return phi
	return result
}

func (cb *CfgBuilder) parseExprYield(expr *ast.ExprYield) Operand {
	var key Operand
	var val Operand

	// TODO: handle index key (0, 1, 2, etc)
	if expr.Key != nil {
		key = cb.readVariable(cb.parseExprNode(expr.Key))
	}
	if expr.Val != nil {
		val = cb.readVariable(cb.parseExprNode(expr.Val))
	}

	yieldOp := NewOpExprYield(val, key, expr.Position)
	cb.currBlock.AddInstructions(yieldOp)

	return yieldOp.Result
}

// function to parse list of expressions node
func (cb *CfgBuilder) parseExprList(exprs []ast.Vertex, mode VAR_MODE) []Operand {
	vars := make([]Operand, 0)
	switch mode {
	case MODE_READ:
		for _, expr := range exprs {
			vars = append(vars, cb.readVariable(cb.parseExprNode(expr)))
		}
	case MODE_WRITE:
		for _, expr := range exprs {
			vars = append(vars, cb.writeVariable(cb.parseExprNode(expr)))
		}
	case MODE_NONE:
		for _, expr := range exprs {
			vars = append(vars, cb.parseExprNode(expr))
		}
	}

	return vars
}

func (cb *CfgBuilder) parseTypeNode(node ast.Vertex) OpType {
	switch n := node.(type) {
	case nil:
		return NewOpTypeMixed(nil)
	case *ast.Name:
		name, _ := astutil.GetNameString(n)
		if IsBuiltInType(name) {
			return NewOpTypeLiteral(name, false, n.Position)
		} else if name == "mixed" {
			return NewOpTypeMixed(n.Position)
		} else if name == "void" {
			return NewOpTypeVoid(n.Position)
		} else {
			declaration := cb.readVariable(cb.parseExprNode(n))
			return NewOpTypeReference(declaration, false, n.Position)
		}
	case *ast.Nullable:
		subType := cb.parseTypeNode(n.Expr)
		switch t := subType.(type) {
		case *OpTypeLiteral:
			t.IsNullable = true
			return t
		case *OpTypeReference:
			t.IsNullable = true
			return t
		default:
			log.Fatal("Error: Invalid nullable type")
		}
	case *ast.Union:
		types := make([]OpType, 0)
		for _, tp := range n.Types {
			types = append(types, cb.parseTypeNode(tp))
		}
		return NewOpTypeUnion(types, n.Position)
	case *ast.Identifier:
		return NewOpTypeLiteral(string(n.Value), false, n.Position)
	default:
		log.Fatal("Error: invalid type node")
	}
	return nil
}

func (cb *CfgBuilder) ParseShortCircuiting(expr ast.Vertex, isOr bool) Operand {
	var left ast.Vertex
	var right ast.Vertex

	switch e := expr.(type) {
	case *ast.ExprBinaryBooleanAnd:
		left = e.Left
		right = e.Right
	case *ast.ExprBinaryLogicalAnd:
		left = e.Left
		right = e.Right
	case *ast.ExprBinaryBooleanOr:
		left = e.Left
		right = e.Right
	case *ast.ExprBinaryLogicalOr:
		left = e.Left
		right = e.Right
	default:
		log.Fatalf("Error invalid expr '%v' in parseShortCircuiting", reflect.TypeOf(e))
		return nil
	}

	// create temporary operand as result
	result := NewOperTemporary(nil)

	// create 2 blocks for if and else condition
	longBlock := NewBlock(cb.GetBlockId())
	endBlock := NewBlock(cb.GetBlockId())
	var ifCond *Block
	var elseCond *Block
	var mode AssertMode
	if isOr {
		ifCond = endBlock
		elseCond = longBlock
		mode = ASSERT_MODE_UNION
	} else {
		ifCond = longBlock
		elseCond = endBlock
		mode = ASSERT_MODE_INTERSECT
	}

	// parse left node first
	leftVal := cb.readVariable(cb.parseExprNode(left))

	// create jumpIf op and adding it as currBlock next instruction
	jmpIf := NewOpStmtJumpIf(leftVal, ifCond, elseCond, expr.GetPosition())
	cb.currBlock.AddInstructions(jmpIf)
	cb.currBlock.IsConditional = true
	longBlock.AddPredecessor(cb.currBlock)
	endBlock.AddPredecessor(cb.currBlock)

	// build long block
	cb.currBlock = longBlock
	rightVal := cb.readVariable(cb.parseExprNode(right))
	rightBool := NewOpExprCastBool(rightVal, nil)
	cb.currBlock.AddInstructions(rightBool)
	cb.currBlock.AddInstructions(NewOpStmtJump(endBlock, expr.GetPosition()))
	endBlock.AddPredecessor(cb.currBlock)

	// build end block
	cb.currBlock = endBlock
	phi := NewOpPhi(result, cb.currBlock, nil)
	phi.AddOperand(NewOperBool(isOr))
	phi.AddOperand(rightBool.Result)
	cb.currBlock.AddPhi(phi)

	// create an assertion
	for _, assert := range leftVal.GetAssertions() {
		result.AddAssertion(assert.Var, assert.Assert, mode)
	}
	for _, assert := range rightVal.GetAssertions() {
		result.AddAssertion(assert.Var, assert.Assert, mode)
	}

	return result
}

func (cb *CfgBuilder) processAssertion(oper Operand, ifBlock *Block, elseBlock *Block) {
	if ifBlock == nil {
		log.Fatalf("Error in processAssertion: ifBlock cannot be nil")
	} else if elseBlock == nil {
		log.Fatalf("Error in processAssertion: elseBlock cannot be nil")
	}
	// TODO: Check if we need push assertion in reverse order
	block := cb.currBlock
	for _, assert := range oper.GetAssertions() {
		// add assertion into if block
		cb.currBlock = ifBlock
		read := cb.readVariable(assert.Var)
		write := cb.writeVariable(assert.Var)
		a := cb.readAssertion(assert.Assert)
		opAssert := NewOpExprAssertion(read, write, a, nil)
		cb.currBlock.AddInstructions(opAssert)

		// add negation of the assertion into else block
		cb.currBlock = elseBlock
		read = cb.readVariable(assert.Var)
		write = cb.writeVariable(assert.Var)
		a = cb.readAssertion(assert.Assert).GetNegation()
		opAssert = NewOpExprAssertion(read, write, a, nil)
		cb.currBlock.AddInstructions(opAssert)
	}
	cb.currBlock = block
}

func (cb *CfgBuilder) readAssertion(assert Assertion) Assertion {
	switch a := assert.(type) {
	case *TypeAssertion:
		vr := cb.readVariable(a.Val)
		return NewTypeAssertion(vr, a.IsNegated)
	case *CompositeAssertion:
		vrs := make([]Assertion, 0)
		for _, assertChild := range a.Val {
			vrs = append(vrs, cb.readAssertion(assertChild))
		}
		return NewCompositeAssertion(vrs, a.Mode, a.IsNegated)
	}
	log.Fatal("Error: Wrong assertion type")
	return nil
}

// add a new variable definition
func (cb *CfgBuilder) writeVariable(vr Operand) Operand {
	// get the original variable
	for vrTemp, ok := vr.(*OperTemporary); ok; {
		vr = vrTemp.Original
		vrTemp, ok = vr.(*OperTemporary)
	}

	// write variable by name
	if vrVar, ok := vr.(*OperVariable); ok {
		switch name := vrVar.Name.(type) {
		case *OperString:
			nameString := name.Val
			vr = NewOperTemporary(vr)
			cb.writeVariableName(nameString, vr, cb.currBlock)
		case *OperVariable:
			// variable variables, just register read
			cb.readVariable(name)
		default:
			log.Fatal("Error: Invalid operand type for a variable name")
		}
	}

	return vr
}

func (cb *CfgBuilder) writeVariableName(name string, val Operand, block *Block) {
	cb.Ctx.SetValueInScope(block, name, val)
}

// TODO: Check name type
// read defined variable
func (cb *CfgBuilder) readVariable(vr Operand) Operand {
	if vr == nil {
		log.Fatal("Error: read nil operand")
	}
	// TODO: Code preprocess
	switch v := vr.(type) {
	case *OperBoundVar:
		return v
	case *OperVariable:
		// if variable name is string, read variable name
		// else if a variable, it's a variable variables
		switch varName := v.Name.(type) {
		case *OperString:
			return cb.readVariableName(varName.Val, cb.currBlock)
		case *OperVariable:
			cb.readVariable(varName)
			return vr
		default:
			log.Fatal("Error variable name")
		}
	case *OperTemporary:
		if v.Original != nil {
			return cb.readVariable(v.Original)
		}
	}

	return vr
}

// TODO: change name type
func (cb *CfgBuilder) readVariableName(name string, block *Block) Operand {
	val, ok := cb.Ctx.getLocalVar(block, name)
	if ok {
		return val
	}

	// symbolic interpreter
	switch name {
	case "$_GET":
		cb.currFunc.ContaintTainted = true
		cb.currBlock.ContaintTainted = true
		return NewOperSymbolic("getsymbolic", true)
	case "$_POST":
		cb.currFunc.ContaintTainted = true
		cb.currBlock.ContaintTainted = true
		return NewOperSymbolic("postsymbolic", true)
	case "$_REQUEST":
		cb.currFunc.ContaintTainted = true
		cb.currBlock.ContaintTainted = true
		return NewOperSymbolic("requestsymbolic", true)
	case "$_FILES":
		cb.currFunc.ContaintTainted = true
		cb.currBlock.ContaintTainted = true
		return NewOperSymbolic("filessymbolic", true)
	case "$_COOKIE":
		cb.currFunc.ContaintTainted = true
		cb.currBlock.ContaintTainted = true
		return NewOperSymbolic("cookiesymbolic", true)
	case "$_SERVERS":
		cb.currFunc.ContaintTainted = true
		cb.currBlock.ContaintTainted = true
		return NewOperSymbolic("serverssymbolic", true)
	}

	return cb.readVariableRecursive(name, block)
}

// TODO: Check name type
func (cb *CfgBuilder) readVariableRecursive(name string, block *Block) Operand {
	vr := Operand(nil)
	if !cb.Ctx.Complete {
		// Incomplete CFG, create an incomplete phi
		vr = NewOperTemporary(NewOperVar(NewOperString(name), nil))
		phi := NewOpPhi(vr, block, nil)
		cb.Ctx.addIncompletePhis(block, name, phi)
		cb.writeVariableName(name, vr, block)
	} else if len(block.Preds) == 1 && !block.Preds[0].Dead {
		// 1 Predecessors, read from predecessor
		vr = cb.readVariableName(name, block.Preds[0])
		cb.writeVariableName(name, vr, block)
	} else {
		// break potential cycles with operandless phi
		vr = NewOperTemporary(NewOperVar(NewOperString(name), nil))
		phi := NewOpPhi(vr, block, nil)
		block.AddPhi(phi)
		cb.writeVariableName(name, vr, block)

		// get phi operand from its predecessors
		for _, pred := range block.Preds {
			if !pred.Dead {
				oper := cb.readVariableName(name, pred)
				phi.AddOperand(oper)
			}
		}
	}

	return vr
}

func (cb *CfgBuilder) getAnonId() int {
	id := cb.AnonId
	cb.AnonId += 1
	return id
}

func (cb *CfgBuilder) GetBlockId() BlockId {
	id := cb.BlockCnt
	cb.BlockCnt += 1
	return id
}

func IsBuiltInType(name string) bool {
	switch name {
	case "self":
		fallthrough
	case "parent":
		fallthrough
	case "static":
		fallthrough
	case "int":
		fallthrough
	case "integer":
		fallthrough
	case "long":
		fallthrough
	case "float":
		fallthrough
	case "double":
		fallthrough
	case "real":
		fallthrough
	case "array":
		fallthrough
	case "object":
		fallthrough
	case "bool":
		fallthrough
	case "boolean":
		fallthrough
	case "null":
		fallthrough
	case "void":
		fallthrough
	case "false":
		fallthrough
	case "true":
		fallthrough
	case "string":
		fallthrough
	case "mixed":
		fallthrough
	case "resource":
		fallthrough
	case "callable":
		return true
	}
	return false
}

func GetTypeAssertFunc(funcName string) (string, bool) {
	funcName = strings.ToLower(funcName)
	switch funcName {
	case "is_array":
		return "array", true
	case "is_bool":
		return "bool", true
	case "is_callable":
		return "callable", true
	case "is_double":
		return "float", true
	case "is_float":
		return "float", true
	case "is_int":
		return "int", true
	case "is_integer":
		return "int", true
	case "is_long":
		return "int", true
	case "is_null":
		return "null", true
	case "is_numeric":
		return "numeric", true
	case "is_object":
		return "object", true
	case "is_real":
		return "float", true
	case "is_string":
		return "string", true
	case "is_resource":
		return "resource", true
	}
	return "", false
}
