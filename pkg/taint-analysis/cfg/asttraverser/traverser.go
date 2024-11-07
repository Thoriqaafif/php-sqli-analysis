package asttraverser

import (
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

func (t *Traverser) Traverse(n ast.Vertex) {
	if n != nil {
		// Enter Node
		for _, nt := range t.NodeTravs {
			val, _ := nt.EnterNode(n)
			if val != nil {
				n = val
			}
		}

		n.Accept(t)

		// Leave Node
		for _, nt := range t.NodeTravs {
			val, _ := nt.LeaveNode(n)
			if val != nil {
				n = val
			}
		}
	}
}

func (t *Traverser) TraverseNodes(ns []ast.Vertex) {
	var insertedNodes []InsertedNode = make([]InsertedNode, 0)

	for i, n := range ns {
		// Enter Node
		for _, nt := range t.NodeTravs {
			val, _ := nt.EnterNode(n)
			if val != nil {
				n = val
			}
		}

		n.Accept(t)

		// Leave Node
		for _, nt := range t.NodeTravs {
			val, nType := nt.LeaveNode(n)
			if val != nil {
				if nType == ReturnReplacedNode {
					n = val
				} else if nType == ReturnInsertedNode {
					insertedNodes = append(insertedNodes, InsertedNode{Idx: i, Node: val})
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

	t.Traverse(n.Expr)

}

func (t *Traverser) Parameter(n *ast.Parameter) {
	t.TraverseNodes(n.AttrGroups)

	t.TraverseNodes(n.Modifiers)

	t.Traverse(n.Type)
	t.Traverse(n.Var)
	t.Traverse(n.DefaultValue)
}

func (t *Traverser) Identifier(n *ast.Identifier) {
	// do nothing
}

func (t *Traverser) Argument(n *ast.Argument) {
	t.Traverse(n.Name)
	t.Traverse(n.Expr)
}

func (t *Traverser) MatchArm(n *ast.MatchArm) {
	t.TraverseNodes(n.Exprs)

	t.Traverse(n.ReturnExpr)
}

func (t *Traverser) Union(n *ast.Union) {
	t.TraverseNodes(n.Types)
}

func (t *Traverser) Intersection(n *ast.Intersection) {
	t.TraverseNodes(n.Types)
}

func (t *Traverser) Attribute(n *ast.Attribute) {
	t.Traverse(n.Name)

	t.TraverseNodes(n.Args)
}

func (t *Traverser) AttributeGroup(n *ast.AttributeGroup) {
	t.TraverseNodes(n.Attrs)
}

func (t *Traverser) StmtBreak(n *ast.StmtBreak) {
	t.Traverse(n.Expr)
}

func (t *Traverser) StmtCase(n *ast.StmtCase) {
	t.Traverse(n.Cond)

	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtCatch(n *ast.StmtCatch) {
	t.TraverseNodes(n.Types)

	t.Traverse(n.Var)

	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtEnum(n *ast.StmtEnum) {
	t.TraverseNodes(n.AttrGroups)

	t.Traverse(n.Name)
	t.Traverse(n.Type)

	t.TraverseNodes(n.Implements)
	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) EnumCase(n *ast.EnumCase) {
	t.TraverseNodes(n.AttrGroups)

	t.Traverse(n.Name)
	t.Traverse(n.Expr)
}

func (t *Traverser) StmtClass(n *ast.StmtClass) {
	t.TraverseNodes(n.AttrGroups)
	t.TraverseNodes(n.Modifiers)

	t.Traverse(n.Name)

	t.TraverseNodes(n.Args)
	t.TraverseNodes(n.Implements)

	t.Traverse(n.Extends)

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

	t.Traverse(n.Name)

	t.TraverseNodes(n.Params)

	t.Traverse(n.ReturnType)
	t.Traverse(n.Stmt)
}

func (t *Traverser) StmtConstList(n *ast.StmtConstList) {
	t.TraverseNodes(n.Consts)
}

func (t *Traverser) StmtConstant(n *ast.StmtConstant) {
	t.Traverse(n.Name)
	t.Traverse(n.Expr)
}

func (t *Traverser) StmtContinue(n *ast.StmtContinue) {
	t.Traverse(n.Expr)
}

func (t *Traverser) StmtDeclare(n *ast.StmtDeclare) {
	t.TraverseNodes(n.Consts)

	t.Traverse(n.Stmt)
}

func (t *Traverser) StmtDefault(n *ast.StmtDefault) {
	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtDo(n *ast.StmtDo) {
	t.Traverse(n.Stmt)
	t.Traverse(n.Cond)
}

func (t *Traverser) StmtEcho(n *ast.StmtEcho) {
	t.TraverseNodes(n.Exprs)
}

func (t *Traverser) StmtElse(n *ast.StmtElse) {
	t.Traverse(n.Stmt)
}

func (t *Traverser) StmtElseIf(n *ast.StmtElseIf) {
	t.Traverse(n.Cond)
	t.Traverse(n.Stmt)
}

func (t *Traverser) StmtExpression(n *ast.StmtExpression) {
	t.Traverse(n.Expr)
}

func (t *Traverser) StmtFinally(n *ast.StmtFinally) {
	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtFor(n *ast.StmtFor) {
	t.TraverseNodes(n.Init)
	t.TraverseNodes(n.Cond)
	t.TraverseNodes(n.Loop)

	t.Traverse(n.Stmt)
}

func (t *Traverser) StmtForeach(n *ast.StmtForeach) {
	t.Traverse(n.Expr)
	t.Traverse(n.Key)
	t.Traverse(n.Var)
	t.Traverse(n.Stmt)
}

func (t *Traverser) StmtFunction(n *ast.StmtFunction) {
	t.TraverseNodes(n.AttrGroups)

	t.Traverse(n.Name)

	t.TraverseNodes(n.Params)

	t.Traverse(n.ReturnType)

	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtGlobal(n *ast.StmtGlobal) {
	t.TraverseNodes(n.Vars)
}

func (t *Traverser) StmtGoto(n *ast.StmtGoto) {
	t.Traverse(n.Label)
}

func (t *Traverser) StmtHaltCompiler(n *ast.StmtHaltCompiler) {
	// Do Nothing
}

func (t *Traverser) StmtIf(n *ast.StmtIf) {
	t.Traverse(n.Cond)
	t.Traverse(n.Stmt)

	t.TraverseNodes(n.ElseIf)

	t.Traverse(n.Else)
}

func (t *Traverser) StmtInlineHtml(n *ast.StmtInlineHtml) {
	// Do Nothing
}

func (t *Traverser) StmtInterface(n *ast.StmtInterface) {
	t.TraverseNodes(n.AttrGroups)

	t.Traverse(n.Name)

	t.TraverseNodes(n.Extends)
	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtLabel(n *ast.StmtLabel) {
	t.Traverse(n.Name)
}

func (t *Traverser) StmtNamespace(n *ast.StmtNamespace) {
	t.Traverse(n.Name)
	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtNop(n *ast.StmtNop) {
	// Do Nothing
}

func (t *Traverser) StmtProperty(n *ast.StmtProperty) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) StmtPropertyList(n *ast.StmtPropertyList) {
	t.TraverseNodes(n.AttrGroups)
	t.TraverseNodes(n.Modifiers)

	t.Traverse(n.Type)

	t.TraverseNodes(n.Props)
}

func (t *Traverser) StmtReturn(n *ast.StmtReturn) {
	t.Traverse(n.Expr)
}

func (t *Traverser) StmtStatic(n *ast.StmtStatic) {
	t.TraverseNodes(n.Vars)
}

func (t *Traverser) StmtStaticVar(n *ast.StmtStaticVar) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) StmtStmtList(n *ast.StmtStmtList) {
	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtSwitch(n *ast.StmtSwitch) {
	t.Traverse(n.Cond)

	t.TraverseNodes(n.Cases)
}

func (t *Traverser) StmtThrow(n *ast.StmtThrow) {
	t.Traverse(n.Expr)
}

func (t *Traverser) StmtTrait(n *ast.StmtTrait) {
	t.TraverseNodes(n.AttrGroups)

	t.Traverse(n.Name)

	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) StmtTraitUse(n *ast.StmtTraitUse) {
	t.TraverseNodes(n.Traits)
	t.TraverseNodes(n.Adaptations)
}

func (t *Traverser) StmtTraitUseAlias(n *ast.StmtTraitUseAlias) {
	t.Traverse(n.Trait)
	t.Traverse(n.Method)
	t.Traverse(n.Modifier)
	t.Traverse(n.Alias)
}

func (t *Traverser) StmtTraitUsePrecedence(n *ast.StmtTraitUsePrecedence) {
	t.Traverse(n.Trait)
	t.Traverse(n.Method)
	t.TraverseNodes(n.Insteadof)
}

func (t *Traverser) StmtTry(n *ast.StmtTry) {
	t.TraverseNodes(n.Stmts)
	t.TraverseNodes(n.Catches)

	t.Traverse(n.Finally)
}

func (t *Traverser) StmtUnset(n *ast.StmtUnset) {
	t.TraverseNodes(n.Vars)
}

func (t *Traverser) StmtUse(n *ast.StmtUseList) {
	t.Traverse(n.Type)
	t.TraverseNodes(n.Uses)
}

func (t *Traverser) StmtGroupUse(n *ast.StmtGroupUseList) {
	t.Traverse(n.Type)
	t.Traverse(n.Prefix)

	t.TraverseNodes(n.Uses)
}

func (t *Traverser) StmtUseDeclaration(n *ast.StmtUse) {
	t.Traverse(n.Type)
	t.Traverse(n.Use)
	t.Traverse(n.Alias)
}

func (t *Traverser) StmtWhile(n *ast.StmtWhile) {
	t.Traverse(n.Cond)
	t.Traverse(n.Stmt)
}

func (t *Traverser) ExprArray(n *ast.ExprArray) {
	t.TraverseNodes(n.Items)
}

func (t *Traverser) ExprArrayDimFetch(n *ast.ExprArrayDimFetch) {
	t.Traverse(n.Var)
	t.Traverse(n.Dim)
}

func (t *Traverser) ExprArrayItem(n *ast.ExprArrayItem) {
	t.Traverse(n.Key)
	t.Traverse(n.Val)
}

func (t *Traverser) ExprArrowFunction(n *ast.ExprArrowFunction) {
	t.TraverseNodes(n.AttrGroups)
	t.TraverseNodes(n.Params)

	t.Traverse(n.ReturnType)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprBitwiseNot(n *ast.ExprBitwiseNot) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprBooleanNot(n *ast.ExprBooleanNot) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprBrackets(n *ast.ExprBrackets) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprClassConstFetch(n *ast.ExprClassConstFetch) {
	t.Traverse(n.Class)
	t.Traverse(n.Const)
}

func (t *Traverser) ExprClone(n *ast.ExprClone) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprClosure(n *ast.ExprClosure) {
	t.TraverseNodes(n.AttrGroups)
	t.TraverseNodes(n.Params)
	t.TraverseNodes(n.Uses)

	t.Traverse(n.ReturnType)

	t.TraverseNodes(n.Stmts)
}

func (t *Traverser) ExprClosureUse(n *ast.ExprClosureUse) {
	t.Traverse(n.Var)
}

func (t *Traverser) ExprConstFetch(n *ast.ExprConstFetch) {
	t.Traverse(n.Const)
}

func (t *Traverser) ExprEmpty(n *ast.ExprEmpty) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprErrorSuppress(n *ast.ExprErrorSuppress) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprEval(n *ast.ExprEval) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprExit(n *ast.ExprExit) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprFunctionCall(n *ast.ExprFunctionCall) {
	t.Traverse(n.Function)

	t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprInclude(n *ast.ExprInclude) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprIncludeOnce(n *ast.ExprIncludeOnce) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprInstanceOf(n *ast.ExprInstanceOf) {
	t.Traverse(n.Expr)
	t.Traverse(n.Class)
}

func (t *Traverser) ExprIsset(n *ast.ExprIsset) {
	t.TraverseNodes(n.Vars)
}

func (t *Traverser) ExprList(n *ast.ExprList) {
	t.TraverseNodes(n.Items)
}

func (t *Traverser) ExprMethodCall(n *ast.ExprMethodCall) {
	t.Traverse(n.Var)
	t.Traverse(n.Method)

	t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprNullsafeMethodCall(n *ast.ExprNullsafeMethodCall) {
	t.Traverse(n.Var)
	t.Traverse(n.Method)

	t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprNew(n *ast.ExprNew) {
	t.Traverse(n.Class)

	t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprPostDec(n *ast.ExprPostDec) {
	t.Traverse(n.Var)
}

func (t *Traverser) ExprPostInc(n *ast.ExprPostInc) {
	t.Traverse(n.Var)
}

func (t *Traverser) ExprPreDec(n *ast.ExprPreDec) {
	t.Traverse(n.Var)
}

func (t *Traverser) ExprPreInc(n *ast.ExprPreInc) {
	t.Traverse(n.Var)
}

func (t *Traverser) ExprPrint(n *ast.ExprPrint) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprPropertyFetch(n *ast.ExprPropertyFetch) {
	t.Traverse(n.Var)
	t.Traverse(n.Prop)
}

func (t *Traverser) ExprNullsafePropertyFetch(n *ast.ExprNullsafePropertyFetch) {
	t.Traverse(n.Var)
	t.Traverse(n.Prop)
}

func (t *Traverser) ExprRequire(n *ast.ExprRequire) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprRequireOnce(n *ast.ExprRequireOnce) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprShellExec(n *ast.ExprShellExec) {
	t.TraverseNodes(n.Parts)
}

func (t *Traverser) ExprStaticCall(n *ast.ExprStaticCall) {
	t.Traverse(n.Class)
	t.Traverse(n.Call)

	t.TraverseNodes(n.Args)
}

func (t *Traverser) ExprStaticPropertyFetch(n *ast.ExprStaticPropertyFetch) {
	t.Traverse(n.Class)
	t.Traverse(n.Prop)
}

func (t *Traverser) ExprTernary(n *ast.ExprTernary) {
	t.Traverse(n.Cond)
	t.Traverse(n.IfTrue)
	t.Traverse(n.IfFalse)
}

func (t *Traverser) ExprUnaryMinus(n *ast.ExprUnaryMinus) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprUnaryPlus(n *ast.ExprUnaryPlus) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprVariable(n *ast.ExprVariable) {
	t.Traverse(n.Name)
}

func (t *Traverser) ExprYield(n *ast.ExprYield) {
	t.Traverse(n.Key)
	t.Traverse(n.Val)
}

func (t *Traverser) ExprYieldFrom(n *ast.ExprYieldFrom) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssign(n *ast.ExprAssign) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssignReference(n *ast.ExprAssignReference) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssignBitwiseAnd(n *ast.ExprAssignBitwiseAnd) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssignBitwiseOr(n *ast.ExprAssignBitwiseOr) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssignBitwiseXor(n *ast.ExprAssignBitwiseXor) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssignCoalesce(n *ast.ExprAssignCoalesce) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssignConcat(n *ast.ExprAssignConcat) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssignDiv(n *ast.ExprAssignDiv) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssignMinus(n *ast.ExprAssignMinus) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssignMod(n *ast.ExprAssignMod) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssignMul(n *ast.ExprAssignMul) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssignPlus(n *ast.ExprAssignPlus) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssignPow(n *ast.ExprAssignPow) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssignShiftLeft(n *ast.ExprAssignShiftLeft) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprAssignShiftRight(n *ast.ExprAssignShiftRight) {
	t.Traverse(n.Var)
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprBinaryBitwiseAnd(n *ast.ExprBinaryBitwiseAnd) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryBitwiseOr(n *ast.ExprBinaryBitwiseOr) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryBitwiseXor(n *ast.ExprBinaryBitwiseXor) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryBooleanAnd(n *ast.ExprBinaryBooleanAnd) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryBooleanOr(n *ast.ExprBinaryBooleanOr) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryCoalesce(n *ast.ExprBinaryCoalesce) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryConcat(n *ast.ExprBinaryConcat) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryDiv(n *ast.ExprBinaryDiv) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryEqual(n *ast.ExprBinaryEqual) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryGreater(n *ast.ExprBinaryGreater) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryGreaterOrEqual(n *ast.ExprBinaryGreaterOrEqual) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryIdentical(n *ast.ExprBinaryIdentical) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryLogicalAnd(n *ast.ExprBinaryLogicalAnd) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryLogicalOr(n *ast.ExprBinaryLogicalOr) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryLogicalXor(n *ast.ExprBinaryLogicalXor) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryMinus(n *ast.ExprBinaryMinus) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryMod(n *ast.ExprBinaryMod) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryMul(n *ast.ExprBinaryMul) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryNotEqual(n *ast.ExprBinaryNotEqual) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryNotIdentical(n *ast.ExprBinaryNotIdentical) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryPlus(n *ast.ExprBinaryPlus) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryPow(n *ast.ExprBinaryPow) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryShiftLeft(n *ast.ExprBinaryShiftLeft) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinaryShiftRight(n *ast.ExprBinaryShiftRight) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinarySmaller(n *ast.ExprBinarySmaller) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinarySmallerOrEqual(n *ast.ExprBinarySmallerOrEqual) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprBinarySpaceship(n *ast.ExprBinarySpaceship) {
	t.Traverse(n.Left)
	t.Traverse(n.Right)
}

func (t *Traverser) ExprCastArray(n *ast.ExprCastArray) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprCastBool(n *ast.ExprCastBool) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprCastDouble(n *ast.ExprCastDouble) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprCastInt(n *ast.ExprCastInt) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprCastObject(n *ast.ExprCastObject) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprCastString(n *ast.ExprCastString) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprCastUnset(n *ast.ExprCastUnset) {
	t.Traverse(n.Expr)
}

func (t *Traverser) ExprMatch(n *ast.ExprMatch) {
	t.Traverse(n.Expr)
	t.TraverseNodes(n.Arms)
}

func (t *Traverser) ExprThrow(n *ast.ExprThrow) {
	t.Traverse(n.Expr)
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
	t.Traverse(n.Name)
	t.Traverse(n.Dim)
}

func (t *Traverser) ScalarEncapsedStringBrackets(n *ast.ScalarEncapsedStringBrackets) {
	t.Traverse(n.Var)
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
