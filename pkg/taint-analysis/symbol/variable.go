package symbol

import "github.com/VKCOM/php-parser/pkg/ast"

type VarID int
type ValType int

const (
	EXPR ValType = iota
	LITR
	PHI
	UNDEF
)

type Var struct {
	Id      VarID
	Type    ValType
	Val     any
	ASTNode ast.Vertex
}

func NewVar(id VarID, tp ValType, val any, node ast.Vertex) *Var {
	return &Var{
		Id:      id,
		Type:    tp,
		Val:     val,
		ASTNode: node,
	}
}
