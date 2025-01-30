package cfg

type FuncCtx struct {
	Labels          map[string]*Block
	Scope           map[*Block]map[string]Operand // var name definition for each block
	incompletePhis  map[*Block]map[string]*OpPhi  // incomplete phis for each block
	Complete        bool                          // Flag represent if block not sealed
	UnresolvedGotos map[string][]*Block           // blocks for each unresolved gotos
	CurrConds       []Operand                     // current conditions
}

func NewFuncCtx() FuncCtx {
	return FuncCtx{
		Labels:          make(map[string]*Block),
		Scope:           make(map[*Block]map[string]Operand),
		incompletePhis:  make(map[*Block]map[string]*OpPhi),
		Complete:        false,
		UnresolvedGotos: make(map[string][]*Block),
		CurrConds:       make([]Operand, 0),
	}
}

func (ctx *FuncCtx) getLabel(name string) (*Block, bool) {
	a, ok := ctx.Labels[name]

	return a, ok
}

func (ctx *FuncCtx) SetValueInScope(block *Block, name string, value Operand) {
	if ctx.Scope[block] == nil {
		ctx.Scope[block] = make(map[string]Operand)
	}

	ctx.Scope[block][name] = value
}

// TODO: check name type
func (ctx *FuncCtx) isLocalVar(block *Block, name string) bool {
	v, ok := ctx.Scope[block]
	if !ok {
		return false
	}

	_, isNameSet := v[name]

	return isNameSet
}

func (ctx *FuncCtx) getLocalVar(block *Block, name string) (Operand, bool) {
	if ctx.isLocalVar(block, name) {
		return ctx.Scope[block][name], true
	}

	return nil, false
}

// TODO: change phi type
func (ctx *FuncCtx) addIncompletePhis(block *Block, name string, phi *OpPhi) {
	if ctx.incompletePhis[block] == nil {
		ctx.incompletePhis[block] = make(map[string]*OpPhi)
	}

	ctx.incompletePhis[block][name] = phi
}

// UnresolvedGotos map[string]*Block
func (ctx *FuncCtx) addUnresolvedGoto(name string, block *Block) {
	if ctx.UnresolvedGotos[name] == nil {
		ctx.UnresolvedGotos[name] = make([]*Block, 0)
	}

	ctx.UnresolvedGotos[name] = append(ctx.UnresolvedGotos[name], block)
}

func (ctx *FuncCtx) getUnresolvedGotos(name string) ([]*Block, bool) {
	a, ok := ctx.UnresolvedGotos[name]

	return a, ok
}

func (ctx *FuncCtx) resolveGoto(name string) {
	delete(ctx.UnresolvedGotos, name)
}

func (ctx *FuncCtx) PushCond(cond Operand) {
	ctx.CurrConds = append(ctx.CurrConds, cond)
}

func (ctx *FuncCtx) PopCond() {
	ctx.CurrConds = ctx.CurrConds[:len(ctx.CurrConds)-1]
}
