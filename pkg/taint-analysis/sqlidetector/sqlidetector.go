package sqlidetector

type Detector struct{}

func NewDetector() *Detector {
	return &Detector{}
}

func (d *Detector) Scan() {
	// parsing
	// convert AST to MSSA
	// find feasible path
	// do taint analysis
}
