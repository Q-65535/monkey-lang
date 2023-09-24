package ast

import (
	"bytes"
	"monkey/token"
	"strings"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	stateMentNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	// todo
}

type LetStatement struct {
	Token token.Token // token.LET
	Name *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode(){};
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal };
func (ls *LetStatement) String() string {
	// todo
}

// @Optimization: maybe we can just use token
type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode(){}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string { return i.Value}


type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode(){}
func (ls *LetStatement) TokenLiteral() string { return rs.Token.Literal };
func (rs *ReturnStatement) String() string {
	// todo
}
