package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func (s *SymbolTable) Define(name string) Symbol {
	sbl := Symbol{Name: name, Scope: GlobalScope, Index: s.numDefinitions}
	s.numDefinitions += 1
	s.store[sbl.Name] = sbl
	return sbl
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	sbl, ok := s.store[name]
	return sbl, ok
}
