// Package to resolve Magic Constant in PHP, such as:
// __CLASS__, __TRAIT__, __NAMESPACE__
// __FUNCTION__, __METHOD__, __LINE__
package mcresolver

import "github.com/VKCOM/php-parser/pkg/ast"

type MagicConstResolver struct {
	classStack    []string
	parentStack   []string
	functionStack []string
	methodStack   []string
}

func NewMagicConstResolver() *MagicConstResolver {
	return &MagicConstResolver{
		classStack:    make([]string, 0),
		parentStack:   make([]string, 0),
		functionStack: make([]string, 0),
		methodStack:   make([]string, 0),
	}
}

func (nsr *MagicConstResolver) EnterNode(n ast.Vertex) {
	switch n := n.(type) {
	}
}

func (nsr *MagicConstResolver) LeaveNode(n ast.Vertex) {
	// do nothing
}
