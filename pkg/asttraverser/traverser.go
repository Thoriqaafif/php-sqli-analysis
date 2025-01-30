package asttraverser

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/VKCOM/php-parser/pkg/ast"
)

type ReturnedNodeType int

const (
	ReturnReplacedNode ReturnedNodeType = iota
	ReturnInsertedNode
)

type InsertedNode struct {
	Idx  int
	Node ast.Vertex
}

type NodeTraverser interface {
	EnterNode(n ast.Vertex) (ast.Vertex, ReturnedNodeType)
	LeaveNode(n ast.Vertex) (ast.Vertex, ReturnedNodeType)
}

type AstTraverser struct {
	NodeTravs []NodeTraverser
}

func NewTraverser() *AstTraverser {
	return &AstTraverser{
		NodeTravs: make([]NodeTraverser, 0),
	}
}

func (t *AstTraverser) AddNodeTraverser(nts ...NodeTraverser) {
	t.NodeTravs = append(t.NodeTravs, nts...)
}

func (t *AstTraverser) Traverse(n ast.Vertex) ast.Vertex {
	if n != nil {
		// Enter Node
		for _, nt := range t.NodeTravs {
			replacedNode, nType := nt.EnterNode(n)
			if replacedNode != nil {
				if nType == ReturnReplacedNode &&
					validReplacement(n, replacedNode) {
					return replacedNode
				} else {
					log.Fatalf("Error in Traverse enter node: Invalid node replacement '%v' - '%v'", reflect.TypeOf(n), reflect.TypeOf(replacedNode))
				}
			}
		}

		n.Accept(t)

		// Leave Node
		for _, nt := range t.NodeTravs {
			replacedNode, nType := nt.LeaveNode(n)
			if replacedNode != nil {
				if nType == ReturnReplacedNode &&
					validReplacement(n, replacedNode) {
					return replacedNode
				} else {
					// log.Fatalf("Error in Traverse leave node: Invalid node replacement '%v' - '%v'", reflect.TypeOf(n), reflect.TypeOf(replacedNode))
				}
			}
		}
	}

	return nil
}

func (t *AstTraverser) TraverseNodes(ns []ast.Vertex) []ast.Vertex {
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
						log.Fatalf("Error in Traverse enter list nodes: Invalid node replacement '%v' - '%v'", reflect.TypeOf(n), reflect.TypeOf(returnedNode))
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
						log.Fatalf("Error in Traverse leave list nodes: Invalid node replacement '%v' - '%v'", reflect.TypeOf(n), reflect.TypeOf(returnedNode))
					}
				} else if nType == ReturnInsertedNode {
					insertedNodes = append(insertedNodes,
						InsertedNode{Idx: i, Node: returnedNode})
				} else {
					fmt.Println("Error while traversing array of nodes")
					os.Exit(1)
				}
			}
		}
	}

	// inserting nodes
	for i := len(insertedNodes) - 1; i >= 0; i-- {
		idx := insertedNodes[i].Idx
		node := insertedNodes[i].Node

		// if there is other node in the right, append it
		if idx < len(ns)-1 {
			left := ns[:idx+1]
			right := append([]ast.Vertex{node}, ns[idx+1:]...)
			ns = append(left, right...)
		} else {
			ns = append(ns, node)
		}
	}

	return ns
}

func (t *AstTraverser) Root(n *ast.Root) {
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *AstTraverser) Nullable(n *ast.Nullable) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) Parameter(n *ast.Parameter) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	n.Modifiers = t.TraverseNodes(n.Modifiers)

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

func (t *AstTraverser) Identifier(n *ast.Identifier) {
	// do nothing
}

func (t *AstTraverser) Argument(n *ast.Argument) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) MatchArm(n *ast.MatchArm) {
	n.Exprs = t.TraverseNodes(n.Exprs)

	if replacedNode := t.Traverse(n.ReturnExpr); replacedNode != nil {
		n.ReturnExpr = replacedNode
	}
}

func (t *AstTraverser) Union(n *ast.Union) {
	n.Types = t.TraverseNodes(n.Types)
}

func (t *AstTraverser) Intersection(n *ast.Intersection) {
	n.Types = t.TraverseNodes(n.Types)
}

func (t *AstTraverser) Attribute(n *ast.Attribute) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *AstTraverser) AttributeGroup(n *ast.AttributeGroup) {
	n.Attrs = t.TraverseNodes(n.Attrs)
}

func (t *AstTraverser) StmtBreak(n *ast.StmtBreak) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) StmtCase(n *ast.StmtCase) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}

	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *AstTraverser) StmtCatch(n *ast.StmtCatch) {
	n.Types = t.TraverseNodes(n.Types)

	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}

	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *AstTraverser) StmtEnum(n *ast.StmtEnum) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}

	n.Implements = t.TraverseNodes(n.Implements)
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *AstTraverser) EnumCase(n *ast.EnumCase) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) StmtClass(n *ast.StmtClass) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Modifiers = t.TraverseNodes(n.Modifiers)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
	n.Implements = t.TraverseNodes(n.Implements)

	if replacedNode := t.Traverse(n.Extends); replacedNode != nil {
		n.Extends = replacedNode
	}

	n.Implements = t.TraverseNodes(n.Implements)
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *AstTraverser) StmtClassConstList(n *ast.StmtClassConstList) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Modifiers = t.TraverseNodes(n.Modifiers)
	n.Consts = t.TraverseNodes(n.Consts)
}

func (t *AstTraverser) StmtClassMethod(n *ast.StmtClassMethod) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Modifiers = t.TraverseNodes(n.Modifiers)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Params = t.TraverseNodes(n.Params)

	if replacedNode := t.Traverse(n.ReturnType); replacedNode != nil {
		n.ReturnType = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *AstTraverser) StmtConstList(n *ast.StmtConstList) {
	n.Consts = t.TraverseNodes(n.Consts)
}

func (t *AstTraverser) StmtConstant(n *ast.StmtConstant) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) StmtContinue(n *ast.StmtContinue) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) StmtDeclare(n *ast.StmtDeclare) {
	n.Consts = t.TraverseNodes(n.Consts)

	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *AstTraverser) StmtDefault(n *ast.StmtDefault) {
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *AstTraverser) StmtDo(n *ast.StmtDo) {
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
}

func (t *AstTraverser) StmtEcho(n *ast.StmtEcho) {
	n.Exprs = t.TraverseNodes(n.Exprs)
}

func (t *AstTraverser) StmtElse(n *ast.StmtElse) {
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *AstTraverser) StmtElseIf(n *ast.StmtElseIf) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *AstTraverser) StmtExpression(n *ast.StmtExpression) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) StmtFinally(n *ast.StmtFinally) {
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *AstTraverser) StmtFor(n *ast.StmtFor) {
	n.Init = t.TraverseNodes(n.Init)
	n.Cond = t.TraverseNodes(n.Cond)
	n.Loop = t.TraverseNodes(n.Loop)

	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *AstTraverser) StmtForeach(n *ast.StmtForeach) {
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

func (t *AstTraverser) StmtFunction(n *ast.StmtFunction) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Params = t.TraverseNodes(n.Params)

	if replacedNode := t.Traverse(n.ReturnType); replacedNode != nil {
		n.ReturnType = replacedNode
	}

	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *AstTraverser) StmtGlobal(n *ast.StmtGlobal) {
	n.Vars = t.TraverseNodes(n.Vars)
}

func (t *AstTraverser) StmtGoto(n *ast.StmtGoto) {
	if replacedNode := t.Traverse(n.Label); replacedNode != nil {
		n.Label = replacedNode
	}
}

func (t *AstTraverser) StmtHaltCompiler(n *ast.StmtHaltCompiler) {
	// Do Nothing
}

func (t *AstTraverser) StmtIf(n *ast.StmtIf) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}

	n.ElseIf = t.TraverseNodes(n.ElseIf)

	if replacedNode := t.Traverse(n.Else); replacedNode != nil {
		n.Else = replacedNode
	}
}

func (t *AstTraverser) StmtInlineHtml(n *ast.StmtInlineHtml) {
	// Do Nothing
}

func (t *AstTraverser) StmtInterface(n *ast.StmtInterface) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Extends = t.TraverseNodes(n.Extends)
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *AstTraverser) StmtLabel(n *ast.StmtLabel) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
}

func (t *AstTraverser) StmtNamespace(n *ast.StmtNamespace) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *AstTraverser) StmtNop(n *ast.StmtNop) {
	// Do Nothing
}

func (t *AstTraverser) StmtProperty(n *ast.StmtProperty) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) StmtPropertyList(n *ast.StmtPropertyList) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Modifiers = t.TraverseNodes(n.Modifiers)

	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}

	n.Props = t.TraverseNodes(n.Props)
}

func (t *AstTraverser) StmtReturn(n *ast.StmtReturn) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) StmtStatic(n *ast.StmtStatic) {
	n.Vars = t.TraverseNodes(n.Vars)
}

func (t *AstTraverser) StmtStaticVar(n *ast.StmtStaticVar) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) StmtStmtList(n *ast.StmtStmtList) {
	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *AstTraverser) StmtSwitch(n *ast.StmtSwitch) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}

	n.Cases = t.TraverseNodes(n.Cases)
}

func (t *AstTraverser) StmtThrow(n *ast.StmtThrow) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) StmtTrait(n *ast.StmtTrait) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)

	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}

	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *AstTraverser) StmtTraitUse(n *ast.StmtTraitUse) {
	n.Traits = t.TraverseNodes(n.Traits)
	n.Adaptations = t.TraverseNodes(n.Adaptations)
}

func (t *AstTraverser) StmtTraitUseAlias(n *ast.StmtTraitUseAlias) {
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

func (t *AstTraverser) StmtTraitUsePrecedence(n *ast.StmtTraitUsePrecedence) {
	if replacedNode := t.Traverse(n.Trait); replacedNode != nil {
		n.Trait = replacedNode
	}
	if replacedNode := t.Traverse(n.Method); replacedNode != nil {
		n.Method = replacedNode
	}
	n.Insteadof = t.TraverseNodes(n.Insteadof)
}

func (t *AstTraverser) StmtTry(n *ast.StmtTry) {
	n.Stmts = t.TraverseNodes(n.Stmts)
	n.Catches = t.TraverseNodes(n.Catches)

	if replacedNode := t.Traverse(n.Finally); replacedNode != nil {
		n.Finally = replacedNode
	}
}

func (t *AstTraverser) StmtUnset(n *ast.StmtUnset) {
	n.Vars = t.TraverseNodes(n.Vars)
}

func (t *AstTraverser) StmtUse(n *ast.StmtUseList) {
	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}
	n.Uses = t.TraverseNodes(n.Uses)
}

func (t *AstTraverser) StmtGroupUse(n *ast.StmtGroupUseList) {
	if replacedNode := t.Traverse(n.Type); replacedNode != nil {
		n.Type = replacedNode
	}
	if replacedNode := t.Traverse(n.Prefix); replacedNode != nil {
		n.Prefix = replacedNode
	}

	n.Uses = t.TraverseNodes(n.Uses)
}

func (t *AstTraverser) StmtUseDeclaration(n *ast.StmtUse) {
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

func (t *AstTraverser) StmtWhile(n *ast.StmtWhile) {
	if replacedNode := t.Traverse(n.Cond); replacedNode != nil {
		n.Cond = replacedNode
	}
	if replacedNode := t.Traverse(n.Stmt); replacedNode != nil {
		n.Stmt = replacedNode
	}
}

func (t *AstTraverser) ExprArray(n *ast.ExprArray) {
	n.Items = t.TraverseNodes(n.Items)
}

func (t *AstTraverser) ExprArrayDimFetch(n *ast.ExprArrayDimFetch) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Dim); replacedNode != nil {
		n.Dim = replacedNode
	}
}

func (t *AstTraverser) ExprArrayItem(n *ast.ExprArrayItem) {
	if replacedNode := t.Traverse(n.Key); replacedNode != nil {
		n.Key = replacedNode
	}
	if replacedNode := t.Traverse(n.Val); replacedNode != nil {
		n.Val = replacedNode
	}
}

func (t *AstTraverser) ExprArrowFunction(n *ast.ExprArrowFunction) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Params = t.TraverseNodes(n.Params)

	if replacedNode := t.Traverse(n.ReturnType); replacedNode != nil {
		n.ReturnType = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprBitwiseNot(n *ast.ExprBitwiseNot) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprBooleanNot(n *ast.ExprBooleanNot) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprBrackets(n *ast.ExprBrackets) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprClassConstFetch(n *ast.ExprClassConstFetch) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
	if replacedNode := t.Traverse(n.Const); replacedNode != nil {
		n.Const = replacedNode
	}
}

func (t *AstTraverser) ExprClone(n *ast.ExprClone) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprClosure(n *ast.ExprClosure) {
	n.AttrGroups = t.TraverseNodes(n.AttrGroups)
	n.Params = t.TraverseNodes(n.Params)
	n.Uses = t.TraverseNodes(n.Uses)

	if replacedNode := t.Traverse(n.ReturnType); replacedNode != nil {
		n.ReturnType = replacedNode
	}

	n.Stmts = t.TraverseNodes(n.Stmts)
}

func (t *AstTraverser) ExprClosureUse(n *ast.ExprClosureUse) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *AstTraverser) ExprConstFetch(n *ast.ExprConstFetch) {
	if replacedNode := t.Traverse(n.Const); replacedNode != nil {
		n.Const = replacedNode
	}
}

func (t *AstTraverser) ExprEmpty(n *ast.ExprEmpty) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprErrorSuppress(n *ast.ExprErrorSuppress) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprEval(n *ast.ExprEval) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprExit(n *ast.ExprExit) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprFunctionCall(n *ast.ExprFunctionCall) {
	if replacedNode := t.Traverse(n.Function); replacedNode != nil {
		n.Function = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *AstTraverser) ExprInclude(n *ast.ExprInclude) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprIncludeOnce(n *ast.ExprIncludeOnce) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprInstanceOf(n *ast.ExprInstanceOf) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
}

func (t *AstTraverser) ExprIsset(n *ast.ExprIsset) {
	n.Vars = t.TraverseNodes(n.Vars)
}

func (t *AstTraverser) ExprList(n *ast.ExprList) {
	n.Items = t.TraverseNodes(n.Items)
}

func (t *AstTraverser) ExprMethodCall(n *ast.ExprMethodCall) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Method); replacedNode != nil {
		n.Method = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *AstTraverser) ExprNullsafeMethodCall(n *ast.ExprNullsafeMethodCall) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Method); replacedNode != nil {
		n.Method = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *AstTraverser) ExprNew(n *ast.ExprNew) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *AstTraverser) ExprPostDec(n *ast.ExprPostDec) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *AstTraverser) ExprPostInc(n *ast.ExprPostInc) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *AstTraverser) ExprPreDec(n *ast.ExprPreDec) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *AstTraverser) ExprPreInc(n *ast.ExprPreInc) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *AstTraverser) ExprPrint(n *ast.ExprPrint) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprPropertyFetch(n *ast.ExprPropertyFetch) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Prop); replacedNode != nil {
		n.Prop = replacedNode
	}
}

func (t *AstTraverser) ExprNullsafePropertyFetch(n *ast.ExprNullsafePropertyFetch) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Prop); replacedNode != nil {
		n.Prop = replacedNode
	}
}

func (t *AstTraverser) ExprRequire(n *ast.ExprRequire) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprRequireOnce(n *ast.ExprRequireOnce) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprShellExec(n *ast.ExprShellExec) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *AstTraverser) ExprStaticCall(n *ast.ExprStaticCall) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
	if replacedNode := t.Traverse(n.Call); replacedNode != nil {
		n.Call = replacedNode
	}

	n.Args = t.TraverseNodes(n.Args)
}

func (t *AstTraverser) ExprStaticPropertyFetch(n *ast.ExprStaticPropertyFetch) {
	if replacedNode := t.Traverse(n.Class); replacedNode != nil {
		n.Class = replacedNode
	}
	if replacedNode := t.Traverse(n.Prop); replacedNode != nil {
		n.Prop = replacedNode
	}
}

func (t *AstTraverser) ExprTernary(n *ast.ExprTernary) {
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

func (t *AstTraverser) ExprUnaryMinus(n *ast.ExprUnaryMinus) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprUnaryPlus(n *ast.ExprUnaryPlus) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprVariable(n *ast.ExprVariable) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
}

func (t *AstTraverser) ExprYield(n *ast.ExprYield) {
	if replacedNode := t.Traverse(n.Key); replacedNode != nil {
		n.Key = replacedNode
	}
	if replacedNode := t.Traverse(n.Val); replacedNode != nil {
		n.Val = replacedNode
	}
}

func (t *AstTraverser) ExprYieldFrom(n *ast.ExprYieldFrom) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssign(n *ast.ExprAssign) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssignReference(n *ast.ExprAssignReference) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssignBitwiseAnd(n *ast.ExprAssignBitwiseAnd) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssignBitwiseOr(n *ast.ExprAssignBitwiseOr) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssignBitwiseXor(n *ast.ExprAssignBitwiseXor) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssignCoalesce(n *ast.ExprAssignCoalesce) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssignConcat(n *ast.ExprAssignConcat) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssignDiv(n *ast.ExprAssignDiv) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssignMinus(n *ast.ExprAssignMinus) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssignMod(n *ast.ExprAssignMod) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssignMul(n *ast.ExprAssignMul) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssignPlus(n *ast.ExprAssignPlus) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssignPow(n *ast.ExprAssignPow) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssignShiftLeft(n *ast.ExprAssignShiftLeft) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprAssignShiftRight(n *ast.ExprAssignShiftRight) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryBitwiseAnd(n *ast.ExprBinaryBitwiseAnd) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryBitwiseOr(n *ast.ExprBinaryBitwiseOr) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryBitwiseXor(n *ast.ExprBinaryBitwiseXor) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryBooleanAnd(n *ast.ExprBinaryBooleanAnd) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryBooleanOr(n *ast.ExprBinaryBooleanOr) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryCoalesce(n *ast.ExprBinaryCoalesce) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryConcat(n *ast.ExprBinaryConcat) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryDiv(n *ast.ExprBinaryDiv) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryEqual(n *ast.ExprBinaryEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryGreater(n *ast.ExprBinaryGreater) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryGreaterOrEqual(n *ast.ExprBinaryGreaterOrEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryIdentical(n *ast.ExprBinaryIdentical) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryLogicalAnd(n *ast.ExprBinaryLogicalAnd) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryLogicalOr(n *ast.ExprBinaryLogicalOr) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryLogicalXor(n *ast.ExprBinaryLogicalXor) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryMinus(n *ast.ExprBinaryMinus) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryMod(n *ast.ExprBinaryMod) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryMul(n *ast.ExprBinaryMul) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryNotEqual(n *ast.ExprBinaryNotEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryNotIdentical(n *ast.ExprBinaryNotIdentical) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryPlus(n *ast.ExprBinaryPlus) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryPow(n *ast.ExprBinaryPow) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryShiftLeft(n *ast.ExprBinaryShiftLeft) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinaryShiftRight(n *ast.ExprBinaryShiftRight) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinarySmaller(n *ast.ExprBinarySmaller) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinarySmallerOrEqual(n *ast.ExprBinarySmallerOrEqual) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprBinarySpaceship(n *ast.ExprBinarySpaceship) {
	if replacedNode := t.Traverse(n.Left); replacedNode != nil {
		n.Left = replacedNode
	}
	if replacedNode := t.Traverse(n.Right); replacedNode != nil {
		n.Right = replacedNode
	}
}

func (t *AstTraverser) ExprCastArray(n *ast.ExprCastArray) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprCastBool(n *ast.ExprCastBool) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprCastDouble(n *ast.ExprCastDouble) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprCastInt(n *ast.ExprCastInt) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprCastObject(n *ast.ExprCastObject) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprCastString(n *ast.ExprCastString) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprCastUnset(n *ast.ExprCastUnset) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ExprMatch(n *ast.ExprMatch) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
	n.Arms = t.TraverseNodes(n.Arms)
}

func (t *AstTraverser) ExprThrow(n *ast.ExprThrow) {
	if replacedNode := t.Traverse(n.Expr); replacedNode != nil {
		n.Expr = replacedNode
	}
}

func (t *AstTraverser) ScalarDnumber(n *ast.ScalarDnumber) {
	// Do Nothing
}

func (t *AstTraverser) ScalarEncapsed(n *ast.ScalarEncapsed) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *AstTraverser) ScalarEncapsedStringPart(n *ast.ScalarEncapsedStringPart) {
	// Do Nothing
}

func (t *AstTraverser) ScalarEncapsedStringVar(n *ast.ScalarEncapsedStringVar) {
	if replacedNode := t.Traverse(n.Name); replacedNode != nil {
		n.Name = replacedNode
	}
	if replacedNode := t.Traverse(n.Dim); replacedNode != nil {
		n.Dim = replacedNode
	}
}

func (t *AstTraverser) ScalarEncapsedStringBrackets(n *ast.ScalarEncapsedStringBrackets) {
	if replacedNode := t.Traverse(n.Var); replacedNode != nil {
		n.Var = replacedNode
	}
}

func (t *AstTraverser) ScalarHeredoc(n *ast.ScalarHeredoc) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *AstTraverser) ScalarLnumber(n *ast.ScalarLnumber) {
	// Do Nothing
}

func (t *AstTraverser) ScalarMagicConstant(n *ast.ScalarMagicConstant) {
	// Do Nothing
}

func (t *AstTraverser) ScalarString(n *ast.ScalarString) {
	// Do Nothing
}

func (t *AstTraverser) NameName(n *ast.Name) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *AstTraverser) NameFullyQualified(n *ast.NameFullyQualified) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *AstTraverser) NameRelative(n *ast.NameRelative) {
	n.Parts = t.TraverseNodes(n.Parts)
}

func (t *AstTraverser) NameNamePart(n *ast.NamePart) {
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
	case *ast.StmtBreak, *ast.StmtCase, *ast.StmtCatch, *ast.StmtEnum, *ast.EnumCase, *ast.StmtClass,
		*ast.StmtClassConstList, *ast.StmtClassMethod, *ast.StmtConstList, *ast.StmtConstant,
		*ast.StmtContinue, *ast.StmtDeclare, *ast.StmtDefault, *ast.StmtDo, *ast.StmtEcho, *ast.StmtElse,
		*ast.StmtElseIf, *ast.StmtExpression, *ast.StmtFinally, *ast.StmtFor, *ast.StmtForeach,
		*ast.StmtFunction, *ast.StmtGlobal, *ast.StmtGoto, *ast.StmtHaltCompiler, *ast.StmtIf,
		*ast.StmtInlineHtml, *ast.StmtInterface, *ast.StmtLabel, *ast.StmtNamespace, *ast.StmtNop,
		*ast.StmtProperty, *ast.StmtPropertyList, *ast.StmtReturn, *ast.StmtStatic, *ast.StmtStaticVar,
		*ast.StmtStmtList, *ast.StmtSwitch, *ast.StmtThrow, *ast.StmtTrait, *ast.StmtTraitUse,
		*ast.StmtTraitUseAlias, *ast.StmtTraitUsePrecedence, *ast.StmtTry, *ast.StmtUnset, *ast.StmtUse,
		*ast.StmtGroupUseList, *ast.StmtUseList, *ast.StmtWhile:
		return true
	}
	return false
}
