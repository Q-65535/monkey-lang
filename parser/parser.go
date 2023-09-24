package parser
}

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

const Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token
}

func New(l *Lexer.lexer) *Parser {
	p := &Parser{
		l: l,
	}
	p.l.nextToken()
	p.l.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// @problem: this should be called curTokenTypeIs
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		fmt.Printf("expect peek error\n")
		return false
	}
}


func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	for !p.curTokenIs(token.EOF) {
		smt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		// return p.parseExpressionStatement()
		return nil
	}
}

func (p *Parser) parseLetStatement() ast.LetStatement {
	stmt := LetStatement{Token: p.curToken}
	if !l.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{p.curToken, p.curToken.Literal}
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() ast.LetStatement {
	stmt := ReturnStatement{Token: p.curToken}
	p.nextToken()
	// skipping parsing expression process
	for !p.curTokenIs(Token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
