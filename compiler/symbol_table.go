package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
	LocalScope  SymbolScope = "LOCAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
	upper          *SymbolTable
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func NewSymbolTableWithUpper(upper *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.upper = upper
	return s
}

func (s *SymbolTable) Define(name string) Symbol {
	var scope SymbolScope
	if s.upper == nil {
		scope = GlobalScope
	} else {
		scope = LocalScope
	}
	sbl := Symbol{Name: name, Scope: scope, Index: s.numDefinitions}
	s.numDefinitions += 1
	s.store[sbl.Name] = sbl
	return sbl
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	sbl, ok := s.store[name]
	if !ok && s.upper != nil {
		return s.upper.Resolve(name)
	}
	return sbl, ok
}

func (s *SymbolTable) ResolveGlobal(name string) (Symbol, bool) {
	globalSymbolTable := s
	for globalSymbolTable.upper != nil {
		globalSymbolTable = globalSymbolTable.upper
	}
	sbl, ok := globalSymbolTable.store[name]
	return sbl, ok
}

func (s *SymbolTable) ResolveLocal(name string) (Symbol, bool) {
	sbl, ok := s.store[name]
	return sbl, ok
}
