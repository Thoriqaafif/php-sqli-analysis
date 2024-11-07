// Package to resolve Magic Constant in PHP, such as:
// __CLASS__, __TRAIT__, __NAMESPACE__
// __FUNCTION__, __METHOD__, __LINE__
package mcresolver

import (
	"fmt"
	"os"
	"strings"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taint-analysis/cfg/asttraverser"
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/position"
)

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

func (mcr *MagicConstResolver) EnterNode(n ast.Vertex) (ast.Vertex, asttraverser.ReturnedNodeType) {
	switch n := n.(type) {
	case *ast.StmtClass:
		// Append class name to class stack
		className := n.Name.(*ast.Identifier)
		classNameStr := string(className.Value)
		mcr.classStack = append(mcr.classStack, classNameStr)

		// Append parent class name to parent stack
		parentClassName := n.Extends.(*ast.Identifier)
		parentNameStr := string(parentClassName.Value)
		mcr.parentStack = append(mcr.parentStack, parentNameStr)

	case *ast.StmtTrait:
		// Append trait name to class stack
		traitName := n.Name.(*ast.Identifier)
		traitNameStr := string(traitName.Value)
		mcr.classStack = append(mcr.classStack, traitNameStr)

	case *ast.StmtClassMethod:
		// Append method name to method and function stack
		functionName := n.Name.(*ast.Identifier)
		functionNameStr := string(functionName.Value)
		mcr.functionStack = append(mcr.functionStack, functionNameStr)

		currClassName := mcr.classStack[len(mcr.classStack)-1]
		methodNameStr := fmt.Sprintf("%s::%s", currClassName, functionNameStr)
		mcr.methodStack = append(mcr.methodStack, methodNameStr)

	case *ast.Name:
		// Get name string
		nodeName := concatNameParts(n.Parts)
		if nodeName == "self" {
			// Error 'self' constant not recognized
			if len(mcr.classStack) == 0 {
				fmt.Println("Cannot use 'self' when no class scope is active")
				os.Exit(1)
			}

			// convert self to the current class name
			currClassName := mcr.classStack[len(mcr.classStack)-1]

			return &ast.NameFullyQualified{
				Position: n.Position,
				Parts:    createNameParts(currClassName, n.Position),
			}, asttraverser.ReturnReplacedNode
		} else if nodeName == "parent" {
			// Error 'parent' constant not recognized
			if len(mcr.parentStack) == 0 {
				fmt.Println("Cannot use 'parent' when current class scope has no parent")
				os.Exit(1)
			}

			// convert 'parent' to the current parent name
			parentName := mcr.parentStack[len(mcr.parentStack)-1]

			return &ast.NameFullyQualified{
				Position: n.Position,
				Parts:    createNameParts(parentName, n.Position),
			}, asttraverser.ReturnReplacedNode
		}

	case *ast.ScalarMagicConstant:
		magicConstStr := string(n.Value)

		if magicConstStr == "__CLASS__" {
			var currClassName string

			// If not in class scope, convert to empty string
			if len(mcr.classStack) == 0 {
				currClassName = ""
			} else {
				currClassName = mcr.classStack[len(mcr.classStack)-1]
			}

			return &ast.ScalarString{
				Position: n.Position,
				Value:    []byte(currClassName),
			}, asttraverser.ReturnReplacedNode
		} else if magicConstStr == "__TRAIT__" {

		}
	}

	return nil, asttraverser.ReturnReplacedNode
}

func (mcr *MagicConstResolver) LeaveNode(n ast.Vertex) (ast.Vertex, asttraverser.ReturnedNodeType) {
	// do nothing

	return nil, asttraverser.ReturnReplacedNode
}

func createNameParts(name string, pos *position.Position) []ast.Vertex {
	nameParts := make([]ast.Vertex, 0, 5)
	parts := strings.Split(name, "\\")

	for _, p := range parts {
		namePart := &ast.NamePart{
			Position: pos,
			Value:    []byte(p),
		}
		nameParts = append(nameParts, namePart)
	}

	return nameParts
}

func concatNameParts(parts ...[]ast.Vertex) string {
	str := ""

	for _, p := range parts {
		for _, n := range p {
			if str == "" {
				str = string(n.(*ast.NamePart).Value)
			} else {
				str = str + "\\" + string(n.(*ast.NamePart).Value)
			}
		}
	}

	return str
}
