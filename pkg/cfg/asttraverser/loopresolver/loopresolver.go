// Package to resolve break and continue
// into Goto for easier analysis
package loopresolver

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg/asttraverser"
	"github.com/VKCOM/php-parser/pkg/ast"
)

type LoopResolver struct {
	continueStack []ast.StmtLabel
	breakStack    []ast.StmtLabel

	labelCtr int
}

func NewLoopResolver() *LoopResolver {
	return &LoopResolver{
		continueStack: make([]ast.StmtLabel, 0),
		breakStack:    make([]ast.StmtLabel, 0),
		labelCtr:      0,
	}
}

func (lr *LoopResolver) EnterNode(n ast.Vertex) (ast.Vertex, asttraverser.ReturnedNodeType) {
	switch n := n.(type) {
	case *ast.StmtBreak:
		// replace break statement with goto
		// based on label in breakStack
		return lr.resolveBreakStack(n), asttraverser.ReturnReplacedNode
	case *ast.StmtContinue:
		// replace continue statement with goto
		// based on label in continueStack
		return lr.resolveContinueStack(n), asttraverser.ReturnReplacedNode
	case *ast.StmtSwitch:
		// break and continue have similar behaviour
		// so just create 1 label for both statements
		label := lr.makeLabel()
		lr.breakStack = append(lr.breakStack, label)
		lr.continueStack = append(lr.continueStack, label)
	case *ast.StmtFor, *ast.StmtForeach, *ast.StmtDo, *ast.StmtWhile:
		// create two label for break and continue statements
		lr.breakStack = append(lr.breakStack, lr.makeLabel())
		lr.continueStack = append(lr.continueStack, lr.makeLabel())
	}

	return nil, asttraverser.ReturnReplacedNode
}

func (lr *LoopResolver) LeaveNode(n ast.Vertex) (ast.Vertex, asttraverser.ReturnedNodeType) {
	switch n := n.(type) {
	case *ast.StmtSwitch:
		// place Goto for continue and break in node
		// after switch statement
		popLabelStack(&lr.continueStack)
		return popLabelStack(&lr.breakStack), asttraverser.ReturnInsertedNode
	case *ast.StmtFor:
		n.Stmt.(*ast.StmtStmtList).Stmts = append(n.Stmt.(*ast.StmtStmtList).Stmts, popLabelStack(&lr.continueStack))
		return popLabelStack(&lr.breakStack), asttraverser.ReturnInsertedNode
	}

	return nil, asttraverser.ReturnReplacedNode
}

func (lr *LoopResolver) resolveBreakStack(n *ast.StmtBreak) *ast.StmtGoto {
	// break statement without parameter
	if n.Expr == nil {
		label := popLabelStack(&lr.breakStack)
		return &ast.StmtGoto{
			Label: label,
		}
	}

	if nExpr, ok := n.Expr.(*ast.ScalarLnumber); ok {
		paramNum, err := strconv.Atoi(string(nExpr.Value))
		if err != nil || paramNum <= 0 {
			fmt.Println("'break' operator accepts only positive integers")
			os.Exit(1)
		}

		// too much break
		if paramNum > len(lr.breakStack) {
			fmt.Printf("Cannot 'break' %d level\n", paramNum)
			os.Exit(1)
		}

		// get appropriate break location
		loc := lr.breakStack[len(lr.breakStack)-paramNum]
		return &ast.StmtGoto{
			Label: &loc,
		}
	} else {
		fmt.Println("'break' operator accepts only positive integers")
		os.Exit(1)
	}

	return nil
}

func (lr *LoopResolver) resolveContinueStack(n *ast.StmtContinue) *ast.StmtGoto {
	// continue statement without parameter
	if n.Expr == nil {
		label := popLabelStack(&lr.continueStack)
		return &ast.StmtGoto{
			Label: label,
		}
	}

	if nExpr, ok := n.Expr.(*ast.ScalarLnumber); ok {
		paramNum, err := strconv.Atoi(string(nExpr.Value))
		if err != nil || paramNum <= 0 {
			fmt.Println("'continue' operator accepts only positive integers")
			os.Exit(1)
		}

		// too much continue
		if paramNum > len(lr.continueStack) {
			fmt.Printf("Cannot 'continue' %d level\n", paramNum)
			os.Exit(1)
		}

		// get appropriate continue location
		loc := lr.continueStack[len(lr.continueStack)-paramNum]
		return &ast.StmtGoto{
			Label: &loc,
		}
	} else {
		fmt.Println("'continue' operator accepts only positive integers")
		os.Exit(1)
	}

	return nil
}

func (lr *LoopResolver) makeLabel() ast.StmtLabel {
	labelName := fmt.Sprintf("compiled_label_%d_%d", rand.Int(), lr.labelCtr)
	lr.labelCtr += 1

	return ast.StmtLabel{
		Name: &ast.Identifier{
			Value: []byte(labelName),
		},
	}
}

func popLabelStack(stack *[]ast.StmtLabel) *ast.StmtLabel {
	top := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]

	return &top
}
