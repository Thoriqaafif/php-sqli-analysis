package cfg

type BlockId int

type Block struct {
	Instructions []Op                // TODO: change type to instruction node
	Preds        []*Block            // predecessor of a basic block
	Phi          map[*OpPhi]struct{} // TODO: change type to phi
	Dead         bool                // flag represent if block

	// helper attribute
	Id              BlockId
	ContaintTainted bool
	IsConditional   bool
	Visited         bool // used in path generator
}

func NewBlock(id BlockId) *Block {
	return &Block{
		Instructions:    make([]Op, 0),
		Preds:           make([]*Block, 0),
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
		Preds:           make([]*Block, 0),
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
	for _, pred := range b.Preds {
		if block == pred {
			return
		}
	}
	b.Preds = append(b.Preds, block)
}

func (b *Block) RemovePredecessor(block *Block) {
	idx := -1
	for i, pred := range b.Preds {
		if block == pred {
			idx = i
		}
	}
	if idx != -1 {
		b.Preds = append(b.Preds[:idx], b.Preds[idx+1:]...)
	}
}

func (b *Block) InPredecessors(block *Block) bool {
	for _, pred := range b.Preds {
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
