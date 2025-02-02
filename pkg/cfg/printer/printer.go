package printer

import (
	"container/list"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg"
	"github.com/VKCOM/php-parser/pkg/position"
)

type Printer struct {
	Writer      io.Writer
	VarIds      map[cfg.Operand]int
	BlockIds    map[*cfg.Block]int
	BlockQueue  *list.List
	OperandsDef map[cfg.Operand]string

	varCnt   int
	blockCnt int
}

func PrintCfg(script *cfg.Script, w io.Writer) {
	p := &Printer{
		Writer:      w,
		VarIds:      make(map[cfg.Operand]int),
		BlockIds:    make(map[*cfg.Block]int),
		BlockQueue:  list.New(),
		OperandsDef: make(map[cfg.Operand]string),
		blockCnt:    1,
		varCnt:      1,
	}

	w.Write([]byte("Imported file:\n"))
	for _, importedFile := range script.IncludedFiles {
		w.Write([]byte(fmt.Sprintln(importedFile)))
	}
	w.Write([]byte("\n"))

	w.Write([]byte("Main"))
	if script.Main.ContaintTainted {
		w.Write([]byte("<tainted>\n"))
	} else {
		w.Write([]byte("\n"))
	}
	p.printFunc(script.Main)
	for _, fn := range script.OrderedFuncs {
		tp := p.renderType(fn.ReturnType)
		str := fmt.Sprintf("Function '%s': %s", fn.GetScopedName(), tp)
		w.Write([]byte(str))
		if fn.ContaintTainted {
			w.Write([]byte("<tainted>\n"))
		} else {
			w.Write([]byte("\n"))
		}

		p.printFunc(fn)
	}

	// print each operand
	w.Write([]byte("Operand Definition:\n"))
	for _, operDef := range p.OperandsDef {
		w.Write([]byte(operDef))
		w.Write([]byte("\n"))
	}
}

func (p *Printer) printFunc(fn *cfg.OpFunc) {
	// render function
	if fn.Cfg != nil {
		p.enqueueBlock(fn.Cfg)
	}

	blocks := make([]*cfg.Block, 0)
	blockDatas := make([]string, 0)

	for e := p.BlockQueue.Front(); e != nil; e = e.Next() {
		var sb strings.Builder
		block := e.Value.(*cfg.Block)
		blocks = append(blocks, block)

		// write all phis
		for phi := range block.Phi {
			// write variable
			operStr := p.RenderOperand(phi.Result)
			sb.WriteString(fmt.Sprintf("\n    %s = phi(", operStr))

			// write list of phi's vars
			phiVarsStr := make([]string, 0)
			for _, phiVar := range phi.GetVars() {
				phiVarsStr = append(phiVarsStr, p.RenderOperand(phiVar))
			}
			if len(phiVarsStr) > 0 {
				sb.WriteString(strings.Join(phiVarsStr, ", "))
			}

			sb.WriteString(")")
		}

		// write all ops
		for _, op := range block.Instructions {
			sb.WriteString(fmt.Sprintf("\n    %s", p.RenderOp(op)))
		}

		blockDatas = append(blockDatas, sb.String())
	}

	if len(blocks) != len(blockDatas) {
		log.Fatal("Error: Missing block metadata")
	}
	for i, block := range blocks {
		blockData := blockDatas[i]
		blockId := p.BlockIds[block]
		p.Writer.Write([]byte(fmt.Sprintf("Block#%d", blockId)))
		if block.ContaintTainted {
			p.Writer.Write([]byte("<tainted>"))
		}
		if block.IsConditional {
			p.Writer.Write([]byte("<conditional>"))
		}

		// write all conditions
		renderedConds := make([]string, 0, len(block.Conds))
		for _, cond := range block.Conds {
			renderedConds = append(renderedConds, p.RenderOperand(cond))
		}
		if len(renderedConds) > 0 {
			p.Writer.Write([]byte(indent("\nConditions: ", 1)))
			p.Writer.Write([]byte(strings.Join(renderedConds, ", ")))
		}

		// write all parents
		for _, pred := range block.Predecessors {
			if predId, ok := p.BlockIds[pred]; ok {
				p.Writer.Write([]byte(indent(fmt.Sprintf("\nParent: Block#%d", predId), 1)))
			}
		}

		p.Writer.Write([]byte(blockData))
		p.Writer.Write([]byte("\n\n"))
	}

	// print all sources
	p.Writer.Write([]byte("Sources:\n"))
	for _, source := range fn.Sources {
		p.Writer.Write([]byte(p.RenderOp(source)))
		p.Writer.Write([]byte("\n"))
	}

	p.reset()
}

func (p *Printer) RenderOp(op cfg.Op) string {
	var sb strings.Builder
	sb.WriteString(op.GetType())
	if IsSource(op) {
		sb.WriteString("<source>")
	}

	// render name of function and assertion
	switch o := op.(type) {
	case cfg.OpCallable:
		fn := o.GetFunc()
		sb.WriteString(fmt.Sprintf("<%s>", fn.Name))
	case *cfg.OpExprAssertion:
		sb.WriteString(fmt.Sprintf("<%s>", p.renderAssertion(o.Assertion)))
	}

	// render position
	sb.WriteString(p.renderPosition(op.GetPosition()))

	// render attribute groups
	switch o := op.(type) {
	case *cfg.OpStmtFunc:
		sb.WriteString(p.renderAttrGroups(o.AttrGroups))
	case *cfg.OpStmtClass:
		sb.WriteString(p.renderAttrGroups(o.AttrGroups))
	case *cfg.OpStmtClassMethod:
		sb.WriteString(p.renderAttrGroups(o.AttrGroups))
		sb.WriteString(fmt.Sprintf("\n        flags: %s", indent(p.renderModifiers(o), 1)))
	case *cfg.OpStmtProperty:
		sb.WriteString(p.renderAttrGroups(o.AttrGroups))
		// render modifier
		sb.WriteString(fmt.Sprintf("\n        flags: %s", indent(p.renderModifiers(o), 1)))
		// render type's property
		sb.WriteString(fmt.Sprintf("\n        declaredType: %s", indent(p.renderType(o.DeclaredType), 1)))
	case *cfg.OpExprParam:
		sb.WriteString(p.renderAttrGroups(o.AttrGroups))
		sb.WriteString(fmt.Sprintf("\n        declaredType: %s", indent(p.renderType(o.DeclaredType), 1)))
	case *cfg.OpExprInclude:
		sb.WriteString(fmt.Sprintf("\n        Type: %s", indent(o.IncludeTypeStr(), 1)))
	case *cfg.OpStmtTraitUse:
		// TODO
	}

	// render each variables in op
	for varName, varOpers := range op.GetOpListVars() {
		for i, varOper := range varOpers {
			sb.WriteString(fmt.Sprintf("\n        %s[%d]: %s", varName, i, indent(p.RenderOperand(varOper), 1)))

			// render operand's position
			sb.WriteString(indent(p.renderPosition(op.GetOpVarListPos(varName, i)), 1))
		}
	}
	for varName, varOper := range op.GetOpVars() {
		if varOper != nil {
			sb.WriteString(fmt.Sprintf("\n        %s: %s", varName, indent(p.RenderOperand(varOper), 1)))

			// render operand's position
			sb.WriteString(indent(p.renderPosition(op.GetOpVarPos(varName)), 1))
		} else if varName == "Class" {
			log.Fatal(reflect.TypeOf(op))
		}
	}

	// render sub blocks
	for subBlockName, subBlock := range cfg.GetSubBlocks(op) {
		p.enqueueBlock(subBlock)
		sb.WriteString(fmt.Sprintf("\n        %s: Block#%d", subBlockName, p.BlockIds[subBlock]))
	}

	return sb.String()
}

func (p *Printer) renderAssertion(assert cfg.Assertion) string {
	var sb strings.Builder
	if assert.Negated() {
		sb.WriteString("not(")
	}

	switch a := assert.(type) {
	case *cfg.TypeAssertion:
		sb.WriteString(fmt.Sprintf("type(%s)", p.RenderOperand(a.Val)))
	case *cfg.CompositeAssertion:
		combinator := "|"
		if a.Mode == cfg.ASSERT_MODE_INTERSECT {
			combinator = "&"
		}
		childAsserts := make([]string, 0, len(a.Val))
		for _, childAssert := range a.Val {
			childAsserts = append(childAsserts, p.renderAssertion(childAssert))
		}
		sb.WriteString("(")
		sb.WriteString(strings.Join(childAsserts, combinator))
		sb.WriteString(")")
	}

	if assert.Negated() {
		sb.WriteString(")")
	}

	return sb.String()
}

func (p *Printer) renderPosition(pos *position.Position) string {
	var sb strings.Builder

	if pos != nil {
		// sb.WriteString(fmt.Sprintf("\n        position['StartLine']: %d", pos.StartLine))
		// sb.WriteString(fmt.Sprintf("\n        position['EndLine']: %d", pos.EndLine))
		// sb.WriteString(fmt.Sprintf("\n        position['StartPos']: %d", pos.StartPos))
		// sb.WriteString(fmt.Sprintf("\n        position['EndPos']: %d", pos.EndPos))
	}

	return sb.String()
}

func (p *Printer) renderAttrGroups(attrGroups []*cfg.OpAttributeGroup) string {
	var sb strings.Builder

	for i, attrGroup := range attrGroups {
		sb.WriteString(fmt.Sprintf("\n    attrGroup[%d]: ", i))

		for attrIndex, attr := range attrGroup.Attrs {
			sb.WriteString(fmt.Sprintf("\n        attr[%d]", attrIndex))
			sb.WriteString(fmt.Sprintf("\n            name: %s", p.RenderOperand(attr.Name)))
			for argIndex, arg := range attr.Args {
				sb.WriteString(fmt.Sprintf("\n            args[%d]: %s", argIndex, p.RenderOperand(arg)))
			}
		}

	}

	return sb.String()
}

func (p *Printer) renderModifiers(op cfg.Op) string {
	var sb strings.Builder

	switch o := op.(type) {
	case *cfg.OpStmtProperty:
		if o.ReadOnly {
			sb.WriteString("readonly|")
		}
		if o.Static {
			sb.WriteString("static|")
		}
		if o.IsPrivate() {
			sb.WriteString("private")
		} else if o.IsProtected() {
			sb.WriteString("protected")
		} else {
			sb.WriteString("public")
		}
	case *cfg.OpStmtClassMethod:
		if o.Final {
			sb.WriteString("final|")
		}
		if o.Abstract {
			sb.WriteString("abstract|")
		}
		sb.WriteString(p.renderModifiers(o.Func))
	case *cfg.OpFunc:
		if o.IsPrivate() {
			sb.WriteString("private")
		} else if o.IsProtected() {
			sb.WriteString("protected")
		} else {
			sb.WriteString("public")
		}
	case *cfg.OpStmtClass:
		if o.IsPrivate() {
			sb.WriteString("private")
		} else if o.IsProtected() {
			sb.WriteString("protected")
		} else {
			sb.WriteString("public")
		}
	}

	return sb.String()
}

func (p *Printer) renderType(tp cfg.OpType) string {
	if tp == nil {
		return ""
	}

	var sb strings.Builder
	if tp.Nullable() {
		sb.WriteString("?")
	}
	switch t := tp.(type) {
	case *cfg.OpTypeLiteral:
		sb.WriteString(t.Name)
	case *cfg.OpTypeMixed:
		sb.WriteString("mixed")
	case *cfg.OpTypeReference:
		sb.WriteString("reference:")
		sb.WriteString(p.RenderOperand(t.Declaration))
	case *cfg.OpTypeUnion:
		if len(t.Types) > 0 {
			i := 0
			for ; i < len(t.Types)-1; i++ {
				sb.WriteString(p.renderType(t.Types[i]))
				sb.WriteString("|")
			}
			sb.WriteString(p.renderType(t.Types[i]))
		}
	case *cfg.OpTypeVoid:
		sb.WriteString("void")
	default:
		log.Fatal("Error: rendering unknown type")
	}

	return sb.String()
}

func (p *Printer) RenderOperand(oper cfg.Operand) string {
	var sb strings.Builder
	var operSb strings.Builder

	switch o := oper.(type) {
	case *cfg.OperBool:
		if o.Val {
			sb.WriteString("LITERAL('true')")
		} else {
			sb.WriteString("LITERAL('false')")
		}
	case *cfg.OperBoundVar:
		prefix := ""
		name := ""
		s := ""
		switch n := o.Name.(type) {
		case *cfg.OperString:
			name = n.Val
		case *cfg.OperVariable:
			name = p.RenderOperand(n.Name)
		case *cfg.OperBoundVar:
			name = p.RenderOperand(n.Name)
		default:
			log.Fatal("Error: Invalid variable name type")
		}
		if o.ByRef {
			prefix = "&"
		}
		switch o.Scope {
		case cfg.VAR_SCOPE_GLOBAL:
			s = fmt.Sprintf("global<%s%s>", prefix, name)
		case cfg.VAR_SCOPE_FUNCTION:
			s = fmt.Sprintf("static<%s%s>", prefix, name)
		case cfg.VAR_SCOPE_LOCAL:
			s = fmt.Sprintf("local<%s%s>", prefix, name)
		case cfg.VAR_SCOPE_OBJECT:
			s = fmt.Sprintf("this<%s%s>", prefix, name)
		default:
			s = fmt.Sprintf("%s%s", prefix, name)
		}
		sb.WriteString(s)
	case *cfg.OperNumber:
		s := strconv.FormatFloat(o.Val, 'f', -1, 64)
		sb.WriteString("LITERAL('")
		sb.WriteString(s)
		sb.WriteString("')")
	case *cfg.OperNull:
		sb.WriteString("NULL")
	case *cfg.OperString:
		sb.WriteString("LITERAL('")
		sb.WriteString(o.Val)
		sb.WriteString("')")
	case *cfg.OperSymbolic:
		sb.WriteString("SYMBOLIC('")
		sb.WriteString(o.Val)
		sb.WriteString("')")
	case *cfg.OperTemporary:
		id := p.getVarId(o)
		s := ""
		if o.Original != nil {
			s = fmt.Sprintf("Var#%d<%s>", id, p.RenderOperand(o.Original))
		} else {
			s = fmt.Sprintf("Var#%d", id)
		}

		if _, ok := p.OperandsDef[o]; !ok {
			operSb.WriteString(fmt.Sprintf("Var#%d\n", id))
			// write op
			operSb.WriteString("Write:\n")
			if opWrite := oper.GetWriteOp(); opWrite != nil {
				operSb.WriteString(fmt.Sprintf("%s\n", opWrite.GetType()))
			}
			// read op
			operSb.WriteString("Read:\n")
			for _, opUsage := range oper.GetUsage() {
				operSb.WriteString(fmt.Sprintf("%s\n", opUsage.GetType()))
			}
			p.OperandsDef[o] = operSb.String()
		}
		sb.WriteString(s)
	case *cfg.OperObject:
		sb.WriteString("OBJECT('")
		sb.WriteString(o.ClassName)
		sb.WriteString("')")
	case *cfg.OperVariable:
		prefix := "$"
		name := ""
		val := ""
		switch n := o.Name.(type) {
		case *cfg.OperString:
			name = n.Val
			if n.Val[0] == '$' {
				name = name[1:]
			} else {
				prefix = ""
			}
		case *cfg.OperVariable:
			name = p.RenderOperand(n)
		case *cfg.OperBoundVar:
			name = p.RenderOperand(n)
		default:
			log.Fatal("Error: Invalid variable name type")
		}
		switch o.Value.(type) {
		case *cfg.OperString, *cfg.OperBool, *cfg.OperNumber, *cfg.OperObject, *cfg.OperNull, *cfg.OperSymbolic:
			val = p.RenderOperand(o.Value)
		default:
			log.Fatalf("Error: Invalid variable name type '%v' in renderOperand", reflect.TypeOf(o.Value))
		}
		sb.WriteString(fmt.Sprintf("name(%s%s):val(%s)", prefix, name, val))
	}
	if oper.IsTainted() {
		sb.WriteString("<tainted>")
	}

	return sb.String()
}

func (p *Printer) getVarId(vr cfg.Operand) int {
	id, ok := p.VarIds[vr]
	if !ok {
		p.VarIds[vr] = p.varCnt
		p.varCnt += 1
		return p.VarIds[vr]
	}

	return id
}

func (p *Printer) reset() {
	p.VarIds = make(map[cfg.Operand]int)
	p.BlockIds = make(map[*cfg.Block]int)
	p.BlockQueue = list.New()
	p.blockCnt = 1
	p.varCnt = 1
}

func (p *Printer) enqueueBlock(block *cfg.Block) {
	if block == nil {
		return
	}
	if _, ok := p.BlockIds[block]; !ok {
		p.BlockIds[block] = p.blockCnt
		p.blockCnt += 1
		p.BlockQueue.PushBack(block)
	}
}

func (p *Printer) GetBlockId(block *cfg.Block) int {
	id, ok := p.BlockIds[block]
	if !ok {
		p.BlockIds[block] = p.blockCnt
		p.blockCnt += 1
		return p.BlockIds[block]
	}

	return id
}

func indent(s string, level int) string {
	if level > 1 {
		s = indent(s, level-1)
	}

	return strings.Replace(s, "\n", "\n    ", -1)
}

func IsSource(op cfg.Op) bool {
	// php source
	if assignOp, ok := op.(*cfg.OpExprAssign); ok {
		// symbolic interpreter ($_POST, $_GET, $_REQUEST, $_FILES, $_COOKIE, $_SERVERS)
		if result, ok := assignOp.Result.(*cfg.OperSymbolic); ok {
			switch result.Val {
			case "postsymbolic":
				fallthrough
			case "getsymbolic":
				fallthrough
			case "requestsymbolic":
				fallthrough
			case "filessymbolic":
				fallthrough
			case "cookiesymbolic":
				fallthrough
			case "serverssymbolic":
				return true
			}
		}
		// filter_input(), apache_request_headers(), getallheaders()
		if assignOp.Expr.IsWritten() {
			if right, ok := assignOp.Expr.GetWriteOp().(*cfg.OpExprFunctionCall); ok {
				funcNameStr, _ := cfg.GetOperName(right.Name)
				// filter_input
				if funcNameStr == "filter_input" {
					// TODO: check again the arguments
					return true
				} else if funcNameStr == "apache_request_headers" || funcNameStr == "getallheaders" {
					return true
				}
			}
		}
	}

	// TODO: laravel source

	return false
}
