package parser

import (
	"baboon/ast"
	"baboon/lexer"
	"baboon/token"
	"fmt"
	"strconv"
)

type Parser struct {
	l *lexer.Lexer

	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.prefixParseFns[token.IDENT] = p.parseIdentifier
	p.prefixParseFns[token.INT] = p.parseIntegerLiteral
	p.prefixParseFns[token.STRING] = p.parseStringLiteral
	p.prefixParseFns[token.TRUE] = p.parseBoolean
	p.prefixParseFns[token.FALSE] = p.parseBoolean
	p.prefixParseFns[token.BANG] = p.parsePrefixExpression
	p.prefixParseFns[token.MINUS] = p.parsePrefixExpression
	p.prefixParseFns[token.LPAREN] = p.parseGroupedExpression
	p.prefixParseFns[token.IF] = p.parseIfExpression
	p.prefixParseFns[token.FUNCTION] = p.parseFunctionExpression
	p.prefixParseFns[token.LBRACKET] = p.parseArrayLiteral

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.infixParseFns[token.PLUS] = p.parseInfixExpression
	p.infixParseFns[token.MINUS] = p.parseInfixExpression
	p.infixParseFns[token.SLASH] = p.parseInfixExpression
	p.infixParseFns[token.ASTERISK] = p.parseInfixExpression
	p.infixParseFns[token.EQ] = p.parseInfixExpression
	p.infixParseFns[token.NEQ] = p.parseInfixExpression
	p.infixParseFns[token.LT] = p.parseInfixExpression
	p.infixParseFns[token.GT] = p.parseInfixExpression
	p.infixParseFns[token.LEQ] = p.parseInfixExpression
	p.infixParseFns[token.GEQ] = p.parseInfixExpression
	p.infixParseFns[token.LPAREN] = p.parseCallExpression
	p.infixParseFns[token.LBRACKET] = p.parseAccessExpression

	// Read two tokens to set both curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// Public getter
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	pt := p.peekToken
	msg := fmt.Sprintf("[%d:%d] expected next token to be %q, got %q instead", pt.Line, pt.Column, t, pt.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.Token) {
	msg := fmt.Sprintf("[%d:%d] no prefix parse function for %q found", t.Line, t.Column, t.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt, ok := p.parseStatement()
		if ok {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() (ast.Statement, bool) {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() (*ast.LetStatement, bool) {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil, false
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil, false
	}
	p.nextToken()

	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil, false
	}

	stmt.Value = value

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return stmt, true
}

func (p *Parser) parseReturnStatement() (*ast.ReturnStatement, bool) {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil, false
	}

	stmt.Value = value

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return stmt, true
}

func (p *Parser) parseExpressionStatement() (*ast.ExpressionStatement, bool) {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	exp := p.parseExpression(LOWEST)
	if exp == nil {
		return nil, false
	}
	stmt.Expression = exp

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return stmt, true
}

func (p *Parser) parseBlockStatement() (*ast.BlockStatement, bool) {
	block := &ast.BlockStatement{Token: p.curToken, Statements: []ast.Statement{}}

	for p.peekToken.Type != token.RBRACE && p.peekToken.Type != token.EOF {
		// starting out with p.curToken.Type == token.LBRACE
		p.nextToken()
		stmt, ok := p.parseStatement()
		if ok {
			block.Statements = append(block.Statements, stmt)
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil, false
	}

	return block, true
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken)
		return nil
	}
	leftExp := prefix()

	for (p.peekToken.Type != token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		// Holly fuck, this call is incredibly important and non-obvious
		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("[%d:%d] could not parse %q as integer", p.curToken.Line, p.curToken.Column, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	items := p.parseExpressionsList(token.RBRACKET)
	if items == nil {
		return nil
	}
	array.Items = items

	return array
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curToken.Type == token.TRUE}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken() // eat the (

	exp := p.parseExpression(LOWEST)

	// eat the )
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	exp := &ast.IfExpression{Token: p.curToken}

	// eat IF
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken() // eat LPAREN

	exp.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	block, ok := p.parseBlockStatement()
	if !ok {
		return nil
	}
	exp.Consequence = block

	if p.peekToken.Type == token.ELSE {
		p.nextToken() // eat RBRACE
		// eat ELSE
		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		block, ok = p.parseBlockStatement()
		if !ok {
			return nil
		}
		exp.Alternative = block
	}

	return exp
}

func (p *Parser) parseFunctionExpression() ast.Expression {
	exp := &ast.FunctionExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	exp.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	body, ok := p.parseBlockStatement()
	if !ok {
		return nil
	}
	exp.Body = body

	return exp
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	params := []*ast.Identifier{}

	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return params
	}

	p.nextToken() // eat LPAREN

	param := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	params = append(params, param)

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()

		param := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		params = append(params, param)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return params
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	right := p.parseExpression(precedence)
	if right == nil {
		msg := fmt.Sprintf("[%d:%d] couldn't parse expression beginning with %q", p.curToken.Line, p.curToken.Column, p.curToken.Type)
		p.errors = append(p.errors, msg)
		return nil
	}
	expression.Right = right

	return expression
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionsList(token.RPAREN)
	return exp
}

func (p *Parser) parseAccessExpression(array ast.Expression) ast.Expression {
	exp := &ast.AccessExpression{Token: p.curToken, Array: array}

	if p.peekToken.Type == token.RBRACKET {
		msg := fmt.Sprintf("[%d:%d] unexpected RBRACKET in access expression", p.curToken.Line, p.curToken.Column)
		p.errors = append(p.errors, msg)
		return nil
	}

	p.nextToken()

	key := p.parseExpression(LOWEST)
	if key == nil {
		msg := fmt.Sprintf("[%d:%d] couldn't parse expression beginning with %q", p.curToken.Line, p.curToken.Column, p.curToken.Type)
		p.errors = append(p.errors, msg)
		return nil
	}
	exp.Key = key

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseExpressionsList(end token.TokenType) []ast.Expression {
	exps := []ast.Expression{}

	if p.peekToken.Type == end {
		p.nextToken()
		return exps
	}

	p.nextToken()

	exps = append(exps, p.parseExpression(LOWEST))
	for p.peekToken.Type == token.COMMA && p.peekToken.Type != token.EOF {
		p.nextToken()
		p.nextToken()
		exps = append(exps, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return exps
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // < >
	SUM         // +
	PRODUCT     // *
	PREFIX      // -x !x
	CALL        // fn()
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NEQ:      EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LEQ:      LESSGREATER,
	token.GEQ:      LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: CALL, // TODO: maybe change to higher?
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

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
		p.peekError(t)
		return false
	}
}
