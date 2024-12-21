// Package to resolve Magic Constant in PHP, such as:
// __CLASS__, __TRAIT__, __NAMESPACE__
// __FUNCTION__, __METHOD__, __LINE__
package mcresolver

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg/asttraverser"
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/position"
)

type MagicConstResolver struct {
	classStack    []string
	parentStack   []string
	functionStack []string
	methodStack   []string
	currNamespace string
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
		if n.Extends != nil {
			parentClassName := n.Extends.(*ast.Identifier)
			parentNameStr := string(parentClassName.Value)
			mcr.parentStack = append(mcr.parentStack, parentNameStr)
		} else {
			mcr.parentStack = append(mcr.parentStack, "")
		}

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

	case *ast.StmtNamespace:
		// set current namespace context
		nameSpaceStr := concatNameParts(n.Name.(*ast.Name).Parts)
		mcr.currNamespace = nameSpaceStr

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
			var currTraitName string

			// If not in trait scope, convert to empty string
			if len(mcr.classStack) == 0 {
				currTraitName = ""
			} else {
				currTraitName = mcr.classStack[len(mcr.classStack)-1]
			}

			return &ast.ScalarString{
				Position: n.Position,
				Value:    []byte(currTraitName),
			}, asttraverser.ReturnReplacedNode
		} else if magicConstStr == "__NAMESPACE__" {
			return &ast.ScalarString{
				Position: n.Position,
				Value:    []byte(mcr.currNamespace),
			}, asttraverser.ReturnReplacedNode
		} else if magicConstStr == "__FUNCTION__" {
			var functionName string

			// If not in function scope, convert to empty string
			if len(mcr.classStack) == 0 {
				functionName = ""
			} else {
				functionName = mcr.functionStack[len(mcr.functionStack)-1]
			}

			return &ast.ScalarString{
				Position: n.Position,
				Value:    []byte(functionName),
			}, asttraverser.ReturnReplacedNode
		} else if magicConstStr == "__METHOD__" {
			var methodName string

			// If not in method scope, convert to empty string
			if len(mcr.methodStack) == 0 {
				methodName = ""
			} else {
				methodName = mcr.methodStack[len(mcr.methodStack)-1]
			}

			return &ast.ScalarString{
				Position: n.Position,
				Value:    []byte(methodName),
			}, asttraverser.ReturnReplacedNode
		} else if magicConstStr == "__LINE__" {
			fmt.Println("__LINE__")
			fmt.Println(1)
			fmt.Println(2)
			fmt.Println(3)
			return &ast.ScalarLnumber{
				Position: n.Position,
				Value:    []byte(strconv.Itoa(n.Position.StartLine)),
			}, asttraverser.ReturnReplacedNode
		} else {
			fmt.Printf("Invalid Magic Constant: %s", magicConstStr)
		}
	}

	return nil, asttraverser.ReturnReplacedNode
}

func (mcr *MagicConstResolver) LeaveNode(n ast.Vertex) (ast.Vertex, asttraverser.ReturnedNodeType) {
	switch n := n.(type) {
	case *ast.StmtClass:
		popStringStack(&mcr.classStack)
		popStringStack(&mcr.parentStack)
	case *ast.StmtTrait:
		popStringStack(&mcr.classStack)
	case *ast.StmtFunction:
		popStringStack(&mcr.functionStack)
	case *ast.StmtClassMethod:
		popStringStack(&mcr.methodStack)
	case *ast.StmtNamespace:
		if len(n.Stmts) > 0 {
			mcr.currNamespace = ""
		}
	}

	return nil, asttraverser.ReturnReplacedNode
}

func popStringStack(st *[]string) string {
	top := (*st)[len(*st)-1]
	*st = (*st)[:len(*st)-1]

	return top
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
