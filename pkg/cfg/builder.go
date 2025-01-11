package cfg

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/asttraverser"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/asttraverser/loopresolver"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/asttraverser/mcresolver"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/asttraverser/nsresolver"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/astutil"
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
	FilePath      string
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
	CurrNamespace string
	Script        *Script
	AnonId        int     // id for naming anonymous thing like closure
	BlockCnt      BlockId // counter to generate block id
	Consts        map[string]Operand
	VarNames      map[string]struct{}

	DefinedArr map[string]map[string]struct{}
	currBlock  *Block
	currFunc   *OpFunc
}

func BuildCFG(src []byte, filePath string, autoloadConfig map[string]string) *Script {
	cb := &CfgBuilder{
		AutloadConfig: autoloadConfig,
		Consts:        make(map[string]Operand),
		AnonId:        0,
		VarNames:      make(map[string]struct{}),
		DefinedArr:    make(map[string]map[string]struct{}),
	}

	fileName := filepath.Base(filePath)
	root := cb.parseAst(src, fileName)

	entryBlock := NewBlock(cb.GetBlockId())
	mainFunc, err := NewFunc("{main}", FUNC_MODIF_PUBLIC, NewOpTypeVoid(nil), entryBlock, nil)
	if err != nil {
		log.Fatalf("Error in parseRoot: %v", err)
	}
	cb.Script = &Script{
		Funcs:    make(map[string]*OpFunc),
		Main:     mainFunc,
		FilePath: filePath,
	}

	cb.parseOpFunc(mainFunc, nil, root.Stmts)

	return cb.Script
}

func (cb *CfgBuilder) parseAst(src []byte, fileName string) *ast.Root {
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
	mcRes := mcresolver.NewMagicConstResolver(fileName)
	travs := asttraverser.NewTraverser()
	travs.AddNodeTraverser(nsRes, lRes, mcRes)
	travs.Traverse(root)

	return root
}

// func (cb *CfgBuilder) parseRoot(n *ast.Root) {
// 	// Create script instance
// 	entryBlock := NewBlock(cb.GetBlockId())
// 	mainFunc, err := NewFunc("{main}", FUNC_MODIF_PUBLIC, NewOpTypeVoid(nil), entryBlock, nil)
// 	if err != nil {
// 		log.Fatalf("Error in parseRoot: %v", err)
// 	}
// 	cb.Script = &Script{
// 		Funcs: make(map[string]*OpFunc),
// 		Main:  mainFunc,
// 	}

// 	cb.parseOpFunc(mainFunc, nil, n.Stmts)
// }

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
		fmt.Println("Error: there are still unresolved gotos")
		// log.Fatal("Error: there are still unresolved gotos")
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
		log.Fatalf("Error: Invalid statement node type '%v'", reflect.TypeOf(n))
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

	name, err := cb.readVariable(cb.parseExprNode(stmt.Name))
	if err != nil {
		log.Fatalf("Error in parseStmtConst: %v", err)
	}
	opConst := NewOpConst(name, val, valBlock, stmt.Position)
	cb.currBlock.AddInstructions(opConst)

	// define the constant in this block
	nameStr, err := GetOperName(name)
	if err != nil {
		log.Fatalf("Error in parseStmtConst: %v", err)
	}
	if cb.currFunc == cb.Script.Main {
		cb.Consts[nameStr] = val
	}
}

func (cb *CfgBuilder) parseStmtDeclare(stmt *ast.StmtDeclare) {
	// TODO: right now, it isn't important
}

func (cb *CfgBuilder) parseStmtEcho(stmt *ast.StmtEcho) {
	for _, expr := range stmt.Exprs {
		exprOper, err := cb.readVariable(cb.parseExprNode(expr))
		if err != nil {
			log.Fatalf("Error in parseStmtEcho: %v", err)
		}
		echoOp := NewOpEcho(exprOper, expr.GetPosition())
		cb.currBlock.AddInstructions(echoOp)
	}
}

func (cb *CfgBuilder) parseStmtReturn(stmt *ast.StmtReturn) {
	expr := Operand(nil)
	if stmt.Expr != nil {
		var err error
		expr, err = cb.readVariable(cb.parseExprNode(stmt.Expr))
		if err != nil {
			log.Fatalf("Error in parseStmtReturn: %v", err)
		}
	}

	returnOp := NewOpReturn(expr, stmt.Position)
	cb.currBlock.AddInstructions(returnOp)

	// script after return will be a dead code
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
	exprs, exprsPos := cb.parseExprList(stmt.Vars, MODE_WRITE)
	op := NewOpUnset(exprs, exprsPos, stmt.Position)
	cb.currBlock.AddInstructions(op)
}

func (cb *CfgBuilder) parseStmtThrow(stmt *ast.StmtThrow) {
	expr, err := cb.readVariable(cb.parseExprNode(stmt.Expr))
	if err != nil {
		log.Fatalf("Error in parseStmtThrow: %v", err)
	}
	op := NewOpThrow(expr, stmt.Position)
	cb.currBlock.AddInstructions(op)
	// script after throw will be a dead code
	cb.currBlock = NewBlock(cb.GetBlockId())
	cb.currBlock.Dead = true
}

func (cb *CfgBuilder) parseStmtTry(stmt *ast.StmtTry) {
	// parse statements inside try block
	cb.parseStmtNodes(stmt.Stmts, cb.currBlock)
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
		cond, err := cb.readVariable(cb.parseExprNode(stmt.Cond))
		if err != nil {
			log.Fatalf("Error in parseStmtSwitch: %v", err)
		}
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
				caseValue := cb.parseExprNode(cn.Cond)
				caseCond := NewOpExprBinaryEqual(cond, caseValue, cn.Position).Result
				// add condition to case block
				cb.Ctx.PushCond(caseCond)
				caseBlock.SetCondition(cb.Ctx.CurrConds)
				targets = append(targets, caseBlock)
				cases = append(cases, caseValue)
				prevBlock, err = cb.parseStmtNodes(cn.Stmts, caseBlock)
				// return the condition
				cb.Ctx.PopCond()
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
				left, err := cb.readVariable(cond)
				if err != nil {
					log.Fatalf("Error in StmtCase: %v", err)
				}
				right, err := cb.readVariable(caseExpr)
				if err != nil {
					log.Fatalf("Error in StmtCase: %v", err)
				}
				opEqual := NewOpExprBinaryEqual(left, right, cn.Position)
				cb.currBlock.AddInstructions(opEqual)

				elseBlock := NewBlock(cb.GetBlockId())
				opJmpIf := NewOpStmtJumpIf(opEqual.Result, ifBlock, elseBlock, cn.Position)
				cb.currBlock.AddInstructions(opJmpIf)
				cb.currBlock.IsConditional = true
				ifBlock.AddPredecessor(cb.currBlock)
				elseBlock.AddPredecessor(cb.currBlock)
				cb.currBlock = elseBlock
				// add condition to if Block
				cb.Ctx.PushCond(opEqual.Result)
				ifBlock.SetCondition(cb.Ctx.CurrConds)
				prevBlock, err = cb.parseStmtNodes(cn.Stmts, ifBlock)
				// return condition
				cb.Ctx.PopCond()
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
		switch csT := cs.(type) {
		case *ast.StmtCase:
			if !astutil.IsScalarNode(csT.Cond) {
				return false
			}
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
	var err error
	switch n := stmt.(type) {
	case *ast.StmtIf:
		condPosition = n.Cond.GetPosition()
		cond, err = cb.readVariable(cb.parseExprNode(n.Cond))
		if err != nil {
			log.Fatalf("Error in parseIf : %v", err)
		}
		switch stmtT := n.Stmt.(type) {
		case *ast.StmtStmtList:
			stmts = stmtT.Stmts
		case *ast.StmtExpression:
			stmtExpr := &ast.StmtExpression{
				Position: stmtT.Expr.GetPosition(),
				Expr:     stmtT.Expr,
			}
			stmts = []ast.Vertex{stmtExpr}
		}
	case *ast.StmtElseIf:
		condPosition = n.Cond.GetPosition()
		cond, err = cb.readVariable(cb.parseExprNode(n.Cond))
		if err != nil {
			log.Fatalf("Error in parseIf: %v", err)
		}
		switch stmtT := n.Stmt.(type) {
		case *ast.StmtStmtList:
			stmts = stmtT.Stmts
		case *ast.StmtExpression:
			stmtExpr := &ast.StmtExpression{
				Position: stmtT.Expr.GetPosition(),
				Expr:     stmtT.Expr,
			}
			stmts = []ast.Vertex{stmtExpr}
		}
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

	// add condition to the block
	cb.Ctx.PushCond(cond)
	ifBlock.SetCondition(cb.Ctx.CurrConds)
	cb.currBlock, err = cb.parseStmtNodes(stmts, ifBlock)
	cb.Ctx.PopCond()
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
			// else if
			if elseIfNode, ok := ifNode.Else.(*ast.StmtElse).Stmt.(*ast.StmtIf); ok {
				cb.parseIf(elseIfNode, endBlock)
				return
			}

			stmts, err := astutil.GetStmtList(ifNode.Else.(*ast.StmtElse).Stmt)
			if err != nil {
				log.Fatalf("Error in parseIf: %v", err)
			}

			// add condition
			negatedCond := NewOpExprBooleanNot(cond, condPosition).Result
			cb.Ctx.PushCond(negatedCond)
			elseBlock.SetCondition(cb.Ctx.CurrConds)

			cb.currBlock, err = cb.parseStmtNodes(stmts, cb.currBlock)
			if err != nil {
				log.Fatalf("Error in parseIf: %v", err)
			}

			cb.Ctx.PopCond()
		}
		jmp := NewOpStmtJump(endBlock, ifNode.Position)
		cb.currBlock.AddInstructions(jmp)
		endBlock.AddPredecessor(cb.currBlock)
	}
}

func (cb *CfgBuilder) parseStmtGoto(stmt *ast.StmtGoto) {
	labelName := ""
	var err error
	switch stmt.Label.(type) {
	case *ast.StmtLabel:
		labelName, err = astutil.GetNameString(stmt.Label.(*ast.StmtLabel).Name)
		if err != nil {
			// TODO
			fmt.Printf("Error in StmtGoto: %v\n", err)
			return
		}
	case *ast.Identifier:
		labelName, err = astutil.GetNameString(stmt.Label)
		if err != nil {
			// TODO
			fmt.Printf("Error in StmtGoto: %v\n", err)
			return
		}
	}

	if labelBlock, ok := cb.Ctx.getLabel(labelName); ok {
		cb.currBlock.AddInstructions(NewOpStmtJump(labelBlock, stmt.Position))
		labelBlock.AddPredecessor(cb.currBlock)
	} else {
		cb.Ctx.addUnresolvedGoto(labelName, cb.currBlock)
	}

	// script after return will be a dead code
	cb.currBlock = NewBlock(cb.GetBlockId())
	cb.currBlock.Dead = true
}

func (cb *CfgBuilder) parseStmtLabel(stmt *ast.StmtLabel) {
	labelName, err := astutil.GetNameString(stmt.Name)
	if err != nil {
		log.Fatal("Error label name in StmtLabel")
	}
	if _, ok := cb.Ctx.getLabel(labelName); ok {
		fmt.Println("Error: label '", labelName, "' have been defined")
		return
	}

	labelBlock := NewBlock(cb.GetBlockId())
	jmp := NewOpStmtJump(labelBlock, stmt.Position)
	cb.currBlock.AddInstructions(jmp)
	labelBlock.AddPredecessor(cb.currBlock)

	// add condition to label block
	labelBlock.SetCondition(cb.Ctx.CurrConds)

	// add jump to label block for every unresolved goto
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
	// no need to add condition cause do block will always be executed
	cb.currBlock = bodyBlock
	cb.currBlock, err = cb.parseStmtNodes(stmt.Stmt.(*ast.StmtStmtList).Stmts, bodyBlock)
	if err != nil {
		log.Fatalf("Error in parseStmtNodes: %v", err)
	}
	cond, err := cb.readVariable(cb.parseExprNode(stmt.Cond))
	if err != nil {
		log.Fatalf("Error in parseStmtDo: %v", err)
	}
	cb.currBlock.AddInstructions(NewOpStmtJumpIf(cond, bodyBlock, endBlock, stmt.Cond.GetPosition()))
	cb.currBlock.IsConditional = true
	cb.processAssertion(cond, bodyBlock, endBlock)
	bodyBlock.AddPredecessor(cb.currBlock)
	endBlock.AddPredecessor(cb.currBlock)

	// add condition to end block
	negatedCond := NewOpExprBooleanNot(cond, nil).Result
	cb.Ctx.PushCond(negatedCond)
	endBlock.SetCondition(cb.Ctx.CurrConds)
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
		vr, _ := cb.parseExprList(stmt.Cond, MODE_NONE)
		cond, err = cb.readVariable(vr[len(vr)-1])
		if err != nil {
			log.Fatalf("Error in parseStmtFor: %v", err)
		}
	} else {
		cond = NewOperBool(true)
	}
	cb.currBlock.AddInstructions(NewOpStmtJumpIf(cond, bodyBlock, endBlock, stmt.Position))
	cb.currBlock.IsConditional = true
	cb.processAssertion(cond, bodyBlock, endBlock)
	bodyBlock.AddPredecessor(cb.currBlock)
	endBlock.AddPredecessor(cb.currBlock)

	// add condition to block
	cb.Ctx.PushCond(cond)
	bodyBlock.SetCondition(cb.Ctx.CurrConds)
	// parse statements inside loop body
	// will create new block cause of label
	stmts, err := astutil.GetStmtList(stmt.Stmt)
	if err != nil {
		log.Fatalf("Error in parseStmtFor: %v", err)
	}
	cb.currBlock, err = cb.parseStmtNodes(stmts, bodyBlock)
	cb.Ctx.PopCond()
	if err != nil {
		log.Fatalf("Error in parseStmtFor: %v", err)
	}
	cb.parseExprList(stmt.Loop, MODE_READ)
	// go back to init block
	cb.currBlock.AddInstructions(NewOpStmtJump(initBlock, stmt.Position))
	initBlock.AddPredecessor(cb.currBlock)
	// add condition to endblock
	negatedCond := NewOpExprBooleanNot(cond, nil).Result
	cb.Ctx.PushCond(negatedCond)
	endBlock.SetCondition(cb.Ctx.CurrConds)
	cb.currBlock = endBlock
}

func (cb *CfgBuilder) parseStmtForeach(stmt *ast.StmtForeach) {
	var err error
	iterable, err := cb.readVariable(cb.parseExprNode(stmt.Expr))
	if err != nil {
		log.Fatalf("Error in parseStmtForEach: %v", err)
	}
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
	initBlock.AddInstructions(NewOpStmtJumpIf(validOp.Result, bodyBlock, endBlock, stmt.Position))
	initBlock.IsConditional = true
	cb.processAssertion(validOp.Result, bodyBlock, endBlock)
	bodyBlock.AddPredecessor(cb.currBlock)
	endBlock.AddPredecessor(cb.currBlock)

	// parse body
	cb.currBlock = bodyBlock
	if stmt.Key != nil {
		keyOp := NewOpExprKey(iterable, stmt.Key.GetPosition())
		keyVar, err := cb.readVariable(cb.parseExprNode(stmt.Key))
		if err != nil {
			log.Fatalf("Error in parseStmtForEach (key): %v", err)
		}
		cb.currBlock.AddInstructions(keyOp)
		assignOp := NewOpExprAssign(keyVar, keyOp.Result, stmt.Key.GetPosition(), stmt.Key.GetPosition(), stmt.Key.GetPosition())
		cb.currBlock.AddInstructions(assignOp)
	}
	isRef := stmt.AmpersandTkn != nil
	valueOp := NewOpExprValue(iterable, isRef, stmt.Var.GetPosition())

	// assign each item to variable
	switch v := stmt.Var.(type) {
	case *ast.ExprList:
		cb.parseAssignList(v.Items, valueOp.Result, nil)
	case *ast.ExprArray:
		cb.parseAssignList(v.Items, valueOp.Result, nil)
	default:
		vr, err := cb.readVariable(cb.parseExprNode(stmt.Var))
		if err != nil {
			log.Fatalf("Error in parseStmtForEach (default): %v", err)
		}
		if isRef {
			cb.currBlock.AddInstructions(NewOpExprAssignRef(vr, valueOp.Result, stmt.Var.GetPosition()))
		} else {
			cb.currBlock.AddInstructions(NewOpExprAssign(vr, valueOp.Result, stmt.Var.GetPosition(), stmt.Var.GetPosition(), stmt.Var.GetPosition()))
		}
	}

	// parse statements inside loop body
	stmts, err := astutil.GetStmtList(stmt.Stmt)
	if err != nil {
		log.Fatalf("Error in parseStmtForEach: %v", err)
	}
	cb.currBlock, err = cb.parseStmtNodes(stmts, cb.currBlock)
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
	cond, err := cb.readVariable(cb.parseExprNode(stmt.Cond))
	if err != nil {
		log.Fatalf("Error in parseStmtWhile: %v", err)
	}
	cb.currBlock.AddInstructions(NewOpStmtJumpIf(cond, bodyBlock, endBlock, stmt.Cond.GetPosition()))
	cb.currBlock.IsConditional = true
	bodyBlock.AddPredecessor(cb.currBlock)
	endBlock.AddPredecessor(cb.currBlock)

	// parse statements inside body loop
	// add condition to body block
	cb.Ctx.PushCond(cond)
	bodyBlock.SetCondition(cb.Ctx.CurrConds)
	stmts, err := astutil.GetStmtList(stmt.Stmt)
	if err != nil {
		log.Fatalf("Error in parseStmtWhile: %v", err)
	}
	cb.currBlock, err = cb.parseStmtNodes(stmts, bodyBlock)
	// return condition
	cb.Ctx.PopCond()
	if err != nil {
		log.Fatalf("Error in parseStmtWhile: %v", err)
	}

	// go back to init block
	cb.currBlock.AddInstructions(NewOpStmtJump(initBlock, stmt.Position))
	initBlock.AddPredecessor(cb.currBlock)

	// add condition to end block
	negatedCond := NewOpExprBooleanNot(cond, nil).Result
	cb.Ctx.PushCond(negatedCond)
	endBlock.SetCondition(cb.Ctx.CurrConds)
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
	implements, _ := cb.parseExprList(stmt.Implements, MODE_NONE)

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
	stmts, err := astutil.GetStmtList(stmt.Stmt)
	if err != nil {
		log.Fatalf("Error in parseStmtClassMethod: %v", err)
	}
	cb.parseOpFunc(fn, stmt.Params, stmts)

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
	name, err := cb.readVariable(cb.parseExprNode(stmt.Name))
	if err != nil {
		log.Fatalf("Error in parseStmtInterface: %v", err)
	}
	tmpClass := cb.CurrClass
	cb.CurrClass = name.(*OperString)

	extends, _ := cb.parseExprList(stmt.Extends, MODE_NONE)
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
	if stmt.Name == nil {
		return
	}
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
			arg, err := cb.readVariable(cb.parseExprNode(argNode.(*ast.Argument).Expr))
			if err != nil {
				log.Fatalf("Error in parseAttrGroup: %v", err)
			}
			args = append(args, arg)
		}
		attrName, err := cb.readVariable(cb.parseExprNode(attrNode.(*ast.Attribute).Name))
		if err != nil {
			log.Fatalf("Error in parseAttrGroup: %v", err)
		}
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
		num := 0.0
		var err error
		if string(exprT.Value[:2]) == "0x" {
			// hex number
			numInt, err := strconv.ParseInt(string(exprT.Value), 0, 64)
			if err != nil {
				log.Fatalf("Error in scalarDNumber: %v", err)
			}
			num = float64(numInt)
		} else {
			num, err = strconv.ParseFloat(string(exprT.Value), 64)
			if err != nil {
				log.Fatal(err)
			}
		}
		return NewOperNumber(num)
	case *ast.ScalarLnumber:
		num := 0.0
		var err error
		if string(exprT.Value[:2]) == "0x" {
			// hex number
			numInt, err := strconv.ParseInt(string(exprT.Value), 0, 64)
			if err != nil {
				log.Fatalf("Error in scalarLNumber: %v", err)
			}
			num = float64(numInt)
		} else {
			num, err = strconv.ParseFloat(string(exprT.Value), 64)
			if err != nil {
				log.Fatal(err)
			}
		}
		return NewOperNumber(num)
	case *ast.ScalarString:
		str := string(exprT.Value)
		if str[0] == '"' {
			str = str[1:]
		}
		if str[len(str)-1] == '"' {
			str = str[:len(str)-1]
		}
		return NewOperString(str)
	case *ast.ScalarEncapsed:
		parts, partsPos := cb.parseExprList(exprT.Parts, MODE_READ)
		op := NewOpExprConcatList(parts, partsPos, exprT.Position)
		cb.currBlock.Instructions = append(cb.currBlock.Instructions, op)
		return op.Result
	case *ast.ScalarEncapsedStringBrackets:
		return cb.parseExprNode(exprT.Var)
	case *ast.ScalarEncapsedStringVar:
		// TODO
		return NewOperString("")
	case *ast.ScalarEncapsedStringPart:
		str := string(exprT.Value)
		return NewOperString(str)
	case *ast.ScalarHeredoc:
		parts, partsPos := cb.parseExprList(exprT.Parts, MODE_READ)
		op := NewOpExprConcatList(parts, partsPos, exprT.Position)
		cb.currBlock.Instructions = append(cb.currBlock.Instructions, op)
		return op.Result
	case *ast.Argument:
		vr, err := cb.readVariable(cb.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in Argument: %v", err)
		}
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
		e, err := cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprCastArray: %v", err)
		}
		op := NewOpExprCastArray(e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastBool:
		vr := cb.parseExprNode(exprT.Expr)
		e, err := cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprCastBool: %v", err)
		}
		op := NewOpExprCastBool(e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastDouble:
		vr := cb.parseExprNode(exprT.Expr)
		e, err := cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprCastDouble: %v", err)
		}
		op := NewOpExprCastDouble(e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastInt:
		vr := cb.parseExprNode(exprT.Expr)
		e, err := cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprCastInt: %v", err)
		}
		op := NewOpExprCastInt(e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastObject:
		vr := cb.parseExprNode(exprT.Expr)
		e, err := cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprCastObject: %v", err)
		}
		op := NewOpExprCastObject(e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastString:
		vr := cb.parseExprNode(exprT.Expr)
		e, err := cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprCastString: %v", err)
		}
		op := NewOpExprCastString(e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprCastUnset:
		vr := cb.parseExprNode(exprT.Expr)
		e, err := cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprCastUnset: %v", err)
		}
		op := NewOpExprCastUnset(e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprUnaryMinus:
		vr := cb.parseExprNode(exprT.Expr)
		val, err := cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprUnaryMinus: %v", err)
		}
		op := NewOpExprUnaryMinus(val, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprUnaryPlus:
		vr := cb.parseExprNode(exprT.Expr)
		val, err := cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprUnaryPlus: %v", err)
		}
		op := NewOpExprUnaryPlus(val, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprArray:
		return cb.parseExprArray(exprT)
	case *ast.ExprArrayDimFetch:
		return cb.parseExprArrayDimFetch(exprT)
	case *ast.ExprBitwiseNot:
		oper, err := cb.readVariable(cb.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprBitwiseNot: %v", err)
		}
		op := NewOpExprBitwiseNot(oper, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBooleanNot:
		cond, err := cb.readVariable(cb.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprBooleanNot: %v", err)
		}
		op := NewOpExprBooleanNot(cond, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprClosure:
		return cb.parseExprClosure(exprT)
	case *ast.ExprArrowFunction:
		return cb.parseExprArrowFunc(exprT)
	case *ast.ExprClassConstFetch:
		class, err := cb.readVariable(cb.parseExprNode(exprT.Class))
		if err != nil {
			log.Fatalf("Error in ExprClassConstFetch (class): %v", err)
		}
		name, err := cb.readVariable(cb.parseExprNode(exprT.Class))
		if err != nil {
			log.Fatalf("Error in ExprClassConstFetch (name): %v", err)
		}
		op := NewOpExprClassConstFetch(class, name, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprClone:
		clone, err := cb.readVariable(cb.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprClone: %v", err)
		}
		op := NewOpExprClone(clone, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprConstFetch:
		return cb.parseExprConstFetch(exprT)
	case *ast.ExprEmpty:
		empty, err := cb.readVariable(cb.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprEmpty: %v", err)
		}
		op := NewOpExprEmpty(empty, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprErrorSuppress:
		return cb.parseExprErrorSuppress(exprT)
	case *ast.ExprEval:
		eval, err := cb.readVariable(cb.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprEval: %v", err)
		}
		op := NewOpExprEval(eval, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprExit:
		return cb.parseExprExit(exprT)
	case *ast.ExprFunctionCall:
		return cb.parseExprFuncCall(exprT)
	case *ast.ExprInclude:
		include, err := cb.readVariable(cb.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprInclude: %v", err)
		}
		// add to include file
		if includeStr, ok := include.(*OperString); ok {
			cb.Script.IncludedFiles = append(cb.Script.IncludedFiles, includeStr.Val)
		}
		op := NewOpExprInclude(include, TYPE_INCLUDE, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprIncludeOnce:
		include, err := cb.readVariable(cb.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprIncludeOnce: %v", err)
		}
		// add to include file
		if includeStr, ok := include.(*OperString); ok {
			cb.Script.IncludedFiles = append(cb.Script.IncludedFiles, includeStr.Val)
		}
		op := NewOpExprInclude(include, TYPE_INCLUDE_ONCE, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprRequire:
		include, err := cb.readVariable(cb.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprRequire: %v", err)
		}
		// add to include file
		if includeStr, ok := include.(*OperString); ok {
			cb.Script.IncludedFiles = append(cb.Script.IncludedFiles, includeStr.Val)
		}
		op := NewOpExprInclude(include, TYPE_REQUIRE, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprRequireOnce:
		include, err := cb.readVariable(cb.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprRequireOnce: %v", err)
		}
		// add to include file
		if includeStr, ok := include.(*OperString); ok {
			cb.Script.IncludedFiles = append(cb.Script.IncludedFiles, includeStr.Val)
		}
		op := NewOpExprInclude(include, TYPE_REQUIRE_ONCE, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprInstanceOf:
		vr, err := cb.readVariable(cb.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprInstanceOf (var): %v", err)
		}
		class, err := cb.readVariable(cb.parseExprNode(exprT.Class))
		if err != nil {
			log.Fatalf("Error in ExprInstanceOf (class): %v", err)
		}
		op := NewOpExprInstanceOf(vr, class, exprT.Position)
		op.Result.AddAssertion(vr, NewTypeAssertion(class, false), ASSERT_MODE_INTERSECT)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprIsset:
		isset, _ := cb.parseExprList(exprT.Vars, MODE_READ)
		op := NewOpExprIsset(isset, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprMethodCall:
		vr, err := cb.readVariable(cb.parseExprNode(exprT.Var))
		if err != nil {
			log.Fatalf("Error in ExprMethodCall (var): %v", err)
		}
		name, err := cb.readVariable(cb.parseExprNode(exprT.Method))
		if err != nil {
			log.Fatalf("Error in ExprMethodCall (name): %v", err)
		}
		args, argsPos := cb.parseExprList(exprT.Args, MODE_READ)
		op := NewOpExprMethodCall(vr, name, args, exprT.Var.GetPosition(), exprT.Method.GetPosition(), argsPos, exprT.Position)
		cb.currBlock.AddInstructions(op)
		cb.currFunc.Calls = append(cb.currFunc.Calls, op)
		return op.Result
	case *ast.ExprNullsafeMethodCall:
		vr, err := cb.readVariable(cb.parseExprNode(exprT.Var))
		if err != nil {
			log.Fatalf("Error in ExprMethodCall (var): %v", err)
		}
		name, err := cb.readVariable(cb.parseExprNode(exprT.Method))
		if err != nil {
			log.Fatalf("Error in ExprMethodCall (name): %v", err)
		}
		args, argsPos := cb.parseExprList(exprT.Args, MODE_READ)
		op := NewOpExprNullSafeMethodCall(vr, name, args, exprT.Var.GetPosition(), exprT.Method.GetPosition(), argsPos, exprT.Position)
		cb.currBlock.AddInstructions(op)
		cb.currFunc.Calls = append(cb.currFunc.Calls, op)
		return op.Result
	case *ast.ExprNew:
		return cb.parseExprNew(exprT)
	case *ast.ExprPostDec:
		vr := cb.parseExprNode(exprT.Var)
		read, err := cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprPostDec: %v", err)
		}
		write := cb.writeVariable(vr)
		opMinus := NewOpExprBinaryMinus(read, NewOperNumber(1), exprT.Position)
		opAssign := NewOpExprAssign(write, opMinus.Result, exprT.Var.GetPosition(), opMinus.Position, exprT.Position)
		cb.currBlock.AddInstructions(opMinus)
		cb.currBlock.AddInstructions(opAssign)
		return read
	case *ast.ExprPostInc:
		vr := cb.parseExprNode(exprT.Var)
		read, err := cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprPostInc: %v", err)
		}
		write := cb.writeVariable(vr)
		opPlus := NewOpExprBinaryPlus(read, NewOperNumber(1), exprT.Position)
		opAssign := NewOpExprAssign(write, opPlus.Result, exprT.Var.GetPosition(), opPlus.Position, exprT.Position)
		cb.currBlock.AddInstructions(opPlus)
		cb.currBlock.AddInstructions(opAssign)
		return read
	case *ast.ExprPreDec:
		vr := cb.parseExprNode(exprT.Var)
		read, err := cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprPreDec: %v", err)
		}
		write := cb.writeVariable(vr)
		opMinus := NewOpExprBinaryMinus(read, NewOperNumber(1), exprT.Position)
		opAssign := NewOpExprAssign(write, opMinus.Result, exprT.Var.GetPosition(), opMinus.Position, exprT.Position)
		cb.currBlock.AddInstructions(opMinus)
		cb.currBlock.AddInstructions(opAssign)
		return opMinus.Result
	case *ast.ExprPreInc:
		vr := cb.parseExprNode(exprT.Var)
		read, err := cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprPreInc: %v", err)
		}
		write := cb.writeVariable(vr)
		opPlus := NewOpExprBinaryPlus(read, NewOperNumber(1), exprT.Position)
		opAssign := NewOpExprAssign(write, opPlus.Result, exprT.Var.GetPosition(), opPlus.Position, exprT.Position)
		cb.currBlock.AddInstructions(opPlus)
		cb.currBlock.AddInstructions(opAssign)
		return opPlus.Result
	case *ast.ExprPrint:
		print, err := cb.readVariable(cb.parseExprNode(exprT.Expr))
		if err != nil {
			log.Fatalf("Error in ExprPrint: %v", err)
		}
		op := NewOpExprPrint(print, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprPropertyFetch:
		vr, err := cb.readVariable(cb.parseExprNode(exprT.Var))
		if err != nil {
			log.Fatalf("Error in ExprPropertyFetch (var): %v", err)
		}
		prop, err := cb.readVariable(cb.parseExprNode(exprT.Prop))
		if err != nil {
			log.Fatalf("Error in ExprPropertyFetch (name): %v", err)
		}
		op := NewOpExprPropertyFetch(vr, prop, exprT.Position)

		varName, _ := GetOperName(vr)
		propStr, ok := GetOperVal(prop).(*OperString)
		if varName != "" && ok {
			propFetchName := "<propfetch>" + varName[1:] + "->" + propStr.Val
			op.Result = NewOperVar(NewOperString(propFetchName), nil)
		}

		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprNullsafePropertyFetch:
		vr, err := cb.readVariable(cb.parseExprNode(exprT.Var))
		if err != nil {
			log.Fatalf("Error in ExprPropertyFetch (var): %v", err)
		}
		prop, err := cb.readVariable(cb.parseExprNode(exprT.Prop))
		if err != nil {
			log.Fatalf("Error in ExprPropertyFetch (name): %v", err)
		}
		op := NewOpExprPropertyFetch(vr, prop, exprT.Position)
		cb.currBlock.AddInstructions(op)

		varName, _ := GetOperName(vr)
		propStr, ok := GetOperVal(prop).(*OperString)
		if varName != "" && ok {
			propFetchName := "<propfetch>" + varName[1:] + "->" + propStr.Val
			op.Result = NewOperVar(NewOperString(propFetchName), nil)
		}

		return op.Result
	case *ast.ExprStaticCall:
		class, err := cb.readVariable(cb.parseExprNode(exprT.Class))
		if err != nil {
			log.Fatalf("Error in ExprStaticCall (class): %v", err)
		}
		name, err := cb.readVariable(cb.parseExprNode(exprT.Call))
		if err != nil {
			log.Fatalf("Error in ExprStaticCall (name): %v", err)
		}
		args, argsPos := cb.parseExprList(exprT.Args, MODE_READ)
		op := NewOpExprStaticCall(class, name, args, exprT.Class.GetPosition(), exprT.Call.GetPosition(), argsPos, exprT.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprStaticPropertyFetch:
		classVar, err := cb.readVariable(cb.parseExprNode(exprT.Class))
		if err != nil {
			log.Fatalf("Error in ExprStaticCall (class): %v", err)
		}
		prop, err := cb.readVariable(cb.parseExprNode(exprT.Prop))
		if err != nil {
			log.Fatalf("Error in ExprStaticCall (name): %v", err)
		}
		op := NewOpExprStaticPropertyFetch(classVar, prop, exprT.Position)

		className, _ := GetOperName(classVar)
		propStr, ok := GetOperVal(prop).(*OperString)
		if className != "" && ok {
			propFetchName := "<staticpropfetch>" + className[1:] + "::" + propStr.Val
			op.Result = NewOperVar(NewOperString(propFetchName), nil)
		}

		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprTernary:
		return cb.parseExprTernary(exprT)
	case *ast.ExprYield:
		return cb.parseExprYield(exprT)
	case *ast.ExprShellExec:
		args, argsPos := cb.parseExprList(exprT.Parts, MODE_READ)
		argOp := NewOpExprConcatList(args, argsPos, exprT.Position)
		cb.currBlock.AddInstructions(argOp)
		funcCallOp := NewOpExprFunctionCall(NewOperString("shell_exec"), []Operand{argOp.Result}, exprT.Position, argsPos, exprT.Position)
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
	case *ast.ExprMatch, *ast.ExprYieldFrom, *ast.ExprThrow:
		log.Fatal("Error: Cannot parse expression node, wrong type '", reflect.TypeOf(expr), "'")
	default:
		log.Fatal("Error: Cannot parse expression node, wrong type '", reflect.TypeOf(expr), "'")
	}

	return nil
}

// function to parse ast.ExprAssign into OpExprAssign
// return right side of assignment operation
func (cb *CfgBuilder) parseExprAssign(expr *ast.ExprAssign) Operand {
	right, err := cb.readVariable(cb.parseExprNode(expr.Expr))
	if err != nil {
		log.Fatalf("Error in parseExprAssign: %v", err)
	}

	// if var a list or array, do list assignment
	switch e := expr.Var.(type) {
	case *ast.ExprList:
		cb.parseAssignList(e.Items, right, e.Position)
		return right
	case *ast.ExprArray:
		cb.parseAssignList(e.Items, right, e.Position)
		return right
	}

	leftNode := cb.parseExprNode(expr.Var)
	left := cb.writeVariable(leftNode)
	op := NewOpExprAssign(left, right, expr.Var.GetPosition(), expr.Expr.GetPosition(), expr.Position)
	cb.currBlock.AddInstructions(op)

	// if right expr is a literal or object
	switch rv := GetOperVal(right).(type) {
	case *OperBool, *OperObject, *OperString, *OperSymbolic, *OperNumber:
		op.Result = rv
		SetOperVal(left, rv)
	}

	return op.Result
}

func (cb *CfgBuilder) parseAssignList(items []ast.Vertex, arrVar Operand, pos *position.Position) {
	var err error
	cnt := 0
	for _, item := range items {
		if item == nil {
			continue
		}
		var key Operand = nil
		arrItem := item.(*ast.ExprArrayItem)
		if arrItem.Val == nil {
			continue
		}

		// if no key, set key to cnt (considered as array)
		if arrItem.Key == nil {
			key = NewOperNumber(float64(cnt))
			cnt += 1
		} else {
			key, err = cb.readVariable(cb.parseExprNode(arrItem.Key))
			if err != nil {
				log.Fatalf("Error in parseAssignList (key): %v", err)
			}
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
		assign := NewOpExprAssign(left, fetch.Result, vr.GetPosition(), fetch.Position, pos)
		cb.currBlock.AddInstructions(assign)
	}
}

// function to parse ast.ExprAssignReference into OpExprAssignRef
// return right side of assignment operation
func (cb *CfgBuilder) parseExprAssignRef(expr *ast.ExprAssignReference) Operand {
	left := cb.writeVariable(cb.parseExprNode(expr.Var))
	right, err := cb.readVariable(cb.parseExprNode(expr.Expr))
	if err != nil {
		log.Fatalf("Error in parseExprAssignRef: %v", err)
	}

	assign := NewOpExprAssignRef(left, right, expr.Position)
	return assign.Result
}

// function to parse other ast.ExprAssign... into OpExprAssign...
// return right side of assignment operation
func (cb *CfgBuilder) parseExprAssignOp(expr ast.Vertex) Operand {
	var vr, e Operand
	var read, write Operand
	var err error
	switch exprT := expr.(type) {
	case *ast.ExprAssignBitwiseAnd:
		vr = cb.parseExprNode(exprT.Var)
		read, err = cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignBitwiseAnd: %v", err)
		}
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryBitwiseAnd(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		cb.currBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignBitwiseOr:
		vr = cb.parseExprNode(exprT.Var)
		read, err = cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignBitwiseOr: %v", err)
		}
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryBitwiseOr(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		cb.currBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignBitwiseXor:
		vr = cb.parseExprNode(exprT.Var)
		read, err = cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignBitwiseXor: %v", err)
		}
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryBitwiseXor(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		cb.currBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignConcat:
		vr = cb.parseExprNode(exprT.Var)
		read, err = cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignConcat: %v", err)
		}
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryConcat(read, e, exprT.Var.GetPosition(), exprT.Expr.GetPosition(), exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		cb.currBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignCoalesce:
		vr = cb.parseExprNode(exprT.Var)
		read, err = cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignCoalesce: %v", err)
		}
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryCoalesce(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		cb.currBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignDiv:
		vr = cb.parseExprNode(exprT.Var)
		read, err = cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignDiv: %v", err)
		}
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryDiv(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		cb.currBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignMinus:
		vr = cb.parseExprNode(exprT.Var)
		read, err = cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignMinus: %v", err)
		}
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryMinus(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		cb.currBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignMod:
		vr = cb.parseExprNode(exprT.Var)
		read, err = cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignMod: %v", err)
		}
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryMod(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		cb.currBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignPlus:
		vr = cb.parseExprNode(exprT.Var)
		read, err = cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignPlus: %v", err)
		}
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryPlus(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		cb.currBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignMul:
		vr = cb.parseExprNode(exprT.Var)
		read, err = cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignMul: %v", err)
		}
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryMul(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		cb.currBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignPow:
		vr = cb.parseExprNode(exprT.Var)
		read, err = cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignPow: %v", err)
		}
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryPow(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		cb.currBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignShiftLeft:
		vr = cb.parseExprNode(exprT.Var)
		read, err = cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignShiftLeft: %v", err)
		}
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryShiftLeft(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		cb.currBlock.AddInstructions(assign)
		return op.Result
	case *ast.ExprAssignShiftRight:
		vr = cb.parseExprNode(exprT.Var)
		read, err = cb.readVariable(vr)
		if err != nil {
			log.Fatalf("Error in ExprAssignShiftRight: %v", err)
		}
		write = cb.writeVariable(vr)
		e = cb.parseExprNode(exprT.Expr)
		op := NewOpExprBinaryShiftRight(read, e, exprT.Position)
		cb.currBlock.AddInstructions(op)
		assign := NewOpExprAssign(write, op.Result, exprT.Var.GetPosition(), op.Position, exprT.Position)
		cb.currBlock.AddInstructions(assign)
		return op.Result
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
		nameVar, err := cb.readVariable(cb.parseExprNode(eu.Var))
		if err != nil {
			log.Fatalf("Error in parseExprClosure: %v", err)
		}
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
	cb.Script.AddFunc(opFunc)

	// create op closure
	closure := NewOpExprClosure(opFunc, uses, expr.Position)
	opFunc.CallableOp = closure

	cb.currBlock.AddInstructions(closure)
	return closure.Result
}

func (cb *CfgBuilder) parseExprArrowFunc(expr *ast.ExprArrowFunction) Operand {
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
	stmtExpr := &ast.StmtExpression{
		Position: expr.Expr.GetPosition(),
		Expr:     expr.Expr,
	}
	stmts := []ast.Vertex{stmtExpr}
	cb.parseOpFunc(opFunc, expr.Params, stmts)
	cb.Script.AddFunc(opFunc)

	// create op closure
	closure := NewOpExprClosure(opFunc, nil, expr.Position)
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
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryBitwiseAnd (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryBitwiseAnd (right): %v", err)
		}
		op := NewOpExprBinaryBitwiseAnd(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryBitwiseOr:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryBitwiseOr (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryBitwiseOr (right): %v", err)
		}
		op := NewOpExprBinaryBitwiseOr(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryBitwiseXor:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryBitwiseXor (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryBitwiseXor (right): %v", err)
		}
		op := NewOpExprBinaryBitwiseXor(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryNotEqual:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryNotEqual (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryBitwiseAndNotEqual (right): %v", err)
		}
		op := NewOpExprBinaryNotEqual(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryCoalesce:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryCoalesce (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryCoalesce (right): %v", err)
		}
		op := NewOpExprBinaryCoalesce(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryConcat:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryConcat (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryConcat (right): %v", err)
		}
		op := NewOpExprBinaryConcat(left, right, e.Left.GetPosition(), e.Right.GetPosition(), e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryDiv:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryDiv (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryDiv (right): %v", err)
		}
		op := NewOpExprBinaryDiv(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryEqual:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryEqual (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryEqual (right): %v", err)
		}
		op := NewOpExprBinaryEqual(left, right, e.Position)

		// handle type assertion using gettype()
		if left.IsWritten() {
			leftOp, isLeftFuncCall := left.GetWriteOp().(*OpExprFunctionCall)
			rightStr, isRightString := GetStringOper(right)
			// left must be function call with name gettype
			// right must be a string
			funcName := ""
			if isLeftFuncCall {
				funcName, err = GetOperName(leftOp.Name)
				if err != nil {
					log.Fatalf("Error in parseBinaryExprNode: %v", err)
				}
			}
			if isLeftFuncCall && funcName == "gettype" && isRightString {
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
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryGreater (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryGreater (right): %v", err)
		}
		op := NewOpExprBinaryGreater(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryGreaterOrEqual:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryGreaterOrEqual (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryGreaterOrEqual (right): %v", err)
		}
		op := NewOpExprBinaryGreaterOrEqual(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryIdentical:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryIdentical (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryIdentical (right): %v", err)
		}
		op := NewOpExprBinaryIdentical(left, right, e.Position)

		// handle type assertion using gettype()
		if left.IsWritten() {
			leftOp, isLeftFuncCall := left.GetWriteOp().(*OpExprFunctionCall)
			rightStr, isRightString := GetStringOper(right)
			// left must be function call with name gettype
			// right must be a string
			funcName := ""
			if isLeftFuncCall {
				funcName, err = GetOperName(leftOp.Name)
				if err != nil {
					log.Fatalf("Error in parseBinaryExprNode: %v", err)
				}
			}
			if isLeftFuncCall && funcName == "gettype" && isRightString {
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
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryLogicalOr (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryLogicalOr (right): %v", err)
		}
		op := NewOpExprBinaryLogicalOr(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryBooleanOr:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryBooleanOr (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryBooleanOr (right): %v", err)
		}
		op := NewOpExprBinaryLogicalOr(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryLogicalAnd:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryLogicalAnd (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryLogicalAnd (right): %v", err)
		}
		op := NewOpExprBinaryLogicalAnd(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryBooleanAnd:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryBooleanAnd (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryBooleanAnd (right): %v", err)
		}
		op := NewOpExprBinaryLogicalAnd(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryLogicalXor:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryLogicalXor (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryLogicalXor (right): %v", err)
		}
		op := NewOpExprBinaryLogicalXor(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryMinus:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryMinus (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryMinus (right): %v", err)
		}
		op := NewOpExprBinaryMinus(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryMod:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryMod (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryMod (right): %v", err)
		}
		op := NewOpExprBinaryMod(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryMul:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryMul (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryMul (right): %v", err)
		}
		op := NewOpExprBinaryMul(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryNotIdentical:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryMul (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryMul (right): %v", err)
		}
		op := NewOpExprBinaryNotIdentical(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryPlus:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryplus (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryPlus (right): %v", err)
		}
		op := NewOpExprBinaryPlus(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryPow:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryPow (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryPow (right): %v", err)
		}
		op := NewOpExprBinaryPow(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryShiftLeft:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryShiftLeft (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryShiftLeft (right): %v", err)
		}
		op := NewOpExprBinaryShiftLeft(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinaryShiftRight:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinaryShiftRight (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinaryShiftRight (right): %v", err)
		}
		op := NewOpExprBinaryShiftRight(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinarySmaller:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinarySmaller (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinarySmaller (right): %v", err)
		}
		op := NewOpExprBinarySmaller(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinarySmallerOrEqual:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinarySmallerOrEqual (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinarySmallerOrEqual (right): %v", err)
		}
		op := NewOpExprBinarySmallerOrEqual(left, right, e.Position)
		cb.currBlock.AddInstructions(op)
		return op.Result
	case *ast.ExprBinarySpaceship:
		left, err := cb.readVariable(cb.parseExprNode(e.Left))
		if err != nil {
			log.Fatalf("Error in ExprBinarySpaceship (left): %v", err)
		}
		right, err := cb.readVariable(cb.parseExprNode(e.Right))
		if err != nil {
			log.Fatalf("Error in ExprBinarySpaceship (right): %v", err)
		}
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
			item, ok := arrItem.(*ast.ExprArrayItem)
			if !ok {
				log.Fatal(reflect.TypeOf(arrItem))
			}
			// empty item
			if item.Val == nil {
				continue
			}

			if item.Key != nil {
				key, err := cb.readVariable(cb.parseExprNode(item.Key))
				if err != nil {
					log.Fatalf("Error in parseExprArray (key): %v", err)
				}
				keys = append(keys, key)
			} else {
				keys = append(keys, NewOperNull())
			}

			val, err := cb.readVariable(cb.parseExprNode(item.Val))
			if err != nil {
				log.Fatalf("Error in parseExprArray (val): %v", err)
			}
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
	vr, err := cb.readVariable(cb.parseExprNode(expr.Var))
	if err != nil {
		log.Fatalf("Error in parseExprArrayDimFetch (var): %v", err)
	}
	var dim Operand
	if expr.Dim != nil {
		dim, err = cb.readVariable(cb.parseExprNode(expr.Dim))
		if err != nil {
			log.Fatalf("Error in parseExprArrayDimFetch (dim): %v", err)
		}
	} else {
		dim = NewOperNull()
	}

	op := NewOpExprArrayDimFetch(vr, dim, expr.Position)
	cb.currBlock.AddInstructions(op)

	// if vs, ok := vr.(*OperSymbolic); ok && vs.Val != "undefined" {
	// 	// symbolic interpreter
	// 	op.Result = NewOperSymbolic(vs.Val, true)
	// 	return op.Result
	// }

	// if result, ok := op.Result.(*OperTemporary); ok && !varDefined {
	// 	varName, _ := GetOperName(vr)
	// 	dimStr, ok := GetOperVal(dim).(*OperString)
	// 	if varName != "" && ok {
	// 		arrayDimName := "<arraydimfetch>" + varName[1:] + "[" + dimStr.Val + "]"
	// 		result.Original = NewOperVar(NewOperString(arrayDimName), nil)
	// 	}
	// }

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
	var err error
	if expr.Expr != nil {
		e, err = cb.readVariable(cb.parseExprNode(expr.Expr))
		if err != nil {
			log.Fatalf("Error in parseExprExit (expr): %v", err)
		}
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
	args, argsPos := cb.parseExprList(expr.Args, MODE_READ)
	name, err := cb.readVariable(cb.parseExprNode(expr.Function))
	if err != nil {
		log.Fatalf("Error in parseExprFuncCall (name): %v", err)
	}
	// TODO: check if need NsFuncCall
	opFuncCall := NewOpExprFunctionCall(name, args, expr.Function.GetPosition(), argsPos, expr.Position)

	if nameStr, ok := name.(*OperString); ok {
		if tp, ok := GetTypeAssertFunc(nameStr.Val); ok {
			assert := NewTypeAssertion(NewOperString(tp), false)
			opFuncCall.Result.AddAssertion(args[0], assert, ASSERT_MODE_INTERSECT)
		} else if nameStr.Val == "settype" {
			read, err := cb.readVariable(opFuncCall.Args[0])
			if err != nil {
				log.Fatalf("Error in ExprFuncCall: %v", err)
			}
			write := cb.writeVariable(opFuncCall.Args[0])
			tp := opFuncCall.Args[1]
			if tpStr, ok := GetOperVal(tp).(*OperString); ok {
				switch tpStr.Val {
				case "boolean", "bool":
					op := NewOpExprCastBool(read, nil)
					cb.currBlock.AddInstructions(op)
					assign := NewOpExprAssign(write, op.Result, nil, op.Position, nil)
					cb.currBlock.AddInstructions(assign)
				case "integer", "int":
					op := NewOpExprCastInt(read, nil)
					cb.currBlock.AddInstructions(op)
					assign := NewOpExprAssign(write, op.Result, nil, op.Position, nil)
					cb.currBlock.AddInstructions(assign)
				case "float", "double":
					op := NewOpExprCastDouble(read, nil)
					cb.currBlock.AddInstructions(op)
					assign := NewOpExprAssign(write, op.Result, nil, op.Position, nil)
					cb.currBlock.AddInstructions(assign)
				case "string":
					op := NewOpExprCastString(read, nil)
					cb.currBlock.AddInstructions(op)
					assign := NewOpExprAssign(write, op.Result, nil, op.Position, nil)
					cb.currBlock.AddInstructions(assign)
				case "array":
					op := NewOpExprCastArray(read, nil)
					cb.currBlock.AddInstructions(op)
					assign := NewOpExprAssign(write, op.Result, nil, op.Position, nil)
					cb.currBlock.AddInstructions(assign)
				case "object":
					op := NewOpExprCastObject(read, nil)
					cb.currBlock.AddInstructions(op)
					assign := NewOpExprAssign(write, op.Result, nil, op.Position, nil)
					cb.currBlock.AddInstructions(assign)
				case "null":
					op := NewOpExprCastUnset(read, nil)
					cb.currBlock.AddInstructions(op)
					assign := NewOpExprAssign(write, op.Result, nil, op.Position, nil)
					cb.currBlock.AddInstructions(assign)
				}
			}
		}
	}

	cb.currBlock.AddInstructions(opFuncCall)
	cb.currFunc.Calls = append(cb.currFunc.Calls, opFuncCall)

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

	args, _ := cb.parseExprList(expr.Args, MODE_READ)
	opNew := NewOpExprNew(className, args, expr.Position)
	cb.currBlock.AddInstructions(opNew)

	// set result type to object operand
	if _, isString := className.(*OperString); isString {
		opNew.Result = NewOperObject(className.(*OperString).Val)
	}

	return opNew.Result
}

func (cb *CfgBuilder) parseExprTernary(expr *ast.ExprTernary) Operand {
	cond, err := cb.readVariable(cb.parseExprNode(expr.Cond))
	if err != nil {
		log.Fatalf("Error in parseExprTernary (cond): %v", err)
	}
	ifBlock := NewBlock(cb.GetBlockId())
	elseBlock := NewBlock(cb.GetBlockId())
	endBlock := NewBlock(cb.GetBlockId())

	jmpIf := NewOpStmtJumpIf(cond, ifBlock, elseBlock, expr.Position)
	cb.currBlock.AddInstructions(jmpIf)
	cb.currBlock.IsConditional = true
	cb.processAssertion(cond, ifBlock, elseBlock)
	ifBlock.AddPredecessor(cb.currBlock)
	elseBlock.AddPredecessor(cb.currBlock)

	// add condition to if block
	cb.Ctx.PushCond(cond)
	ifBlock.SetCondition(cb.Ctx.CurrConds)
	// build ifTrue block
	cb.currBlock = ifBlock
	ifVar := NewOperTemporary(nil)
	var ifAssignOp *OpExprAssign
	// if there is ifTrue value, assign ifVar with it
	// else, assign with 1
	if expr.IfTrue != nil {
		ifVal, err := cb.readVariable(cb.parseExprNode(expr.IfTrue))
		if err != nil {
			log.Fatalf("Error in parseExprTernary (if): %v", err)
		}
		ifAssignOp = NewOpExprAssign(ifVar, ifVal, nil, expr.IfTrue.GetPosition(), expr.Position)
	} else {
		ifAssignOp = NewOpExprAssign(ifVar, NewOperNumber(1), nil, expr.Position, expr.Position)
	}
	cb.currBlock.AddInstructions(ifAssignOp)
	// add jump op to end block
	jmp := NewOpStmtJump(endBlock, expr.Position)
	cb.currBlock.AddInstructions(jmp)
	// return the condition
	cb.Ctx.PopCond()

	// add condition to else block
	negatedCond := NewOpExprBooleanNot(cond, nil).Result
	cb.Ctx.PushCond(negatedCond)
	elseBlock.SetCondition(cb.Ctx.CurrConds)
	// build ifFalse block
	cb.currBlock = elseBlock
	elseVar := NewOperTemporary(nil)
	elseVal, err := cb.readVariable(cb.parseExprNode(expr.IfFalse))
	if err != nil {
		log.Fatalf("Error in parseExprTernary (else): %v", err)
	}
	elseAssignOp := NewOpExprAssign(elseVar, elseVal, nil, expr.IfFalse.GetPosition(), expr.Position)
	cb.currBlock.AddInstructions(elseAssignOp)
	// add jump to end block
	jmp = NewOpStmtJump(endBlock, expr.Position)
	cb.currBlock.AddInstructions(jmp)
	endBlock.AddPredecessor(cb.currBlock)
	// return else block
	cb.Ctx.PopCond()

	// build end block
	cb.currBlock = endBlock
	result := NewOperTemporary(nil)
	phi := NewOpPhi(result, cb.currBlock, expr.Position)
	phi.AddOperand(ifVar)
	phi.AddOperand(elseVar)
	cb.currBlock.AddPhi(phi)

	// return phi
	return result
}

func (cb *CfgBuilder) parseExprYield(expr *ast.ExprYield) Operand {
	var key Operand
	var val Operand
	var err error

	// TODO: handle index key (0, 1, 2, etc)
	if expr.Key != nil {
		key, err = cb.readVariable(cb.parseExprNode(expr.Key))
		if err != nil {
			log.Fatalf("Error in parseExprYield (key): %v", err)
		}
	}
	if expr.Val != nil {
		val, err = cb.readVariable(cb.parseExprNode(expr.Val))
		if err != nil {
			log.Fatalf("Error in parseExprYield (val): %v", err)
		}
	}

	yieldOp := NewOpExprYield(val, key, expr.Position)
	cb.currBlock.AddInstructions(yieldOp)

	return yieldOp.Result
}

// function to parse list of expressions node
func (cb *CfgBuilder) parseExprList(exprs []ast.Vertex, mode VAR_MODE) ([]Operand, []*position.Position) {
	vars := make([]Operand, 0, len(exprs))
	positions := make([]*position.Position, 0, len(exprs))
	switch mode {
	case MODE_READ:
		for _, expr := range exprs {
			exprNode := cb.parseExprNode(expr)
			vr, err := cb.readVariable(exprNode)
			if err != nil {
				log.Fatalf("Error in parseExprList (var): %v", err)
			}
			vars = append(vars, vr)
			positions = append(positions, expr.GetPosition())
		}
	case MODE_WRITE:
		for _, expr := range exprs {
			vars = append(vars, cb.writeVariable(cb.parseExprNode(expr)))
			positions = append(positions, expr.GetPosition())
		}
	case MODE_NONE:
		for _, expr := range exprs {
			vars = append(vars, cb.parseExprNode(expr))
			positions = append(positions, expr.GetPosition())
		}
	}

	return vars, positions
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
			declaration, err := cb.readVariable(cb.parseExprNode(n))
			if err != nil {
				log.Fatalf("Error in parseTypeNode (declaration): %v", err)
			}
			return NewOpTypeReference(declaration, false, n.Position)
		}
	case *ast.NameFullyQualified:
		name, _ := astutil.GetNameString(n)
		if IsBuiltInType(name) {
			return NewOpTypeLiteral(name, false, n.Position)
		} else if name == "mixed" {
			return NewOpTypeMixed(n.Position)
		} else if name == "void" {
			return NewOpTypeVoid(n.Position)
		} else {
			declaration, err := cb.readVariable(cb.parseExprNode(n))
			if err != nil {
				log.Fatalf("Error in parseTypeNode (declaration): %v", err)
			}
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
		log.Fatalf("Error: invalid type node '%v'", reflect.TypeOf(n))
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
	leftVal, err := cb.readVariable(cb.parseExprNode(left))
	if err != nil {
		log.Fatalf("Error in parseShortCircuiting: %v", err)
	}

	// create jumpIf op and adding it as currBlock next instruction
	jmpIf := NewOpStmtJumpIf(leftVal, ifCond, elseCond, expr.GetPosition())
	cb.currBlock.AddInstructions(jmpIf)
	cb.currBlock.IsConditional = true
	longBlock.AddPredecessor(cb.currBlock)
	endBlock.AddPredecessor(cb.currBlock)

	// build long block
	cb.currBlock = longBlock
	rightVal, err := cb.readVariable(cb.parseExprNode(right))
	if err != nil {
		log.Fatalf("Error in parseShortCircuiting: %v", err)
	}
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
	block := cb.currBlock
	for _, assert := range oper.GetAssertions() {
		// add assertion into if block
		cb.currBlock = ifBlock
		read, err := cb.readVariable(assert.Var)
		if err != nil {
			log.Fatalf("Error in processAssertion (if): %v", err)
		}
		write := cb.writeVariable(assert.Var)
		a := cb.readAssertion(assert.Assert)
		opAssert := NewOpExprAssertion(read, write, a, nil)
		cb.currBlock.AddInstructions(opAssert)

		// add negation of the assertion into else block
		cb.currBlock = elseBlock
		read, err = cb.readVariable(assert.Var)
		if err != nil {
			log.Fatalf("Error in processAssertion (else): %v", err)
		}
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
		vr, err := cb.readVariable(a.Val)
		if err != nil {
			log.Fatalf("Error in readAssertion (if): %v", err)
		}
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
	for vrTemp, ok := vr.(*OperTemporary); ok && vrTemp.Original != nil; {
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
			// default:
			// 	log.Fatalf("Error: Invalid operand type '%v' for a variable name", reflect.TypeOf(vrVar.Name))
		}
	}

	return vr
}

func (cb *CfgBuilder) writeVariableName(name string, val Operand, block *Block) {
	cb.VarNames[name] = struct{}{}
	cb.Ctx.SetValueInScope(block, name, val)
}

// TODO: Check name type
// read defined variable
func (cb *CfgBuilder) readVariable(vr Operand) (Operand, error) {
	if vr == nil {
		return nil, fmt.Errorf("read nil operand")
	}
	// TODO: Code preprocess
	switch v := vr.(type) {
	case *OperBoundVar:
		return v, nil
	case *OperVariable:
		// if variable name is string, read variable name
		// else if a variable, it's a variable variables
		switch varName := v.Name.(type) {
		case *OperString:
			return cb.readVariableName(varName.Val, cb.currBlock), nil
		case *OperVariable:
			_, err := cb.readVariable(varName)
			if err != nil {
				return nil, err
			}
			return vr, nil
		case *OperTemporary:
			_, err := cb.readVariable(varName)
			if err != nil {
				return nil, err
			}
			return vr, nil
		default:
			log.Fatalf("Error variable name '%v'", reflect.TypeOf(varName))
		}
	case *OperTemporary:
		if v.Original != nil {
			res, err := cb.readVariable(v.Original)
			if err != nil {
				return nil, err
			}
			return res, nil
		}
	}

	return vr, nil
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
