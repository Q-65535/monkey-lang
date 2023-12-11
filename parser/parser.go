package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
	ARRAYACCESS
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: ARRAYACCESS,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l: l,
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionExpression)
	p.registerPrefix(token.LPAREN, p.parseLParen)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)

	// register infix parsing functions
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseArrayAccessExpression)

	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) addError(format string, a ...interface{}) {
	p.errors = append(p.errors, fmt.Sprintf(format, a...))
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
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
		fmt.Printf("expect peek error, expect: '%s' got: '%s'\n", t, p.peekToken.Type)
		return false
	}
}
func (p *Parser) peekPrecedence() int {
	if pre, ok := precedences[p.peekToken.Type]; ok {
		return pre
	}
	// default precedence is LOWEST
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if pre, ok := precedences[p.curToken.Type]; ok {
		return pre
	}
	// default precedence is LOWEST
	return LOWEST
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
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
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		msg := fmt.Sprintf("parsing let statement error: the token after 'let' is not an identifier but: %s", p.peekToken)
		p.errors = append(p.errors, msg)
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		msg := fmt.Sprintf("parsing let statement error: the token after identifier is not '=' but: %s", p.peekToken)
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseIfExpression() ast.Expression {
	exp := &ast.IfExpression{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		msg := fmt.Sprintf("the token after if is not (, but: %s", p.peekToken)
		p.errors = append(p.errors, msg)
		return nil
	}
	exp.Condition = p.parseLParen()
	if !p.expectPeek(token.LBRACE) {
		msg := fmt.Sprintf("the token after if condition expression is not {, but: %s", p.peekToken)
		p.errors = append(p.errors, msg)
		return nil
	}
	exp.Consequence = p.parseLbrace()
	if !p.peekTokenIs(token.ELSE) {
		return exp
	} else {
		// @Robustness: we should here also use expectPeek to check left brace token and else token
		p.nextToken()
		p.nextToken()
		exp.Altenative = p.parseLbrace()
	}
	return exp
}

func (p *Parser) parseFunctionExpression() ast.Expression {
	exp := &ast.FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		p.addError("parsing function error: the token after if is not left paren, but: %s\n", p.peekToken)
		return nil
	}
	exp.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		p.addError("parsing function error: the token after function parameter is not '{', but: %s\n", p.peekToken)
		return nil
	}
	exp.Body = p.parseLbrace()
	return exp
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	parameters := []*ast.Identifier{}
	// no parameters, just return
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return parameters
	}
	// skip '(' token
	p.nextToken()
	par := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	parameters = append(parameters, par)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		par := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		parameters = append(parameters, par)
	}
	if !p.expectPeek(token.RPAREN) {
		msg := fmt.Sprintf("parsing function error: the token after prameters is not ')', but: %s", p.peekToken)
		p.errors = append(p.errors, msg)
		return nil
	}
	return parameters
}

func (p *Parser) parseLbrace() *ast.BlockStatement {
	block := &ast.BlockStatement{}
	block.Statements = []ast.Statement{}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	exp := &ast.CallExpression{
		Token:    p.curToken,
		Function: left,
	}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}
	// skip '(' token
	p.nextToken()
	arg := p.parseExpression(LOWEST)
	args = append(args, arg)
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		arg = p.parseExpression(LOWEST)
		args = append(args, arg)
	}
	if !p.expectPeek(token.RPAREN) {
		msg := fmt.Sprintf("parsing call error: the token after arguments is not ')' but: %s", p.peekToken)
		p.errors = append(p.errors, msg)
		return nil
	}
	return args
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.addError("no function for parsing token of type %s", p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	// @Logic: the for loop condition could cause a problem when prefix() is for parsing if or fn expression,
	// because there is no semicolon after the expression and might go into the for loop, which is not
	// what we want to see.
	for !p.peekTokenIs(token.SEMICOLON) && p.peekPrecedence() > precedence {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) parseLParen() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		p.addError("parsing left paren expession: the token after expression is not ')' but: %s\n", p.peekToken)
		return nil
	}
	return exp
}

func (p *Parser) parseIdentifier() ast.Expression {
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	return ident
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %s as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}
func (p *Parser) parseBooleanLiteral() ast.Expression {
	lit := &ast.BooleanLiteral{Token: p.curToken}
	var val bool
	if lit.Token.Type == token.FALSE {
		val = false
	} else if lit.Token.Type == token.TRUE {
		val = true
	} else {
		fmt.Printf("parsing boolean..., but current token is not a boolean token!")
	}
	lit.Value = val
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	arr := &ast.ArrayLiteral{Token: p.curToken, Elements: []ast.Expression{}}
	// empty array
	if p.peekTokenIs(token.RBRACKET) {
		p.nextToken()
		return arr
	}
	// skip '[' token
	p.nextToken()
	arr.Elements = append(arr.Elements, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		arr.Elements = append(arr.Elements, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(token.RBRACKET) {
		p.addError("parsing array literal error, expect ] as the end of exrepssion, but got %s\n", p.peekToken)
	}
	return arr
}

func (p *Parser) parseArrayAccessExpression(left ast.Expression) ast.Expression {
	ac := &ast.ArrayAccessExpression{Token: p.curToken, Array: left}
	if !p.expectPeek(token.INT) {
		p.addError("parsing array access error, expect integer as index, but got %s\n", p.peekToken)
	}
	ac.Index = p.parseIntegerLiteral()
	if !p.expectPeek(token.RBRACKET) {
		p.addError("parsing array access error, expect ] as the end of expression, but got %s\n", p.peekToken)
	}
	return ac
}

// @Problem: the function name is really confusing, parsePrefixExpression is one of a group of functions
// and the group itself is called prefixParseFns.
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(LOWEST)
	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}
