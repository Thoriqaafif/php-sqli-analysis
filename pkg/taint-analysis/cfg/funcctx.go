package cfg

type FuncCtx struct {
	Labels          []Block
	Scope           map[BlockID]map[any]Operand // TODO: Change type
	incompletePhis  map[BlockID]map[any]Phi     // TODO: Change type
	Complete        bool                        // TODO: understand the function
	UnresolvedGotos map[any]any                 // TODO: Change type
}

func NewFuncCtx() FuncCtx {
	return FuncCtx{
		Scope:          make(map[BlockID]map[any]Operand),
		incompletePhis: make(map[BlockID]map[any]Phi),
		Complete:       false,
	}
}

// TODO: check name type
func (ctx *FuncCtx) SetValueInScope(blockId BlockID, name any, value Operand) {
	ctx.Scope[blockId][name] = value
}

// TODO: check name type
func (ctx *FuncCtx) isLocalVar(blockId BlockID, name any) bool {
	v, ok := ctx.Scope[blockId]
	if !ok {
		return false
	}

	if _, ok := v[name]; ok {
		return true
	}

	return false
}

// TODO: change phi type
func (ctx *FuncCtx) addIncompletePhi(blockId BlockID, name any, phi Phi) {
	ctx.incompletePhis[blockId][name] = phi
}
