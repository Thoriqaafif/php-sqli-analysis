package cfg

type FuncFlag int
type FuncReturnType int

const (
	FLAG_PUBLIC FuncFlag = iota
	FLAG_PROTECTED
	FLAG_PRIVATE
	FLAG_STATIC
	FLAG_ABSTRACT
	FLAG_FINAL
	FLAG_RETURNS_REF
	FLAG_CLOSURE
)

const (
	LITERAL FuncReturnType = iota
	MIXED
	NULLABLE
	REFERENCE
	UNION
	VOID
)

type Func struct {
	OpGeneral
	Name       string
	Flag       FuncFlag
	ReturnType FuncReturnType
	Class      *Literal
	Params     []Param
	Cfg        Block
}

func NewClassFunc(name string, flag FuncFlag, returnType FuncReturnType, class Literal, id BlockID) (Op, error) {
	newBlock := NewBlock(id, make([]Block, 0))

	f := &Func{
		Name:       name,
		Flag:       flag,
		ReturnType: returnType,
		Class:      &class,
		Params:     make([]Param, 0),
		Cfg:        *newBlock,
	}
	return f, nil
}

func NewFunc(name string, flag FuncFlag, returnType FuncReturnType, id BlockID) Op {
	newBlock := NewBlock(id, make([]Block, 0))

	return &Func{
		Name:       name,
		Flag:       flag,
		ReturnType: returnType,
		Class:      nil,
		Params:     make([]Param, 0),
		Cfg:        *newBlock,
	}
}

func (f *Func) getScopedName() string {
	if f.Class == nil {
		className := f.Class.Val.(string)
		return className + "::" + f.Name
	}
	return f.Name
}
