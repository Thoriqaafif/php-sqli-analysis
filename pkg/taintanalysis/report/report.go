package report

type ScanReport struct {
	Paths struct {
		Scanned []string `json:"scanned"`
	} `json:"paths"`
	Results []Result `json:"results"`
}

func NewScanReport(scannedPaths []string) *ScanReport {
	return &ScanReport{
		Paths: struct {
			Scanned []string `json:"scanned"`
		}{
			Scanned: scannedPaths,
		},
		Results: make([]Result, 0),
	}
}

func (s *ScanReport) AddPath(path string) {
	s.Paths.Scanned = append(s.Paths.Scanned, path)
}

func (s *ScanReport) AddResult(result Result) {
	s.Results = append(s.Results, result)
}

type Result struct {
	Start Loc    `json:"start"`
	End   Loc    `json:"end"`
	Path  string `json:"path"`
	Extra struct {
		DataFlowTrace struct {
			IntermediateVars []Node `json:"intermediate_vars"`
			TaintSource      Node   `json:"taint_source"`
			TaintSink        Node   `json:"taint_sink"`
		} `json:"dataflow_trace"`
		Message string `json:"message"`
	} `json:"extra"`
}

func NewResult(start, end Loc, path string) *Result {
	return &Result{
		Start: start,
		End:   end,
		Path:  path,
		Extra: struct {
			DataFlowTrace struct {
				IntermediateVars []Node `json:"intermediate_vars"`
				TaintSource      Node   `json:"taint_source"`
				TaintSink        Node   `json:"taint_sink"`
			} `json:"dataflow_trace"`
			Message string `json:"message"`
		}{
			DataFlowTrace: struct {
				IntermediateVars []Node `json:"intermediate_vars"`
				TaintSource      Node   `json:"taint_source"`
				TaintSink        Node   `json:"taint_sink"`
			}{
				IntermediateVars: make([]Node, 0),
			},
		},
	}
}

func (r *Result) SetSource(source Node) {
	r.Extra.DataFlowTrace.TaintSource = source
}

func (r *Result) SetSink(sink Node) {
	r.Extra.DataFlowTrace.TaintSink = sink
}

func (r *Result) AddIntermediateVar(vr Node) {
	r.Extra.DataFlowTrace.IntermediateVars = append(r.Extra.DataFlowTrace.IntermediateVars, vr)
}

func (r *Result) SetMessage(message string) {
	r.Extra.Message = message
}

func (r *Result) Clone() Result {
	intermediateVars := make([]Node, len(r.Extra.DataFlowTrace.IntermediateVars))
	copy(intermediateVars, r.Extra.DataFlowTrace.IntermediateVars)
	return Result{
		Start: r.Start,
		End:   r.End,
		Path:  r.Path,
		Extra: struct {
			DataFlowTrace struct {
				IntermediateVars []Node `json:"intermediate_vars"`
				TaintSource      Node   `json:"taint_source"`
				TaintSink        Node   `json:"taint_sink"`
			} `json:"dataflow_trace"`
			Message string `json:"message"`
		}{
			DataFlowTrace: struct {
				IntermediateVars []Node `json:"intermediate_vars"`
				TaintSource      Node   `json:"taint_source"`
				TaintSink        Node   `json:"taint_sink"`
			}{
				IntermediateVars: intermediateVars,
				TaintSource:      r.Extra.DataFlowTrace.TaintSource,
				TaintSink:        r.Extra.DataFlowTrace.TaintSink,
			},
		},
	}
}

type Loc struct {
	Line   int `json:"line"`
	Offset int `json:"offset"`
}

func NewLoc(line, offset int) Loc {
	return Loc{
		Line:   line,
		Offset: offset,
	}
}

type Node struct {
	Content  string `json:"content"`
	Location struct {
		Start Loc    `json:"start"`
		End   Loc    `json:"end"`
		Path  string `json:"path"`
	} `json:"location"`
}

func NewCodeNode(content, path string, start, end Loc) *Node {
	return &Node{
		Content: content,
		Location: struct {
			Start Loc    `json:"start"`
			End   Loc    `json:"end"`
			Path  string `json:"path"`
		}{
			Start: start,
			End:   end,
			Path:  path,
		},
	}
}
