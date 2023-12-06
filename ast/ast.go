package ast

import (
	"bytes"
	"monkey/token"
	// "strings"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type ExpressionStatement struct {
	// @Question: why we need this token, how it is used?
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	} else {
		return ""
	}
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
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type LetStatement struct {
	Token token.Token // token.LET
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString("=")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// @Optimization: maybe we can just use token
type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (ls *ReturnStatement) TokenLiteral() string { return ls.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Altenative  *BlockStatement
}

func (is *IfExpression) expressionNode()      {}
func (is *IfExpression) TokenLiteral() string { return is.Token.Literal }
func (is *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString(is.TokenLiteral())
	out.WriteString("(")
	out.WriteString(is.Condition.String())
	out.WriteString(")")
	out.WriteString("\n")
	if is.Consequence != nil {
		out.WriteString("Consequence:\n")
		out.WriteString(is.Consequence.String())
	}
	if is.Altenative != nil {
		out.WriteString("else:\n")
		out.WriteString(is.Altenative.String())
	}
	return out.String()
}

type BlockStatement struct {
	Statements []Statement
}

func (b *BlockStatement) expressionNode() {}
func (b *BlockStatement) TokenLiteral() string {
	if len(b.Statements) > 0 {
		return b.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (b *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range b.Statements {
		out.WriteString(s.String() + "\n")
	}
	return out.String()
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return (bl.Token.Literal + "(tok)") }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (str *StringLiteral) expressionNode()      {}
func (str *StringLiteral) TokenLiteral() string { return str.Token.Literal }
func (str *StringLiteral) String() string       { return (str.Token.Literal + "(tok)") }

type FunctionLiteral struct {
	Token      token.Token // the fn token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	for _, par := range fl.Parameters {
		out.WriteString(par.Value + ",")
	}
	// out.WriteString(fl.Condition.String())
	out.WriteString(")")
	out.WriteString("\n")
	if fl.Body != nil {
		out.WriteString("Body:\n")
		out.WriteString(fl.Body.String())
	}
	return out.String()
}

type CallExpression struct {
	Token     token.Token // '(' token
	Function  Expression
	Arguments []Expression
}

func (c *CallExpression) expressionNode()      {}
func (c *CallExpression) TokenLiteral() string { return c.Token.Literal }
func (c *CallExpression) String() string {
	var out bytes.Buffer
	out.WriteString(c.Function.String())
	out.WriteString("(")
	for _, arg := range c.Arguments {
		out.WriteString(arg.String())
		// @Cleanup: this causes extra comma at the end
		out.WriteString(",")
	}
	out.WriteString(")")
	return out.String()
}

type PrefixExpression struct {
	Token    token.Token // the prefix token
	Operator string      // retrived from Token
	Right    Expression
}

func (p *PrefixExpression) expressionNode()      {}
func (p *PrefixExpression) TokenLiteral() string { return p.Token.Literal }
func (p *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(p.Operator)
	out.WriteString(p.Right.String())
	out.WriteString(")")
	return out.String()
}

type InfixExpression struct {
	Token    token.Token // the infix token (e.g., +, -)
	Operator string
	Left     Expression
	Right    Expression
}

func (p *InfixExpression) expressionNode()      {}
func (p *InfixExpression) TokenLiteral() string { return p.Token.Literal }
func (p *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(p.Left.String())
	out.WriteString(p.Operator)
	out.WriteString(p.Right.String())
	out.WriteString(")")
	return out.String()
}
