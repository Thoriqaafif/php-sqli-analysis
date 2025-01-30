package cfg

type BlockId int

type Block struct {
	Instructions []Op
	Predecessors []*Block
	Phi          map[*OpPhi]struct{}
	Dead         bool

	// helper attribute
	Id              BlockId
	ContaintTainted bool
	IsConditional   bool
	Conds           []Operand // used in path generator
}

func NewBlock(id BlockId) *Block {
	return &Block{
		Instructions:    make([]Op, 0),
		Predecessors:    make([]*Block, 0),
		Phi:             make(map[*OpPhi]struct{}),
		Dead:            false,
		Id:              id,
		ContaintTainted: false,
		IsConditional:   false,
	}
}

func NewConditionalBlock(id BlockId) *Block {
	return &Block{
		Instructions:    make([]Op, 0),
		Predecessors:    make([]*Block, 0),
		Phi:             make(map[*OpPhi]struct{}),
		Dead:            false,
		Id:              id,
		ContaintTainted: false,
		IsConditional:   false,
	}
}

func (b *Block) AddPhi(phi *OpPhi) {
	b.Phi[phi] = struct{}{}
}

func (b *Block) RemovePhi(phi *OpPhi) {
	delete(b.Phi, phi)
}

func (b *Block) AddInstructions(op Op) {
	b.Instructions = append(b.Instructions, op)
}

func (b *Block) AddPredecessor(block *Block) {
	for _, pred := range b.Predecessors {
		if block == pred {
			return
		}
	}
	b.Predecessors = append(b.Predecessors, block)
}

func (b *Block) RemovePredecessor(block *Block) {
	idx := -1
	for i, pred := range b.Predecessors {
		if block == pred {
			idx = i
		}
	}
	if idx != -1 {
		b.Predecessors = append(b.Predecessors[:idx], b.Predecessors[idx+1:]...)
	}
}

func (b *Block) InPredecessors(block *Block) bool {
	for _, pred := range b.Predecessors {
		if pred == block {
			return true
		}
	}
	return false
}

func (b *Block) GetPhi() []*OpPhi {
	res := make([]*OpPhi, 0, len(b.Phi))
	for phi := range b.Phi {
		res = append(res, phi)
	}
	return res
}

func (b *Block) SetCondition(conds []Operand) {
	for _, cond := range conds {
		cond.AddCondUsage(b)
	}
	b.Conds = make([]Operand, len(conds))
	copy(b.Conds, conds)
}
