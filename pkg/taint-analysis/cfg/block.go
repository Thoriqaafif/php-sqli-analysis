package cfg

type BlockID int

type Block struct {
	ID           BlockID
	Instructions []Op    // TODO: change type to instruction node
	Preds        []Block // predecessor of a basic block
	Phi          []Phi   // TODO: change type to phi
	Dead         bool    // TODO: understand the function
}

func NewBlock(id BlockID, preds []Block) *Block {
	return &Block{
		ID:    id,
		Preds: preds,
		Dead:  false,
	}
}

func (b *Block) AddPred(block Block) {
	for _, pred := range b.Preds {
		if block.ID == pred.ID {
			return
		}
	}
	b.Preds = append(b.Preds, block)
}
