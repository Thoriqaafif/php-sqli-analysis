package asttraverser

import (
	"fmt"
	"os"

	"github.com/VKCOM/php-parser/pkg/ast"
)

type ReturnedNodeType int

const (
	ReturnReplacedNode ReturnedNodeType = iota
	ReturnInsertedNode
)

type NodeTraverser interface {
	EnterNode(n ast.Vertex) (ast.Vertex, ReturnedNodeType)
	LeaveNode(n ast.Vertex) (ast.Vertex, ReturnedNodeType)
}

type InsertedNode struct {
	Idx  int
	Node ast.Vertex
}

type Traverser struct {
	NodeTravs []NodeTraverser
}

func NewTraverser() *Traverser {
	return &Traverser{
		NodeTravs: make([]NodeTraverser, 0),
	}
}

func (t *Traverser) AddNodeTraverser(nts ...NodeTraverser) {
	t.NodeTravs = append(t.NodeTravs, nts...)
}

func (t *Traverser) Traverse(n ast.Vertex) ast.Vertex {
	fmt.Println(n)
	if n != nil {
		// Enter Node
		for _, nt := range t.NodeTravs {
			replacedNode, nType := nt.EnterNode(n)
			if replacedNode != nil {
				if nType == ReturnReplacedNode && validReplacement(n, replacedNode) {
					return replacedNode
				} else {
					fmt.Println("Invalid node replacement")
					os.Exit(1)
				}
			}
		}

		n.Accept(t)

		// Leave Node
		for _, nt := range t.NodeTravs {
			replacedNode, nType := nt.LeaveNode(n)
			if replacedNode != nil {
				if nType == ReturnReplacedNode && validReplacement(n, replacedNode) {
					return replacedNode
				} else {
					fmt.Println("Invalid node replacement")
					os.Exit(1)
				}
			}
		}
	}

	return nil
}

func (t *Traverser) TraverseNodes(ns []ast.Vertex) {
	var insertedNodes []InsertedNode = make([]InsertedNode, 0)

	for i, n := range ns {
		// Enter Node
		for _, nt := range t.NodeTravs {
			returnedNode, nType := nt.EnterNode(n)
			if returnedNode != nil {
				if nType == ReturnReplacedNode {
					if validReplacement(n, returnedNode) {
						ns[i] = returnedNode
					} else {
						fmt.Println("Invalid node replacement")
						os.Exit(1)
					}
				} else {
					fmt.Println("Error while traversing array of nodes")
					os.Exit(1)
				}
			}
		}

		n.Accept(t)

		// Leave Node
		for _, nt := range t.NodeTravs {
			returnedNode, nType := nt.LeaveNode(n)
			if returnedNode != nil {
				if nType == ReturnReplacedNode {
					if validReplacement(n, returnedNode) {
						ns[i] = returnedNode
					} else {
						fmt.Println("Invalid node replacement")
						os.Exit(1)
					}
				} else if nType == ReturnInsertedNode {
					insertedNodes = append(insertedNodes, InsertedNode{Idx: i, Node: returnedNode})
				} else {
					fmt.Println("Error while traversing array of nodes")
					os.Exit(1)
				}
			}
		}
	}

	// inserting nodes
	for i := len(insertedNodes) - 1; i >= 0; i++ {
		idx := insertedNodes[i].Idx
		node := insertedNodes[i].Node

		left := ns[:idx]
		left = append(left, node)
		right := ns[idx+1:]

		ns = append(left, right...)
	}
}

func (t *Traverser) Root(n *ast.Root) {
	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) Nullable(n *ast.Nullable) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) Parameter(n *ast.Parameter) {
	t.TraverseNodes(n.AttrGroups)

	t.TraverseNodes(n.Modifiers)

	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.DefaultValue); replacedNode != nil {
		n.DefaultValue = replacedNode
	}
}

func (t *Traverser) Identifier(n *ast.Identifier) {
	// do nothing
}

func (t *Traverser) Argument(n *ast.Argument) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) MatchArm(n *ast.MatchArm) {
	t.TraverseNodes(n.Exprs)

	if replacedNode := t.Traverse(n.ReturnExpr); replacedNode != nil {
		n.ReturnExpr = replacedNode
	}
}

func (t *Traverser) Union(n *ast.Union) {
	t.TraverseNodes(n.Types)
}

func (t *Traverser) Intersection(n *ast.Intersection) {
	t.TraverseNodes(n.Types)
}

func (t *Traverser) Attribute(n *ast.Attribute) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	t.TraverseNodes(n.Args)
}

func (t *Traverser) AttributeGroup(n *ast.AttributeGroup) {
	t.TraverseNodes(n.Attrs)
}

func (t *Traverser) StmtBreak(n *ast.StmtBreak) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtCase(n *ast.StmtCase) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}

	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtCatch(n *ast.StmtCatch) {
	t.TraverseNodes(n.Types)

	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}

	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtEnum(n *ast.StmtEnum) {
	t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}

	t.TraverseNodes(n.Implements)
	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) EnumCase(n *ast.EnumCase) {
	t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtClass(n *ast.StmtClass) {
	t.TraverseNodes(n.AttrGroups)
	t.TraverseNodes(n.Modifiers)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	t.TraverseNodes(n.Args)
	t.TraverseNodes(n.Implements)

	if replacedNode := t.Traverse(n.Extends); replacedNode != nil {
		n.Extends = replacedNode
	}

	t.TraverseNodes(n.Implements)
	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtClassConstList(n *ast.StmtClassConstList) {
	t.TraverseNodes(n.AttrGroups)
	t.TraverseNodes(n.Modifiers)
	t.TraverseNodes(n.Consts)
}

func (t *Traverser) StmtClassMethod(n *ast.StmtClassMethod) {
	t.TraverseNodes(n.AttrGroups)
	t.TraverseNodes(n.Modifiers)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	t.TraverseNodes(n.Params)

	if replacedNode := t.Traverse(n.ReturnType); replacedNode != nil {
		n.ReturnType = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *Traverser) StmtConstList(n *ast.StmtConstList) {
	t.TraverseNodes(n.Consts)
}

func (t *Traverser) StmtConstant(n *ast.StmtConstant) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtContinue(n *ast.StmtContinue) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtDeclare(n *ast.StmtDeclare) {
	t.TraverseNodes(n.Consts)

	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *Traverser) StmtDefault(n *ast.StmtDefault) {
	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtDo(n *ast.StmtDo) {
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
}

func (t *Traverser) StmtEcho(n *ast.StmtEcho) {
	t.TraverseNodes(n.Exprs)
}

func (t *Traverser) StmtElse(n *ast.StmtElse) {
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *Traverser) StmtElseIf(n *ast.StmtElseIf) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *Traverser) StmtExpression(n *ast.StmtExpression) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtFinally(n *ast.StmtFinally) {
	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtFor(n *ast.StmtFor) {
	t.TraverseNodes(n.Init)
	t.TraverseNodes(n.Cond)
	t.TraverseNodes(n.Loop)

	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *Traverser) StmtForeach(n *ast.StmtForeach) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
	if replacedNode := t.Traverse(n.Key); replacedNode != nil {
		n.Key = replacedNode
	}
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *Traverser) StmtFunction(n *ast.StmtFunction) {
	t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	t.TraverseNodes(n.Params)

	if replacedNode := t.Traverse(n.ReturnType); replacedNode != nil {
		n.ReturnType = replacedNode
	}

	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtGlobal(n *ast.StmtGlobal) {
	t.TraverseNodes(n.Vars)
}

func (t *Traverser) StmtGoto(n *ast.StmtGoto) {
	if replacedNode := t.Traverse(n.Label); replacedNode != nil {
		n.Label = replacedNode
	}
}

func (t *Traverser) StmtHaltCompiler(n *ast.StmtHaltCompiler) {
	// Do Nothing
}

func (t *Traverser) StmtIf(n *ast.StmtIf) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}

	t.TraverseNodes(n.ElseIf)

	if replacedNode := t.Traverse(n.Else); replacedNode != nil {
		n.Else = replacedNode
	}
}

func (t *Traverser) StmtInlineHtml(n *ast.StmtInlineHtml) {
	// Do Nothing
}

func (t *Traverser) StmtInterface(n *ast.StmtInterface) {
	t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	t.TraverseNodes(n.Extends)
	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtLabel(n *ast.StmtLabel) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
}

func (t *Traverser) StmtNamespace(n *ast.StmtNamespace) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtNop(n *ast.StmtNop) {
	// Do Nothing
}

func (t *Traverser) StmtProperty(n *ast.StmtProperty) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtPropertyList(n *ast.StmtPropertyList) {
	t.TraverseNodes(n.AttrGroups)
	t.TraverseNodes(n.Modifiers)

	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}

	t.TraverseNodes(n.Props)
}

func (t *Traverser) StmtReturn(n *ast.StmtReturn) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtStatic(n *ast.StmtStatic) {
	t.TraverseNodes(n.Vars)
}

func (t *Traverser) StmtStaticVar(n *ast.StmtStaticVar) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtStmtList(n *ast.StmtStmtList) {
	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtSwitch(n *ast.StmtSwitch) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}

	t.TraverseNodes(n.Cases)
}

func (t *Traverser) StmtThrow(n *ast.StmtThrow) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) StmtTrait(n *ast.StmtTrait) {
	t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtTraitUse(n *ast.StmtTraitUse) {
	t.TraverseNodes(n.Traits)
	t.TraverseNodes(n.Adaptations)
}

func (t *Traverser) StmtTraitUseAlias(n *ast.StmtTraitUseAlias) {
	if replacedNode := t.Traverse(n.Trait); replacedNode != nil {
		n.Trait = replacedNode
	}
	if replacedNode := t.Traverse(n.Method); replacedNode != nil {
		n.Method = replacedNode
	}
	if replacedNode := t.Traverse(n.Modifier); replacedNode != nil {
		n.Modifier = replacedNode
	}
	if replacedNode := t.Traverse(n.Alias); replacedNode != nil {
		n.Alias = replacedNode
	}
}

func (t *Traverser) StmtTraitUsePrecedence(n *ast.StmtTraitUsePrecedence) {
	if replacedNode := t.Traverse(n.Trait); replacedNode != nil {
		n.Trait = replacedNode
	}
	if replacedNode := t.Traverse(n.Method); replacedNode != nil {
		n.Method = replacedNode
	}
	t.TraverseNodes(n.Insteadof)
}

func (t *Traverser) StmtTry(n *ast.StmtTry) {
	t.TraverseNodes(n.Stmts)
	t.TraverseNodes(n.Catches)

	if replacedNode := t.Traverse(n.Finally); replacedNode != nil {
		n.Finally = replacedNode
	}
}

func (t *Traverser) StmtUnset(n *ast.StmtUnset) {
	t.TraverseNodes(n.Vars)
}

func (t *Traverser) StmtUse(n *ast.StmtUseList) {
	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}
	t.TraverseNodes(n.Uses)
}

func (t *Traverser) StmtGroupUse(n *ast.StmtGroupUseList) {
	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}
	if replacedNode := t.Traverse(n.Prefix); replacedNode != nil {
		n.Prefix = replacedNode
	}

	t.TraverseNodes(n.Uses)
}

func (t *Traverser) StmtUseDeclaration(n *ast.StmtUse) {
	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}
	if replacedNode := t.Traverse(n.Use); replacedNode != nil {
		n.Use = replacedNode
	}
	if replacedNode := t.Traverse(n.Alias); replacedNode != nil {
		n.Alias = replacedNode
	}
}

func (t *Traverser) StmtWhile(n *ast.StmtWhile) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *Traverser) ExprArray(n *ast.ExprArray) {
	t.TraverseNodes(n.Items)
}

func (t *Traverser) ExprArrayDimFetch(n *ast.ExprArrayDimFetch) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Dim); replacedNode != nil {
		n.Dim = replacedNode
	}
}

func (t *Traverser) ExprArrayItem(n *ast.ExprArrayItem) {
	if replacedNode := t.Traverse(n.Key); replacedNode != nil {
		n.Key = replacedNode
	}
	if replacedNode := t.Traverse(n.Val); replacedNode != nil {
		n.Val = replacedNode
	}
}

func (t *Traverser) ExprArrowFunction(n *ast.ExprArrowFunction) {
	t.TraverseNodes(n.AttrGroups)
	t.TraverseNodes(n.Params)

	if replacedNode := t.Traverse(n.ReturnType); replacedNode != nil {
		n.ReturnType = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprBitwiseNot(n *ast.ExprBitwiseNot) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprBooleanNot(n *ast.ExprBooleanNot) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprBrackets(n *ast.ExprBrackets) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprClassConstFetch(n *ast.ExprClassConstFetch) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
	if replacedNode := t.Traverse(n.Const); replacedNode != nil {
		n.Const = replacedNode
	}
}

func (t *Traverser) ExprClone(n *ast.ExprClone) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprClosure(n *ast.ExprClosure) {
	t.TraverseNodes(n.AttrGroups)
	t.TraverseNodes(n.Params)
	t.TraverseNodes(n.Uses)

	if replacedNode := t.Traverse(n.ReturnType); replacedNode != nil {
		n.ReturnType = replacedNode
	}

	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) ExprClosureUse(n *ast.ExprClosureUse) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *Traverser) ExprConstFetch(n *ast.ExprConstFetch) {
	if replacedNode := t.Traverse(n.Const); replacedNode != nil {
		n.Const = replacedNode
	}
}

func (t *Traverser) ExprEmpty(n *ast.ExprEmpty) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprErrorSuppress(n *ast.ExprErrorSuppress) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprEval(n *ast.ExprEval) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprExit(n *ast.ExprExit) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprFunctionCall(n *ast.ExprFunctionCall) {
	if replacedNode := t.Traverse(n.Function); replacedNode != nil {
		n.Function = replacedNode
	}

	t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprInclude(n *ast.ExprInclude) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprIncludeOnce(n *ast.ExprIncludeOnce) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprInstanceOf(n *ast.ExprInstanceOf) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
}

func (t *Traverser) ExprIsset(n *ast.ExprIsset) {
	t.TraverseNodes(n.Vars)
}

func (t *Traverser) ExprList(n *ast.ExprList) {
	t.TraverseNodes(n.Items)
}

func (t *Traverser) ExprMethodCall(n *ast.ExprMethodCall) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Method); replacedNode != nil {
		n.Method = replacedNode
	}

	t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprNullsafeMethodCall(n *ast.ExprNullsafeMethodCall) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Method); replacedNode != nil {
		n.Method = replacedNode
	}

	t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprNew(n *ast.ExprNew) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}

	t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprPostDec(n *ast.ExprPostDec) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *Traverser) ExprPostInc(n *ast.ExprPostInc) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *Traverser) ExprPreDec(n *ast.ExprPreDec) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *Traverser) ExprPreInc(n *ast.ExprPreInc) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *Traverser) ExprPrint(n *ast.ExprPrint) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprPropertyFetch(n *ast.ExprPropertyFetch) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Prop); replacedNode != nil {
		n.Prop = replacedNode
	}
}

func (t *Traverser) ExprNullsafePropertyFetch(n *ast.ExprNullsafePropertyFetch) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Prop); replacedNode != nil {
		n.Prop = replacedNode
	}
}

func (t *Traverser) ExprRequire(n *ast.ExprRequire) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprRequireOnce(n *ast.ExprRequireOnce) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprShellExec(n *ast.ExprShellExec) {
	t.TraverseNodes(n.Parts)
}

func (t *Traverser) ExprStaticCall(n *ast.ExprStaticCall) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
	if replacedNode := t.Traverse(n.Call); replacedNode != nil {
		n.Call = replacedNode
	}

	t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprStaticPropertyFetch(n *ast.ExprStaticPropertyFetch) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
	if replacedNode := t.Traverse(n.Prop); replacedNode != nil {
		n.Prop = replacedNode
	}
}

func (t *Traverser) ExprTernary(n *ast.ExprTernary) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
	if replacedNode := t.Traverse(n.IfTrue); replacedNode != nil {
		n.IfTrue = replacedNode
	}
	if replacedNode := t.Traverse(n.IfFalse); replacedNode != nil {
		n.IfFalse = replacedNode
	}
}

func (t *Traverser) ExprUnaryMinus(n *ast.ExprUnaryMinus) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprUnaryPlus(n *ast.ExprUnaryPlus) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprVariable(n *ast.ExprVariable) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
}

func (t *Traverser) ExprYield(n *ast.ExprYield) {
	if replacedNode := t.Traverse(n.Key); replacedNode != nil {
		n.Key = replacedNode
	}
	if replacedNode := t.Traverse(n.Val); replacedNode != nil {
		n.Val = replacedNode
	}
}

func (t *Traverser) ExprYieldFrom(n *ast.ExprYieldFrom) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssign(n *ast.ExprAssign) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignReference(n *ast.ExprAssignReference) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignBitwiseAnd(n *ast.ExprAssignBitwiseAnd) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignBitwiseOr(n *ast.ExprAssignBitwiseOr) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignBitwiseXor(n *ast.ExprAssignBitwiseXor) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignCoalesce(n *ast.ExprAssignCoalesce) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignConcat(n *ast.ExprAssignConcat) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignDiv(n *ast.ExprAssignDiv) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignMinus(n *ast.ExprAssignMinus) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignMod(n *ast.ExprAssignMod) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignMul(n *ast.ExprAssignMul) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignPlus(n *ast.ExprAssignPlus) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignPow(n *ast.ExprAssignPow) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignShiftLeft(n *ast.ExprAssignShiftLeft) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprAssignShiftRight(n *ast.ExprAssignShiftRight) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprBinaryBitwiseAnd(n *ast.ExprBinaryBitwiseAnd) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryBitwiseOr(n *ast.ExprBinaryBitwiseOr) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryBitwiseXor(n *ast.ExprBinaryBitwiseXor) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryBooleanAnd(n *ast.ExprBinaryBooleanAnd) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryBooleanOr(n *ast.ExprBinaryBooleanOr) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryCoalesce(n *ast.ExprBinaryCoalesce) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryConcat(n *ast.ExprBinaryConcat) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryDiv(n *ast.ExprBinaryDiv) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryEqual(n *ast.ExprBinaryEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryGreater(n *ast.ExprBinaryGreater) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryGreaterOrEqual(n *ast.ExprBinaryGreaterOrEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryIdentical(n *ast.ExprBinaryIdentical) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryLogicalAnd(n *ast.ExprBinaryLogicalAnd) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryLogicalOr(n *ast.ExprBinaryLogicalOr) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryLogicalXor(n *ast.ExprBinaryLogicalXor) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryMinus(n *ast.ExprBinaryMinus) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryMod(n *ast.ExprBinaryMod) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryMul(n *ast.ExprBinaryMul) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryNotEqual(n *ast.ExprBinaryNotEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryNotIdentical(n *ast.ExprBinaryNotIdentical) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryPlus(n *ast.ExprBinaryPlus) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryPow(n *ast.ExprBinaryPow) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryShiftLeft(n *ast.ExprBinaryShiftLeft) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinaryShiftRight(n *ast.ExprBinaryShiftRight) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinarySmaller(n *ast.ExprBinarySmaller) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinarySmallerOrEqual(n *ast.ExprBinarySmallerOrEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprBinarySpaceship(n *ast.ExprBinarySpaceship) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *Traverser) ExprCastArray(n *ast.ExprCastArray) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprCastBool(n *ast.ExprCastBool) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprCastDouble(n *ast.ExprCastDouble) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprCastInt(n *ast.ExprCastInt) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprCastObject(n *ast.ExprCastObject) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprCastString(n *ast.ExprCastString) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprCastUnset(n *ast.ExprCastUnset) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ExprMatch(n *ast.ExprMatch) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
	t.TraverseNodes(n.Arms)
}

func (t *Traverser) ExprThrow(n *ast.ExprThrow) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *Traverser) ScalarDnumber(n *ast.ScalarDnumber) {
	// Do Nothing
}

func (t *Traverser) ScalarEncapsed(n *ast.ScalarEncapsed) {
	t.TraverseNodes(n.Parts)
}

func (t *Traverser) ScalarEncapsedStringPart(n *ast.ScalarEncapsedStringPart) {
	// Do Nothing
}

func (t *Traverser) ScalarEncapsedStringVar(n *ast.ScalarEncapsedStringVar) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Dim); replacedNode != nil {
		n.Dim = replacedNode
	}
}

func (t *Traverser) ScalarEncapsedStringBrackets(n *ast.ScalarEncapsedStringBrackets) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *Traverser) ScalarHeredoc(n *ast.ScalarHeredoc) {
	t.TraverseNodes(n.Parts)
}

func (t *Traverser) ScalarLnumber(n *ast.ScalarLnumber) {
	// Do Nothing
}

func (t *Traverser) ScalarMagicConstant(n *ast.ScalarMagicConstant) {
	// Do Nothing
}

func (t *Traverser) ScalarString(n *ast.ScalarString) {
	// Do Nothing
}

func (t *Traverser) NameName(n *ast.Name) {
	t.TraverseNodes(n.Parts)
}

func (t *Traverser) NameFullyQualified(n *ast.NameFullyQualified) {
	t.TraverseNodes(n.Parts)
}

func (t *Traverser) NameRelative(n *ast.NameRelative) {
	t.TraverseNodes(n.Parts)
}

func (t *Traverser) NameNamePart(n *ast.NamePart) {
	// Do Nothing
}

func validReplacement(v1 ast.Vertex, v2 ast.Vertex) bool {
	isV1Stmt := isStmt(v1)
	isV2Stmt := isStmt(v2)
	if isV1Stmt && !isV2Stmt {
		return false
	} else if !isV1Stmt && isV2Stmt {
		return false
	}

	return true
}

func isStmt(v ast.Vertex) bool {
	switch v.(type) {
	case *ast.StmtBreak, *ast.StmtCase, *ast.StmtCatch, *ast.StmtEnum, *ast.EnumCase, *ast.StmtClass, *ast.StmtClassConstList, *ast.StmtClassMethod, *ast.StmtConstList, *ast.StmtConstant, *ast.StmtContinue, *ast.StmtDeclare, *ast.StmtDefault, *ast.StmtDo, *ast.StmtEcho, *ast.StmtElse, *ast.StmtElseIf, *ast.StmtExpression, *ast.StmtFinally, *ast.StmtFor, *ast.StmtForeach, *ast.StmtFunction, *ast.StmtGlobal, *ast.StmtGoto, *ast.StmtHaltCompiler, *ast.StmtIf, *ast.StmtInlineHtml, *ast.StmtInterface, *ast.StmtLabel, *ast.StmtNamespace, *ast.StmtNop, *ast.StmtProperty, *ast.StmtPropertyList, *ast.StmtReturn, *ast.StmtStatic, *ast.StmtStaticVar, *ast.StmtStmtList, *ast.StmtSwitch, *ast.StmtThrow, *ast.StmtTrait, *ast.StmtTraitUse, *ast.StmtTraitUseAlias, *ast.StmtTraitUsePrecedence, *ast.StmtTry, *ast.StmtUnset, *ast.StmtUse, *ast.StmtGroupUseList, *ast.StmtUseList, *ast.StmtWhile:
		return true
	}
	return false
}
