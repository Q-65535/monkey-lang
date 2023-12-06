package object

import (
	"bytes"
	"fmt"
	"monkey/ast"
	"strings"
)

type ObjectType string
type BuiltinFunction func(args ...Object) Object

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	STRING_OBJ       = "STRING"
	BUILTIN_OBJ      = "BUILTIN"
)

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s}
}

func NewCloseEnvironment(outer *Environment) *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: outer}
}

func (e *Environment) Get(key string) (Object, bool) {
	obj, ok := e.store[key]
	if ok {
		return obj, ok
	}
	if e.outer != nil {
		return e.outer.Get(key)
	}
	return obj, ok
}

func (e *Environment) Set(key string, val Object) {
	e.store[key] = val
}

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType {
	return INTEGER_OBJ
}
func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

type Boolean struct {
	Value bool
}

func (i *Boolean) Type() ObjectType {
	return BOOLEAN_OBJ
}
func (i *Boolean) Inspect() string {
	return fmt.Sprintf("%t", i.Value)
}

type Null struct{}

func (i *Null) Type() ObjectType {
	return NULL_OBJ
}
func (i *Null) Inspect() string {
	return "null"
}

type ReturnValue struct {
	Value Object
}

func (i *ReturnValue) Type() ObjectType {
	return RETURN_VALUE_OBJ
}
func (i *ReturnValue) Inspect() string {
	return fmt.Sprintf("return val: " + i.Value.Inspect())
}

type Error struct {
	ErrorMessage string
}

func (e *Error) Type() ObjectType {
	return ERROR_OBJ
}
func (i *Error) Inspect() string {
	return i.ErrorMessage
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType {
	return FUNCTION_OBJ
}

func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}

type String struct {
	Value string
}

func (s *String) Type() ObjectType {
	return STRING_OBJ
}
func (s *String) Inspect() string {
	return s.Value
}

type Builtin struct {
	Fn BuiltinFunction
}

func (s *Builtin) Type() ObjectType {
	return BUILTIN_OBJ
}
func (s *Builtin) Inspect() string {
	return "builtin function"
}
