package simplifier

import (
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg/traverser"
)

type Simplifier struct {
	Removed             map[*cfg.Block]struct{}
	RecursionProtection map[cfg.Op]struct{}
	TrivPhiCandidate    map[*cfg.OpPhi]*cfg.Block

	traverser.NullTraverser
}

func NewSimplifier() *Simplifier {
	return &Simplifier{}
}

func (t *Simplifier) EnterFunc(fn *cfg.OpFunc) {
	t.Removed = make(map[*cfg.Block]struct{})
	t.RecursionProtection = make(map[cfg.Op]struct{})
}

func (t *Simplifier) LeaveFunc(fn *cfg.OpFunc) {
	// remove trivial phi
	if fn.Cfg != nil {
		t.TrivPhiCandidate = make(map[*cfg.OpPhi]*cfg.Block)
		t.removeTrivialPhi(fn.Cfg)
	}
}

func (t *Simplifier) EnterOp(op cfg.Op, block *cfg.Block) {
	if InOpSet(t.RecursionProtection, op) {
		return
	}
	AddToOpSet(t.RecursionProtection, op)

	// optimize for jump
	for targetName, target := range cfg.GetSubBlocks(op) {
		if len(target.Instructions) <= 0 {
			continue
		}
		jmpOp, isJumpOp := target.Instructions[0].(*cfg.OpStmtJump)
		if !isJumpOp {
			continue
		}
		jmpTarget := jmpOp.Target
		if InBlockSet(t.Removed, target) {
			cfg.ChangeSubBlock(op, targetName, jmpTarget)
			jmpTarget.AddPredecessor(block)
		} else {
			// optimize child
			t.EnterOp(jmpOp, target)

			// prevent kill infinite tight loop
			if jmpOp.Target == target {
				continue
			}

			// for a phi block
			phis := target.GetPhi()
			foundPhis := make([]*cfg.OpPhi, 0)
			for _, phi := range phis {
				for subPhi := range jmpTarget.Phi {
					// subPhi use phi value
					if subPhi.HasOperand(phi.Result) {
						foundPhis = append(foundPhis, subPhi)
						break
					}
				}
			}
			// not all phi used by subblock
			if len(foundPhis) != len(target.Phi) {
				continue
			}
			// here, we can remove phi node and jmp
			for i := 0; i < len(target.Phi); i++ {
				phi := phis[i]
				foundPhi := foundPhis[i]
				foundPhi.RemoveOperand(phi.Result)
				for oper := range phi.Vars {
					foundPhi.AddOperand(oper)
				}
			}
			target.Phi = make(map[*cfg.OpPhi]struct{})
			AddToBlockSet(t.Removed, target)
			target.Dead = true

			// remove target from list of preds
			jmpTarget.RemovePredecessor(target)
			jmpTarget.AddPredecessor(block)

			cfg.ChangeSubBlock(op, targetName, jmpTarget)
		}
	}
	RemoveFromOpSet(t.RecursionProtection, op)
}

func (t *Simplifier) removeTrivialPhi(block *cfg.Block) {
	toReplace := make(map[*cfg.Block]struct{})
	replaced := make(map[*cfg.Block]struct{})
	AddToBlockSet(toReplace, block)
	for len(toReplace) > 0 {
		for currBlock := range toReplace {
			RemoveFromBlockSet(toReplace, currBlock)
			AddToBlockSet(replaced, currBlock)
			for phi := range currBlock.Phi {
				if t.tryRemoveTrivialPhi(phi, currBlock) {
					currBlock.RemovePhi(phi)
				}
			}
			for _, op := range currBlock.Instructions {
				for _, subBlock := range cfg.GetSubBlocks(op) {
					if !InBlockSet(replaced, subBlock) {
						AddToBlockSet(toReplace, subBlock)
					}
				}
			}
		}
	}
	for len(t.TrivPhiCandidate) > 0 {
		for phi, currBlock := range t.TrivPhiCandidate {
			delete(t.TrivPhiCandidate, phi)
			if t.tryRemoveTrivialPhi(phi, currBlock) {
				currBlock.RemovePhi(phi)
			}
		}
	}
}

func (t *Simplifier) tryRemoveTrivialPhi(phi *cfg.OpPhi, block *cfg.Block) bool {
	// phi variables more than 1, not trivial
	if len(phi.Vars) > 1 {
		return false
	}

	var vr cfg.Operand
	if len(phi.Vars) == 0 {
		// unitialized variable
		return true
	}

	vr = phi.GetVars()[0]
	// remove phi, change with its variable
	t.replaceVariables(phi.Result, vr, block)

	return true
}

// remove operand which become trivial from a phi
func (t *Simplifier) replaceVariables(from, to cfg.Operand, block *cfg.Block) {
	toReplace := make(map[*cfg.Block]struct{})
	replaced := make(map[*cfg.Block]struct{})
	AddToBlockSet(toReplace, block)
	for len(toReplace) > 0 {
		for block := range toReplace {
			RemoveFromBlockSet(toReplace, block)
			AddToBlockSet(replaced, block)
			for phi := range block.Phi {
				if phi.HasOperand(from) {
					// removing operand from phi, hence phi maybe become trivial
					t.TrivPhiCandidate[phi] = block
					phi.RemoveOperand(from)
					phi.AddOperand(to)
				}
			}
			for _, op := range block.Instructions {
				t.replaceOpVariable(from, to, op)
				for _, subBlock := range cfg.GetSubBlocks(op) {
					if !InBlockSet(replaced, subBlock) {
						AddToBlockSet(toReplace, subBlock)
					}
				}
				// propagate new value
				switch o := op.(type) {
				case *cfg.OpExprAssign:
					result := cfg.Operand(nil)
					switch r := o.Expr.(type) {
					case *cfg.OperBool, *cfg.OperNumber, *cfg.OperObject, *cfg.OperString, *cfg.OperSymbolic:
						result = o.Expr
					case *cfg.OperVariable:
						result = r.Value
					case *cfg.OperTemporary:
						if rv, ok := r.Original.(*cfg.OperVariable); ok {
							result = rv.Value
						}
					}

					if result != nil {
						o.Result = result
						// get left variable, then give the value
						switch l := o.Var.(type) {
						case *cfg.OperVariable:
							l.Value = o.Result
						case *cfg.OperTemporary:
							if lv, ok := l.Original.(*cfg.OperVariable); ok {
								lv.Value = o.Result
							}
						}
					}
				}
			}
		}
	}
}

func (t *Simplifier) replaceOpVariable(from, to cfg.Operand, op cfg.Op) {
	for vrName, vr := range op.GetOpVars() {
		if vr == from {
			// change previous operand which is trivial phi
			op.ChangeOpVar(vrName, to)
			from.RemoveUsage(op)
			if cfg.IsWriteVar(op, vrName) {
				to.AddWriteOp(op)
			} else {
				to.AddUsage(op)
			}
		}
	}
	for vrName, vrList := range op.GetOpListVars() {
		new := make([]cfg.Operand, len(vrList))
		for i, vr := range vrList {
			if vr == from {
				new[i] = to
				to.AddUsage(op)
				from.RemoveUsage(op)
			} else {
				new[i] = vr
			}
		}
		op.ChangeOpListVar(vrName, new)
	}
}

func AddToBlockSet(set map[*cfg.Block]struct{}, item *cfg.Block) {
	set[item] = struct{}{}
}

func RemoveFromBlockSet(set map[*cfg.Block]struct{}, item *cfg.Block) {
	delete(set, item)
}

func InBlockSet(set map[*cfg.Block]struct{}, item *cfg.Block) bool {
	if _, ok := set[item]; ok {
		return true
	}
	return false
}

func AddToOpSet(set map[cfg.Op]struct{}, item cfg.Op) {
	set[item] = struct{}{}
}

func RemoveFromOpSet(set map[cfg.Op]struct{}, item cfg.Op) {
	delete(set, item)
}

func InOpSet(set map[cfg.Op]struct{}, item cfg.Op) bool {
	if _, ok := set[item]; ok {
		return true
	}
	return false
}
