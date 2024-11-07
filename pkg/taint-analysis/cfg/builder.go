package cfg

import (
	"log"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
)

const (
	MODE_NONE  = 0
	MODE_READ  = 1
	MODE_WRITE = 2
)

type Script struct {
	Funcs []Func
	Main  Func
}

type Builder struct {
	AstTraverser  []any  // TODO: change type
	FileName      string //
	Ctx           FuncCtx
	CurrClass     any    // TODO: change type to operand.Literal
	CurrNamespace string // TODO: check type again
	CurrScript    Script
	AnonId        int // TODO: understand the function
	blocks        map[BlockID]Block
	currBlock     BlockID
	counter       int
}

// TODO: add ast visitor for name, loop, and magic string resolver
func BuildCFG(src []byte, fileName string, astTraverser []any) Script { // TODO: change astTraverser type
	b := Builder{
		AstTraverser: astTraverser,
		FileName:     fileName,
		counter:      0,
	}

	// Error handler
	var parserErrors []*errors.Error
	errorHandler := func(e *errors.Error) {
		parserErrors = append(parserErrors, e)
	}

	rootNode, err := parser.Parse(src, conf.Config{
		Version:          &version.Version{Major: 5, Minor: 6},
		ErrorHandlerFunc: errorHandler,
	})
	root, ok := rootNode.(*ast.Root)

	if err != nil || !ok {
		log.Fatal("Error:" + err.Error())
	}

	// TODO: Parse AST
	// TODO: Traverse AST using resolver

	// Create script instance
	scr := Script{
		Funcs: make([]Func, 0),
		Main:  NewFunc("{main}", FLAG_PUBLIC, VOID, BlockID(b.getId())),
	}
	b.parseFunc(scr.Main, scr.Main.Params, root.Stmts)
}

// add a new variable definition
func (b *Builder) writeVariable()

// TODO: Check name type
// read defined variable
func (b *Builder) readVariable(vr Operand) Operand {
	// TODO: Code preprocess
	switch v := vr.(type) {
	case *BoundVar:
		return v
	case *Var:
		if name, ok := v.Name.(*Literal); ok {
			block := b.blocks[b.currBlock]
			return b.readVariableName(name.Val, block.ID)
		} else {
			b.readVariable(v.Name)
			return v
		}
	case *Temporary:
		if v.Original != nil {
			return b.readVariable(v.Original)
		}
	}

	return vr
}

// TODO: Check name type
func (b *Builder) readVariableRecursive(name any, blockId BlockID) Operand {
	block := b.blocks[blockId]
	if b.Ctx.Complete {
		// 1 Predecessors, just read from it
		if len(block.Preds) == 1 && !block.Preds[0].Dead {
			return b.readVariableName(name, block.Preds[0].ID)
		}

		vr := NewTemporary(NewVar(NewLiteral(name)))
		phi := NewPhi(vr)
	}
}

func (b *Builder) readVariableName(name any, blockId BlockID) Operand {
	def := b.Ctx.Scope[blockId]

	if val, ok := def[name]; ok {
		return val
	}
	return b.readVariableRecursive(name, blockId)
}

func (b *Builder) parseFunc(fn Func, params []Param, stmts []ast.Vertex) {
	// create new func context
	prevContext := b.Ctx
	b.Ctx = NewFuncCtx()
}

func (b *Builder) getId() int {
	id := b.counter
	b.counter += 1
	return id
}
