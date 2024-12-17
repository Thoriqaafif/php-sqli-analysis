package astutil

import (
	"fmt"
	"reflect"

	"github.com/VKCOM/php-parser/pkg/ast"
)

func GetNameString(nameNode ast.Vertex) (string, error) {
	switch name := nameNode.(type) {
	case *ast.Name:
		return ConcatNameParts(name.Parts), nil
	case *ast.NameFullyQualified:
		return ConcatNameParts(name.Parts), nil
	case *ast.NameRelative:
		return ConcatNameParts(name.Parts), nil
	case *ast.Identifier:
		return string(name.Value), nil
	}
	return "", fmt.Errorf("incompatible name type '%s'", reflect.TypeOf(nameNode))
}

func IsScalarNode(n ast.Vertex) bool {
	if n == nil {
		return false
	}
	switch n.(type) {
	case *ast.ScalarDnumber:
		return true
	case *ast.ScalarString:
		return true
	case *ast.ScalarLnumber:
		return true
	}

	return false
}

func ConcatNameParts(parts ...[]ast.Vertex) string {
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
