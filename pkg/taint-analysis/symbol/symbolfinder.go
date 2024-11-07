package symbol

type SymbolFinder struct {
	VarTable map[string]VarID
	Vars     map[VarID]Var
}

// func
