// Package to resolve name based on its
// namespace path
package nsresolver

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taint-analysis/cfg/asttraverser"
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/position"
)

// NamespaceResolver visitor
type NamespaceResolver struct {
	Namespace     *Namespace
	ResolvedNames map[ast.Vertex]string

	goDeep           bool
	anonClassCounter int
}

// NewNamespaceResolver NamespaceResolver type constructor
func NewNamespaceResolver() *NamespaceResolver {
	return &NamespaceResolver{
		Namespace:        NewNamespace(""),
		ResolvedNames:    map[ast.Vertex]string{},
		goDeep:           true,
		anonClassCounter: 0,
	}
}

func (nsr *NamespaceResolver) EnterNode(n ast.Vertex) (ast.Vertex, asttraverser.ReturnedNodeType) {
	switch n := n.(type) {
	case *ast.StmtNamespace:
		return nsr.StmtNamespace(n), asttraverser.ReturnReplacedNode
	case *ast.StmtUseList:
		return nsr.StmtUse(n), asttraverser.ReturnReplacedNode
	case *ast.StmtGroupUseList:
		return nsr.StmtGroupUse(n), asttraverser.ReturnReplacedNode
	case *ast.StmtClass:
		return nsr.StmtClass(n), asttraverser.ReturnReplacedNode
	case *ast.StmtInterface:
		return nsr.StmtInterface(n), asttraverser.ReturnReplacedNode
	case *ast.StmtTrait:
		return nsr.StmtTrait(n), asttraverser.ReturnReplacedNode
	case *ast.StmtFunction:
		return nsr.StmtFunction(n), asttraverser.ReturnReplacedNode
	case *ast.StmtClassMethod:
		return nsr.StmtClassMethod(n), asttraverser.ReturnReplacedNode
	case *ast.ExprClosure:
		return nsr.ExprClosure(n), asttraverser.ReturnReplacedNode
	case *ast.StmtPropertyList:
		return nsr.StmtPropertyList(n), asttraverser.ReturnReplacedNode
	case *ast.StmtConstList:
		return nsr.StmtConstList(n), asttraverser.ReturnReplacedNode
	case *ast.ExprStaticCall:
		return nsr.ExprStaticCall(n), asttraverser.ReturnReplacedNode
	case *ast.ExprStaticPropertyFetch:
		return nsr.ExprStaticPropertyFetch(n), asttraverser.ReturnReplacedNode
	case *ast.ExprClassConstFetch:
		return nsr.ExprClassConstFetch(n), asttraverser.ReturnReplacedNode
	case *ast.ExprNew:
		return nsr.ExprNew(n), asttraverser.ReturnReplacedNode
	case *ast.ExprInstanceOf:
		return nsr.ExprInstanceOf(n), asttraverser.ReturnReplacedNode
	case *ast.StmtCatch:
		return nsr.StmtCatch(n), asttraverser.ReturnReplacedNode
	case *ast.ExprFunctionCall:
		return nsr.ExprFunctionCall(n), asttraverser.ReturnReplacedNode
	case *ast.ExprConstFetch:
		return nsr.ExprConstFetch(n), asttraverser.ReturnReplacedNode
	case *ast.StmtTraitUse:
		return nsr.StmtTraitUse(n), asttraverser.ReturnReplacedNode
	}

	return nil, asttraverser.ReturnReplacedNode
}

func (nsr *NamespaceResolver) LeaveNode(n ast.Vertex) (ast.Vertex, asttraverser.ReturnedNodeType) {
	// do nothing
	return nil, asttraverser.ReturnReplacedNode
}

func (nsr *NamespaceResolver) StmtNamespace(n *ast.StmtNamespace) ast.Vertex {
	if n.Name == nil {
		nsr.Namespace = NewNamespace("")
	} else {
		NSParts := n.Name.(*ast.Name).Parts
		nsr.Namespace = NewNamespace(concatNameParts(NSParts))
	}

	return nil
}

func (nsr *NamespaceResolver) StmtUse(n *ast.StmtUseList) ast.Vertex {
	useType := ""
	if n.Type != nil {
		useType = string(n.Type.(*ast.Identifier).Value)
	}

	for _, nn := range n.Uses {
		nsr.AddAlias(useType, nn, nil)
	}

	nsr.goDeep = false

	return nil
}

func (nsr *NamespaceResolver) StmtGroupUse(n *ast.StmtGroupUseList) ast.Vertex {
	useType := ""
	if n.Type != nil {
		useType = string(n.Type.(*ast.Identifier).Value)
	}

	for _, nn := range n.Uses {
		nsr.AddAlias(useType, nn, n.Prefix.(*ast.Name).Parts)
	}

	nsr.goDeep = false

	return nil
}

func (nsr *NamespaceResolver) StmtClass(n *ast.StmtClass) ast.Vertex {
	fmt.Println("StmtClass")

	if n.Extends != nil {
		nsr.ResolveName(n.Extends, "")
	}

	if n.Implements != nil {
		for _, interfaceName := range n.Implements {
			nsr.ResolveName(interfaceName, "")
		}
	}

	if n.Name != nil {
		nsr.AddNamespacedName(n.Name.(*ast.Identifier), string(n.Name.(*ast.Identifier).Value))
	} else {
		// anonymous class
		nsr.AddNamespacedName(n.Name.(*ast.Identifier), fmt.Sprintf("{anonymousClass}#%d", nsr.anonClassCounter))
		nsr.anonClassCounter += 1
	}

	return nil
}

func (nsr *NamespaceResolver) StmtInterface(n *ast.StmtInterface) ast.Vertex {
	if n.Extends != nil {
		for _, interfaceName := range n.Extends {
			nsr.ResolveName(interfaceName, "")
		}
	}

	nsr.AddNamespacedName(n.Name.(*ast.Identifier), string(n.Name.(*ast.Identifier).Value))

	return nil
}

func (nsr *NamespaceResolver) StmtTrait(n *ast.StmtTrait) ast.Vertex {
	nsr.AddNamespacedName(n.Name.(*ast.Identifier), string(n.Name.(*ast.Identifier).Value))

	return nil
}

func (nsr *NamespaceResolver) StmtFunction(n *ast.StmtFunction) ast.Vertex {
	nsr.AddNamespacedName(n.Name.(*ast.Identifier), string(n.Name.(*ast.Identifier).Value))

	for _, parameter := range n.Params {
		nsr.ResolveType(parameter.(*ast.Parameter).Type)
	}

	if n.ReturnType != nil {
		nsr.ResolveType(n.ReturnType)
	}

	return nil
}

func (nsr *NamespaceResolver) StmtClassMethod(n *ast.StmtClassMethod) ast.Vertex {
	for _, parameter := range n.Params {
		nsr.ResolveType(parameter.(*ast.Parameter).Type)
	}

	if n.ReturnType != nil {
		nsr.ResolveType(n.ReturnType)
	}

	return nil
}

func (nsr *NamespaceResolver) ExprClosure(n *ast.ExprClosure) ast.Vertex {
	for _, parameter := range n.Params {
		nsr.ResolveType(parameter.(*ast.Parameter).Type)
	}

	if n.ReturnType != nil {
		nsr.ResolveType(n.ReturnType)
	}

	return nil
}

func (nsr *NamespaceResolver) StmtPropertyList(n *ast.StmtPropertyList) ast.Vertex {
	if n.Type != nil {
		nsr.ResolveType(n.Type)
	}

	return nil
}

func (nsr *NamespaceResolver) StmtConstList(n *ast.StmtConstList) ast.Vertex {
	for _, constant := range n.Consts {
		constant := constant.(*ast.StmtConstant)
		nsr.AddNamespacedName(constant.Name.(*ast.Identifier), string(constant.Name.(*ast.Identifier).Value))
	}

	return nil
}

func (nsr *NamespaceResolver) ExprStaticCall(n *ast.ExprStaticCall) ast.Vertex {
	nsr.ResolveName(n.Class, "")

	return nil
}

func (nsr *NamespaceResolver) ExprStaticPropertyFetch(n *ast.ExprStaticPropertyFetch) ast.Vertex {
	nsr.ResolveName(n.Class, "")

	return nil
}

func (nsr *NamespaceResolver) ExprClassConstFetch(n *ast.ExprClassConstFetch) ast.Vertex {
	nsr.ResolveName(n.Class, "")

	return nil
}

func (nsr *NamespaceResolver) ExprNew(n *ast.ExprNew) ast.Vertex {
	fmt.Println("ExprNew")
	nsr.ResolveName(n.Class, "")

	return nil
}

func (nsr *NamespaceResolver) ExprInstanceOf(n *ast.ExprInstanceOf) ast.Vertex {
	nsr.ResolveName(n.Class, "")

	return nil
}

func (nsr *NamespaceResolver) StmtCatch(n *ast.StmtCatch) ast.Vertex {
	for _, t := range n.Types {
		nsr.ResolveName(t, "")
	}

	return nil
}

func (nsr *NamespaceResolver) ExprFunctionCall(n *ast.ExprFunctionCall) ast.Vertex {
	nsr.ResolveName(n.Function, "function")

	return nil
}

func (nsr *NamespaceResolver) ExprConstFetch(n *ast.ExprConstFetch) ast.Vertex {
	nsr.ResolveName(n.Const, "const")

	return nil
}

func (nsr *NamespaceResolver) StmtTraitUse(n *ast.StmtTraitUse) ast.Vertex {
	for _, t := range n.Traits {
		nsr.ResolveName(t, "")
	}

	for _, a := range n.Adaptations {
		switch aa := a.(type) {
		case *ast.StmtTraitUsePrecedence:
			refTrait := aa.Trait
			if refTrait != nil {
				nsr.ResolveName(refTrait, "")
			}
			for _, insteadOf := range aa.Insteadof {
				nsr.ResolveName(insteadOf, "")
			}

		case *ast.StmtTraitUseAlias:
			refTrait := aa.Trait
			if refTrait != nil {
				nsr.ResolveName(refTrait, "")
			}
		}
	}

	return nil
}

// AddAlias adds a new alias
func (nsr *NamespaceResolver) AddAlias(useType string, nn ast.Vertex, prefix []ast.Vertex) {
	switch use := nn.(type) {
	case *ast.StmtUse:
		if use.Type != nil {
			useType = string(use.Type.(*ast.Identifier).Value)
		}

		useNameParts := use.Use.(*ast.Name).Parts
		var alias string
		if use.Alias == nil {
			alias = string(useNameParts[len(useNameParts)-1].(*ast.NamePart).Value)
		} else {
			alias = string(use.Alias.(*ast.Identifier).Value)
		}

		nsr.Namespace.AddAlias(useType, concatNameParts(prefix, useNameParts), alias)
	}
}

// AddNamespacedName adds namespaced name by node
func (nsr *NamespaceResolver) AddNamespacedName(nn *ast.Identifier, nodeName string) {
	var resolvedName string
	if nsr.Namespace.Namespace == "" {
		nsr.ResolvedNames[nn] = nodeName
		resolvedName = nodeName
	} else {
		nsr.ResolvedNames[nn] = nsr.Namespace.Namespace + "\\" + nodeName
		fmt.Println(nsr.ResolvedNames[nn])
		resolvedName = nsr.Namespace.Namespace + "\\" + nodeName
	}

	nn.Value = []byte(resolvedName)
}

// ResolveName adds a resolved fully qualified name by node
func (nsr *NamespaceResolver) ResolveName(nameNode ast.Vertex, aliasType string) {
	fmt.Printf("ResolveName: %v\n", nameNode)
	resolved, err := nsr.Namespace.ResolveName(nameNode, aliasType)
	if err == nil {
		nsr.ResolvedNames[nameNode] = resolved
		fmt.Printf("ResolveName: %v, %v\n", resolved, aliasType)

		switch nameNode := nameNode.(type) {
		case *ast.Name:
			nameNode.Parts = createNameParts(resolved, nameNode.Position)
		case *ast.NameFullyQualified:
			nameNode.Parts = createNameParts(resolved, nameNode.Position)
		case *ast.NameRelative:
			nameNode.Parts = createNameParts(resolved, nameNode.Position)
		}
	}
}

// ResolveType adds a resolved fully qualified type name
func (nsr *NamespaceResolver) ResolveType(n ast.Vertex) {
	switch nn := n.(type) {
	case *ast.Nullable:
		nsr.ResolveType(nn.Expr)
	case *ast.Name:
		nsr.ResolveName(n, "")
	case *ast.NameRelative:
		nsr.ResolveName(n, "")
	case *ast.NameFullyQualified:
		nsr.ResolveName(n, "")
	}
}

// Namespace context
type Namespace struct {
	Namespace string
	Aliases   map[string]map[string]string
}

// NewNamespace constructor
func NewNamespace(NSName string) *Namespace {
	return &Namespace{
		Namespace: NSName,
		Aliases: map[string]map[string]string{
			"":         {},
			"const":    {},
			"function": {},
		},
	}
}

// AddAlias adds a new alias
func (ns *Namespace) AddAlias(aliasType string, aliasName string, alias string) {
	aliasType = strings.ToLower(aliasType)

	if aliasType == "const" {
		ns.Aliases[aliasType][alias] = aliasName
	} else {
		ns.Aliases[aliasType][strings.ToLower(alias)] = aliasName
	}
}

// ResolveName returns a resolved fully qualified name
func (ns *Namespace) ResolveName(nameNode ast.Vertex, aliasType string) (string, error) {
	switch n := nameNode.(type) {
	case *ast.NameFullyQualified:
		// Fully qualifid name is already resolved
		return concatNameParts(n.Parts), nil

	case *ast.NameRelative:
		if ns.Namespace == "" {
			return concatNameParts(n.Parts), nil
		}
		return ns.Namespace + "\\" + concatNameParts(n.Parts), nil

	case *ast.Name:
		if aliasType == "const" && len(n.Parts) == 1 {
			part := strings.ToLower(string(n.Parts[0].(*ast.NamePart).Value))
			if part == "true" || part == "false" || part == "null" {
				return part, nil
			}
		}

		if aliasType == "" && len(n.Parts) == 1 {
			part := strings.ToLower(string(n.Parts[0].(*ast.NamePart).Value))

			switch part {
			case "self":
				fallthrough
			case "static":
				fallthrough
			case "parent":
				fallthrough
			case "int":
				fallthrough
			case "float":
				fallthrough
			case "bool":
				fallthrough
			case "string":
				fallthrough
			case "void":
				fallthrough
			case "iterable":
				fallthrough
			case "object":
				return part, nil
			}
		}

		aliasName, err := ns.ResolveAlias(nameNode, aliasType)
		if err != nil {
			fmt.Printf("resolveName: %s\n", concatNameParts(n.Parts))
			// resolve as relative name if alias not found
			if ns.Namespace == "" {
				return concatNameParts(n.Parts), nil
			}
			return ns.Namespace + "\\" + concatNameParts(n.Parts), nil
		}

		if len(n.Parts) > 1 {
			// if name qualified, replace first part by alias
			return aliasName + "\\" + concatNameParts(n.Parts[1:]), nil
		}

		return aliasName, nil
	}

	return "", errors.New("must be instance of name.Names")
}

// ResolveAlias returns alias or error if not found
func (ns *Namespace) ResolveAlias(nameNode ast.Vertex, aliasType string) (string, error) {
	aliasType = strings.ToLower(aliasType)
	nameParts := nameNode.(*ast.Name).Parts

	firstPartStr := string(nameParts[0].(*ast.NamePart).Value)

	if len(nameParts) > 1 { // resolve aliases for qualified names, always against class alias type
		firstPartStr = strings.ToLower(firstPartStr)
		aliasType = ""
	} else {
		if aliasType != "const" { // constants are case-sensitive
			firstPartStr = strings.ToLower(firstPartStr)
		}
	}

	aliasName, ok := ns.Aliases[aliasType][firstPartStr]
	if !ok {
		return "", errors.New("not found")
	}

	return aliasName, nil
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
