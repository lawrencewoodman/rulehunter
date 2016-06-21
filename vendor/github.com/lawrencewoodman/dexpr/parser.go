// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the GO_LICENSE file.
//
// This parser is based on go's parser package, but has been customized to
// parse expressions suitable for dexpr.
//
// The parser accepts a larger language than is syntactically permitted by
// the Go spec, for simplicity, and for improved robustness in the presence
// of syntax errors. For instance, in method declarations, the receiver is
// treated like an ordinary parameter list and thus may contain multiple
// entries where the spec permits exactly one. Consequently, the corresponding
// field in the AST (ast.FuncDecl.Recv) field is not restricted to one entry.
//
package dexpr

import (
	"fmt"
	"go/ast"
	"go/scanner"
	"go/token"
	"strconv"
	"strings"
	"unicode"
)

// The parser structure holds the parser's internal state.
type parser struct {
	file    *token.File
	errors  scanner.ErrorList
	scanner scanner.Scanner

	// Next token
	pos token.Pos   // token position
	tok token.Token // one token look-ahead
	lit string      // token literal

	// Error recovery
	// (used to limit the number of calls to syncXXX functions
	// w/o making scanning progress - avoids potential endless
	// loops across multiple parser functions during error recovery)
	syncPos token.Pos // last synchronization position
	syncCnt int       // number of calls to syncXXX without progress

	// Non-syntactic parser control
	exprLev int  // < 0: in control clause, >= 0: in expression
	inRhs   bool // if set, the parser is parsing a rhs expression

	// Ordinary identifier scopes
	pkgScope   *ast.Scope        // pkgScope.Outer == nil
	topScope   *ast.Scope        // top-most scope; may be pkgScope
	unresolved []*ast.Ident      // unresolved identifiers
	imports    []*ast.ImportSpec // list of imports

	// Label scopes
	// (maintained by open/close LabelScope)
	labelScope  *ast.Scope     // label scope for current function
	targetStack [][]*ast.Ident // stack of unresolved labels
}

// ParseExpr obtains the AST of an expression x.
// The position information recorded in the AST is undefined. The filename used
// in error messages is the empty string.
func parseExpr(x string) (ast.Expr, error) {
	text := []byte(x)
	fset := token.NewFileSet()

	var p parser
	defer func() {
		if e := recover(); e != nil {
			// resume same panic if it's not a bailout
			if _, ok := e.(bailout); !ok {
				panic(e)
			}
		}
		p.errors.Sort()
	}()

	// parse expr
	p.init(fset, text)
	// Set up pkg-level scopes to avoid nil-pointer errors.
	// This is not needed for a correct expression x as the
	// parser will be ok with a nil topScope, but be cautious
	// in case of an erroneous x.
	p.openScope()
	p.pkgScope = p.topScope
	e := p.parseRhsOrType()
	p.closeScope()
	assert(p.topScope == nil, "unbalanced scopes")

	// If a semicolon was inserted, consume it;
	// report an error if there's more tokens.
	if p.tok == token.SEMICOLON && p.lit == "\n" {
		p.next()
	}
	p.expect(token.EOF)

	if p.errors.Len() > 0 {
		p.errors.Sort()
		return nil, p.errors.Err()
	}

	return e, nil
}

func (p *parser) init(fset *token.FileSet, src []byte) {
	p.file = fset.AddFile("", -1, len(src))
	eh := func(pos token.Position, msg string) { p.errors.Add(pos, msg) }
	p.scanner.Init(p.file, src, eh, 0)
	p.next()
}

// ----------------------------------------------------------------------------
// Scoping support

func (p *parser) openScope() {
	p.topScope = ast.NewScope(p.topScope)
}

func (p *parser) closeScope() {
	p.topScope = p.topScope.Outer
}

// The unresolved object is a sentinel to mark identifiers that have been added
// to the list of unresolved identifiers. The sentinel is only used for verifying
// internal consistency.
var unresolved = new(ast.Object)

// If x is an identifier, tryResolve attempts to resolve x by looking up
// the object it denotes. If no object is found and collectUnresolved is
// set, x is marked as unresolved and collected in the list of unresolved
// identifiers.
//
func (p *parser) tryResolve(x ast.Expr, collectUnresolved bool) {
	// nothing to do if x is not an identifier or the blank identifier
	ident, _ := x.(*ast.Ident)
	if ident == nil {
		return
	}
	assert(ident.Obj == nil, "identifier already declared or resolved")
	if ident.Name == "_" {
		return
	}
	// try to resolve the identifier
	for s := p.topScope; s != nil; s = s.Outer {
		if obj := s.Lookup(ident.Name); obj != nil {
			ident.Obj = obj
			return
		}
	}
	// all local scopes are known, so any unresolved identifier
	// must be found either in the file scope, package scope
	// (perhaps in another file), or universe scope --- collect
	// them so that they can be resolved later
	if collectUnresolved {
		ident.Obj = unresolved
		p.unresolved = append(p.unresolved, ident)
	}
}

func (p *parser) resolve(x ast.Expr) {
	p.tryResolve(x, true)
}

// ----------------------------------------------------------------------------
// Parsing support

// Advance to the next token.
func (p *parser) next() {
	p.pos, p.tok, p.lit = p.scanner.Scan()
	p.tok = keywordTokenToIdent(p.tok)
}

func keywordTokenToIdent(inToken token.Token) token.Token {
	keywords := []token.Token{
		token.BREAK, token.CASE, token.CHAN, token.CONST, token.CONTINUE,
		token.DEFAULT, token.DEFER, token.ELSE, token.FALLTHROUGH,
		token.FOR, token.FUNC, token.GO, token.GOTO, token.IF, token.IMPORT,
		token.INTERFACE, token.MAP, token.PACKAGE, token.RANGE, token.RETURN,
		token.SELECT, token.STRUCT, token.SWITCH, token.TYPE, token.VAR,
	}
	for _, t := range keywords {
		if t == inToken {
			return token.IDENT
		}
	}
	return inToken
}

// A bailout panic is raised to indicate early termination.
type bailout struct{}

func (p *parser) error(pos token.Pos, msg string) {
	epos := p.file.Position(pos)
	p.errors.Add(epos, msg)
}

func (p *parser) errorExpected(pos token.Pos, msg string) {
	msg = "expected " + msg
	if pos == p.pos {
		// the error happened at the current position;
		// make the error message more specific
		if p.tok == token.SEMICOLON && p.lit == "\n" {
			msg += ", found newline"
		} else {
			msg += ", found '" + p.tok.String() + "'"
			if p.tok.IsLiteral() {
				msg += " " + p.lit
			}
		}
	}
	p.error(pos, msg)
}

func (p *parser) expect(tok token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok {
		p.errorExpected(pos, "'"+tok.String()+"'")
	}
	p.next() // make progress
	return pos
}

// expectClosing is like expect but provides a better error message
// for the common case of a missing comma before a newline.
//
func (p *parser) expectClosing(tok token.Token, context string) token.Pos {
	if p.tok != tok && p.tok == token.SEMICOLON && p.lit == "\n" {
		p.error(p.pos, "missing ',' before newline in "+context)
		p.next()
	}
	return p.expect(tok)
}

func (p *parser) expectSemi() {
	// semicolon is optional before a closing ')' or '}'
	if p.tok != token.RPAREN && p.tok != token.RBRACE {
		switch p.tok {
		case token.COMMA:
			// permit a ',' instead of a ';' but complain
			p.errorExpected(p.pos, "';'")
			fallthrough
		case token.SEMICOLON:
			p.next()
		default:
			p.errorExpected(p.pos, "';'")
			syncStmt(p)
		}
	}
}

func (p *parser) atComma(context string, follow token.Token) bool {
	if p.tok == token.COMMA {
		return true
	}
	if p.tok != follow {
		msg := "missing ','"
		if p.tok == token.SEMICOLON && p.lit == "\n" {
			msg += " before newline"
		}
		p.error(p.pos, msg+" in "+context)
		return true // "insert" comma and continue
	}
	return false
}

func assert(cond bool, msg string) {
	if !cond {
		panic("go/parser internal error: " + msg)
	}
}

// syncStmt advances to the next statement.
// Used for synchronization after an error.
//
func syncStmt(p *parser) {
	for {
		switch p.tok {
		case token.BREAK, token.CONST, token.CONTINUE, token.DEFER,
			token.FALLTHROUGH, token.FOR, token.GO, token.GOTO,
			token.IF, token.RETURN, token.SELECT, token.SWITCH,
			token.TYPE, token.VAR:
			// Return only if parser made some progress since last
			// sync or if it has not reached 10 sync calls without
			// progress. Otherwise consume at least one token to
			// avoid an endless parser loop (it is possible that
			// both parseOperand and parseStmt call syncStmt and
			// correctly do not advance, thus the need for the
			// invocation limit p.syncCnt).
			if p.pos == p.syncPos && p.syncCnt < 10 {
				p.syncCnt++
				return
			}
			if p.pos > p.syncPos {
				p.syncPos = p.pos
				p.syncCnt = 0
				return
			}
			// Reaching here indicates a parser bug, likely an
			// incorrect token list in this function, but it only
			// leads to skipping of possibly correct code if a
			// previous error is present, and thus is preferred
			// over a non-terminating parse.
		case token.EOF:
			return
		}
		p.next()
	}
}

// safePos returns a valid file position for a given position: If pos
// is valid to begin with, safePos returns pos. If pos is out-of-range,
// safePos returns the EOF position.
//
// This is hack to work around "artificial" end positions in the AST which
// are computed by adding 1 to (presumably valid) token positions. If the
// token positions are invalid due to parse errors, the resulting end position
// may be past the file's EOF position, which would lead to panics if used
// later on.
//
func (p *parser) safePos(pos token.Pos) (res token.Pos) {
	defer func() {
		if recover() != nil {
			res = token.Pos(p.file.Base() + p.file.Size()) // EOF position
		}
	}()
	_ = p.file.Offset(pos) // trigger a panic if position is out-of-range
	return pos
}

// ----------------------------------------------------------------------------
// Identifiers

func (p *parser) parseIdent() *ast.Ident {
	pos := p.pos
	name := "_"
	if p.tok == token.IDENT {
		name = p.lit
		p.next()
	} else {
		p.expect(token.IDENT) // use expect() error handling
	}
	return &ast.Ident{NamePos: pos, Name: name}
}

func (p *parser) parseIdentList() (list []*ast.Ident) {
	list = append(list, p.parseIdent())
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.parseIdent())
	}

	return
}

// ----------------------------------------------------------------------------
// Common productions

// If lhs is set, result list elements which are identifiers are not resolved.
func (p *parser) parseExprList(lhs bool) (list []ast.Expr) {
	list = append(list, p.checkExpr(p.parseExpr(lhs)))
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.checkExpr(p.parseExpr(lhs)))
	}

	return
}

func (p *parser) parseLhsList() []ast.Expr {
	old := p.inRhs
	p.inRhs = false
	list := p.parseExprList(true)
	switch p.tok {
	case token.DEFINE:
		// lhs of a short variable declaration
		// but doesn't enter scope until later:
		// caller must call p.shortVarDecl(p.makeIdentList(list))
		// at appropriate time.
	case token.COLON:
		// lhs of a label declaration or a communication clause of a select
		// statement (parseLhsList is not called when parsing the case clause
		// of a switch statement):
		// - labels are declared by the caller of parseLhsList
		// - for communication clauses, if there is a stand-alone identifier
		//   followed by a colon, we have a syntax error; there is no need
		//   to resolve the identifier in that case
	default:
		// identifiers must be declared elsewhere
		for _, x := range list {
			p.resolve(x)
		}
	}
	p.inRhs = old
	return list
}

func (p *parser) parseRhsList() []ast.Expr {
	old := p.inRhs
	p.inRhs = true
	list := p.parseExprList(false)
	p.inRhs = old
	return list
}

// ----------------------------------------------------------------------------
// Types

func (p *parser) parseType() ast.Expr {
	typ := p.tryType()

	if typ == nil {
		pos := p.pos
		p.errorExpected(pos, "type")
		p.next() // make progress
		return &ast.BadExpr{From: pos, To: p.pos}
	}

	return typ
}

// If the result is an identifier, it is not resolved.
func (p *parser) parseTypeName() ast.Expr {
	ident := p.parseIdent()
	// don't resolve ident yet - it may be a parameter or field name

	if p.tok == token.PERIOD {
		// ident is a package name
		p.next()
		p.resolve(ident)
		sel := p.parseIdent()
		return &ast.SelectorExpr{X: ident, Sel: sel}
	}

	return ident
}

func (p *parser) parseArrayType() ast.Expr {
	lbrack := p.expect(token.LBRACK)
	p.exprLev++
	var len ast.Expr
	// always permit ellipsis for more fault-tolerant parsing
	if p.tok == token.ELLIPSIS {
		len = &ast.Ellipsis{Ellipsis: p.pos}
		p.next()
	} else if p.tok != token.RBRACK {
		len = p.parseRhs()
	}
	p.exprLev--
	p.expect(token.RBRACK)
	elt := p.parseType()

	return &ast.ArrayType{Lbrack: lbrack, Len: len, Elt: elt}
}

func (p *parser) makeIdentList(list []ast.Expr) []*ast.Ident {
	idents := make([]*ast.Ident, len(list))
	for i, x := range list {
		ident, isIdent := x.(*ast.Ident)
		if !isIdent {
			if _, isBad := x.(*ast.BadExpr); !isBad {
				// only report error if it's a new one
				p.errorExpected(x.Pos(), "identifier")
			}
			ident = &ast.Ident{NamePos: x.Pos(), Name: "_"}
		}
		idents[i] = ident
	}
	return idents
}

func (p *parser) parsePointerType() *ast.StarExpr {
	star := p.expect(token.MUL)
	base := p.parseType()

	return &ast.StarExpr{Star: star, X: base}
}

// If the result is an identifier, it is not resolved.
func (p *parser) tryVarType(isParam bool) ast.Expr {
	if isParam && p.tok == token.ELLIPSIS {
		pos := p.pos
		p.next()
		typ := p.tryIdentOrType() // don't use parseType so we can provide better error message
		if typ != nil {
			p.resolve(typ)
		} else {
			p.error(pos, "'...' parameter is missing type")
			typ = &ast.BadExpr{From: pos, To: p.pos}
		}
		return &ast.Ellipsis{Ellipsis: pos, Elt: typ}
	}
	return p.tryIdentOrType()
}

// If the result is an identifier, it is not resolved.
func (p *parser) parseVarType(isParam bool) ast.Expr {
	typ := p.tryVarType(isParam)
	if typ == nil {
		pos := p.pos
		p.errorExpected(pos, "type")
		p.next() // make progress
		typ = &ast.BadExpr{From: pos, To: p.pos}
	}
	return typ
}

func (p *parser) parseParameterList(scope *ast.Scope, ellipsisOk bool) (params []*ast.Field) {
	// 1st ParameterDecl
	// A list of identifiers looks like a list of type names.
	var list []ast.Expr
	for {
		list = append(list, p.parseVarType(ellipsisOk))
		if p.tok != token.COMMA {
			break
		}
		p.next()
		if p.tok == token.RPAREN {
			break
		}
	}

	// analyze case
	if typ := p.tryVarType(ellipsisOk); typ != nil {
		// IdentifierList Type
		idents := p.makeIdentList(list)
		field := &ast.Field{Names: idents, Type: typ}
		params = append(params, field)
		// Go spec: The scope of an identifier denoting a function
		// parameter or result variable is the function body.
		p.resolve(typ)
		if !p.atComma("parameter list", token.RPAREN) {
			return
		}
		p.next()
		for p.tok != token.RPAREN && p.tok != token.EOF {
			idents := p.parseIdentList()
			typ := p.parseVarType(ellipsisOk)
			field := &ast.Field{Names: idents, Type: typ}
			params = append(params, field)
			// Go spec: The scope of an identifier denoting a function
			// parameter or result variable is the function body.
			p.resolve(typ)
			if !p.atComma("parameter list", token.RPAREN) {
				break
			}
			p.next()
		}
		return
	}

	// Type { "," Type } (anonymous parameters)
	params = make([]*ast.Field, len(list))
	for i, typ := range list {
		p.resolve(typ)
		params[i] = &ast.Field{Type: typ}
	}
	return
}

func (p *parser) parseParameters(scope *ast.Scope, ellipsisOk bool) *ast.FieldList {
	var params []*ast.Field
	lparen := p.expect(token.LPAREN)
	if p.tok != token.RPAREN {
		params = p.parseParameterList(scope, ellipsisOk)
	}
	rparen := p.expect(token.RPAREN)

	return &ast.FieldList{Opening: lparen, List: params, Closing: rparen}
}

func (p *parser) parseResult(scope *ast.Scope) *ast.FieldList {
	if p.tok == token.LPAREN {
		return p.parseParameters(scope, false)
	}

	typ := p.tryType()
	if typ != nil {
		list := make([]*ast.Field, 1)
		list[0] = &ast.Field{Type: typ}
		return &ast.FieldList{List: list}
	}

	return nil
}

func (p *parser) parseSignature(scope *ast.Scope) (params, results *ast.FieldList) {
	params = p.parseParameters(scope, true)
	results = p.parseResult(scope)

	return
}

func (p *parser) parseFuncType() (*ast.FuncType, *ast.Scope) {
	pos := p.expect(token.FUNC)
	scope := ast.NewScope(p.topScope) // function scope
	params, results := p.parseSignature(scope)

	return &ast.FuncType{Func: pos, Params: params, Results: results}, scope
}

func (p *parser) parseMapType() *ast.MapType {
	pos := p.expect(token.MAP)
	p.expect(token.LBRACK)
	key := p.parseType()
	p.expect(token.RBRACK)
	value := p.parseType()

	return &ast.MapType{Map: pos, Key: key, Value: value}
}

// If the result is an identifier, it is not resolved.
func (p *parser) tryIdentOrType() ast.Expr {
	switch p.tok {
	case token.IDENT:
		return p.parseTypeName()
	case token.LBRACK:
		return p.parseArrayType()
	case token.MUL:
		return p.parsePointerType()
	case token.FUNC:
		typ, _ := p.parseFuncType()
		return typ
	case token.MAP:
		return p.parseMapType()
	case token.LPAREN:
		lparen := p.pos
		p.next()
		typ := p.parseType()
		rparen := p.expect(token.RPAREN)
		return &ast.ParenExpr{Lparen: lparen, X: typ, Rparen: rparen}
	}

	// no type found
	return nil
}

func (p *parser) tryType() ast.Expr {
	typ := p.tryIdentOrType()
	if typ != nil {
		p.resolve(typ)
	}
	return typ
}

// ----------------------------------------------------------------------------
// Blocks

func (p *parser) parseStmtList() (list []ast.Stmt) {
	for p.tok != token.CASE && p.tok != token.RBRACE && p.tok != token.EOF {
		list = append(list, p.parseStmt())
	}

	return
}

func (p *parser) parseBody(scope *ast.Scope) *ast.BlockStmt {
	lbrace := p.expect(token.LBRACE)
	p.topScope = scope // open function scope
	list := p.parseStmtList()
	p.closeScope()
	rbrace := p.expect(token.RBRACE)

	return &ast.BlockStmt{Lbrace: lbrace, List: list, Rbrace: rbrace}
}

func (p *parser) parseBlockStmt() *ast.BlockStmt {
	lbrace := p.expect(token.LBRACE)
	p.openScope()
	list := p.parseStmtList()
	p.closeScope()
	rbrace := p.expect(token.RBRACE)

	return &ast.BlockStmt{Lbrace: lbrace, List: list, Rbrace: rbrace}
}

// ----------------------------------------------------------------------------
// Expressions

func (p *parser) parseFuncTypeOrLit() ast.Expr {
	typ, scope := p.parseFuncType()
	if p.tok != token.LBRACE {
		// function type only
		return typ
	}

	p.exprLev++
	body := p.parseBody(scope)
	p.exprLev--

	return &ast.FuncLit{Type: typ, Body: body}
}

// parseOperand may return an expression or a raw type (incl. array
// types of the form [...]T. Callers must verify the result.
// If lhs is set and the result is an identifier, it is not resolved.
//
func (p *parser) parseOperand(lhs bool) ast.Expr {
	switch p.tok {
	case token.IDENT:
		x := p.parseIdent()
		if !lhs {
			p.resolve(x)
		}
		return x
	case token.INT, token.FLOAT, token.IMAG, token.CHAR, token.STRING:
		x := &ast.BasicLit{ValuePos: p.pos, Kind: p.tok, Value: p.lit}
		p.next()
		return x

	case token.LPAREN:
		lparen := p.pos
		p.next()
		p.exprLev++
		x := p.parseRhsOrType() // types may be parenthesized: (some type)
		p.exprLev--
		rparen := p.expect(token.RPAREN)
		return &ast.ParenExpr{Lparen: lparen, X: x, Rparen: rparen}

	case token.FUNC:
		return p.parseFuncTypeOrLit()
	}

	if typ := p.tryIdentOrType(); typ != nil {
		// could be type for composite literal or conversion
		_, isIdent := typ.(*ast.Ident)
		assert(!isIdent, "type cannot be identifier")
		return typ
	}

	// we have an error
	pos := p.pos
	p.errorExpected(pos, "operand")
	syncStmt(p)
	return &ast.BadExpr{From: pos, To: p.pos}
}

func (p *parser) parseSelector(x ast.Expr) ast.Expr {
	sel := p.parseIdent()

	return &ast.SelectorExpr{X: x, Sel: sel}
}

func (p *parser) parseTypeAssertion(x ast.Expr) ast.Expr {
	lparen := p.expect(token.LPAREN)
	var typ ast.Expr
	if p.tok == token.TYPE {
		// type switch: typ == nil
		p.next()
	} else {
		typ = p.parseType()
	}
	rparen := p.expect(token.RPAREN)

	return &ast.TypeAssertExpr{X: x, Type: typ, Lparen: lparen, Rparen: rparen}
}

func (p *parser) parseIndexOrSlice(x ast.Expr) ast.Expr {
	const N = 3 // change the 3 to 2 to disable 3-index slices
	lbrack := p.expect(token.LBRACK)
	p.exprLev++
	var index [N]ast.Expr
	var colons [N - 1]token.Pos
	if p.tok != token.COLON {
		index[0] = p.parseRhs()
	}
	ncolons := 0
	for p.tok == token.COLON && ncolons < len(colons) {
		colons[ncolons] = p.pos
		ncolons++
		p.next()
		if p.tok != token.COLON && p.tok != token.RBRACK && p.tok != token.EOF {
			index[ncolons] = p.parseRhs()
		}
	}
	p.exprLev--
	rbrack := p.expect(token.RBRACK)

	if ncolons > 0 {
		// slice expression
		slice3 := false
		if ncolons == 2 {
			slice3 = true
			// Check presence of 2nd and 3rd index here rather than during type-checking
			// to prevent erroneous programs from passing through gofmt (was issue 7305).
			if index[1] == nil {
				p.error(colons[0], "2nd index required in 3-index slice")
				index[1] = &ast.BadExpr{From: colons[0] + 1, To: colons[1]}
			}
			if index[2] == nil {
				p.error(colons[1], "3rd index required in 3-index slice")
				index[2] = &ast.BadExpr{From: colons[1] + 1, To: rbrack}
			}
		}
		return &ast.SliceExpr{X: x, Lbrack: lbrack, Low: index[0], High: index[1], Max: index[2], Slice3: slice3, Rbrack: rbrack}
	}

	return &ast.IndexExpr{X: x, Lbrack: lbrack, Index: index[0], Rbrack: rbrack}
}

func (p *parser) parseCallOrConversion(fun ast.Expr) *ast.CallExpr {
	lparen := p.expect(token.LPAREN)
	p.exprLev++
	var list []ast.Expr
	var ellipsis token.Pos
	for p.tok != token.RPAREN && p.tok != token.EOF && !ellipsis.IsValid() {
		list = append(list, p.parseRhsOrType()) // builtins may expect a type: make(some type, ...)
		if p.tok == token.ELLIPSIS {
			ellipsis = p.pos
			p.next()
		}
		if !p.atComma("argument list", token.RPAREN) {
			break
		}
		p.next()
	}
	p.exprLev--
	rparen := p.expectClosing(token.RPAREN, "argument list")

	return &ast.CallExpr{Fun: fun, Lparen: lparen, Args: list, Ellipsis: ellipsis, Rparen: rparen}
}

func (p *parser) parseValue(keyOk bool) ast.Expr {
	if p.tok == token.LBRACE {
		return p.parseLiteralValue(nil)
	}

	// Because the parser doesn't know the composite literal type, it cannot
	// know if a key that's an identifier is a struct field name or a name
	// denoting a value. The former is not resolved by the parser or the
	// resolver.
	//
	// Instead, _try_ to resolve such a key if possible. If it resolves,
	// it a) has correctly resolved, or b) incorrectly resolved because
	// the key is a struct field with a name matching another identifier.
	// In the former case we are done, and in the latter case we don't
	// care because the type checker will do a separate field lookup.
	//
	// If the key does not resolve, it a) must be defined at the top
	// level in another file of the same package, the universe scope, or be
	// undeclared; or b) it is a struct field. In the former case, the type
	// checker can do a top-level lookup, and in the latter case it will do
	// a separate field lookup.
	x := p.checkExpr(p.parseExpr(keyOk))
	if keyOk {
		if p.tok == token.COLON {
			// Try to resolve the key but don't collect it
			// as unresolved identifier if it fails so that
			// we don't get (possibly false) errors about
			// undeclared names.
			p.tryResolve(x, false)
		} else {
			// not a key
			p.resolve(x)
		}
	}

	return x
}

func (p *parser) parseElement() ast.Expr {
	x := p.parseValue(true)
	if p.tok == token.COLON {
		colon := p.pos
		p.next()
		x = &ast.KeyValueExpr{Key: x, Colon: colon, Value: p.parseValue(false)}
	}

	return x
}

func (p *parser) parseElementList() (list []ast.Expr) {
	for p.tok != token.RBRACE && p.tok != token.EOF {
		list = append(list, p.parseElement())
		if !p.atComma("composite literal", token.RBRACE) {
			break
		}
		p.next()
	}

	return
}

func (p *parser) parseLiteralValue(typ ast.Expr) ast.Expr {
	lbrace := p.expect(token.LBRACE)
	var elts []ast.Expr
	p.exprLev++
	if p.tok != token.RBRACE {
		elts = p.parseElementList()
	}
	p.exprLev--
	rbrace := p.expectClosing(token.RBRACE, "composite literal")
	return &ast.CompositeLit{Type: typ, Lbrace: lbrace, Elts: elts, Rbrace: rbrace}
}

// checkExpr checks that x is an expression (and not a type).
func (p *parser) checkExpr(x ast.Expr) ast.Expr {
	switch unparen(x).(type) {
	case *ast.BadExpr:
	case *ast.Ident:
	case *ast.BasicLit:
	case *ast.FuncLit:
	case *ast.CompositeLit:
	case *ast.ParenExpr:
		panic("unreachable")
	case *ast.SelectorExpr:
	case *ast.IndexExpr:
	case *ast.SliceExpr:
	case *ast.TypeAssertExpr:
		// If t.Type == nil we have a type assertion of the form
		// y.(type), which is only allowed in type switch expressions.
		// It's hard to exclude those but for the case where we are in
		// a type switch. Instead be lenient and test this in the type
		// checker.
	case *ast.CallExpr:
	case *ast.StarExpr:
	case *ast.UnaryExpr:
	case *ast.BinaryExpr:
	default:
		// all other nodes are not proper expressions
		p.errorExpected(x.Pos(), "expression")
		x = &ast.BadExpr{From: x.Pos(), To: p.safePos(x.End())}
	}
	return x
}

// isTypeName reports whether x is a (qualified) TypeName.
func isTypeName(x ast.Expr) bool {
	switch t := x.(type) {
	case *ast.BadExpr:
	case *ast.Ident:
	case *ast.SelectorExpr:
		_, isIdent := t.X.(*ast.Ident)
		return isIdent
	default:
		return false // all other nodes are not type names
	}
	return true
}

// isLiteralType reports whether x is a legal composite literal type.
func isLiteralType(x ast.Expr) bool {
	switch t := x.(type) {
	case *ast.BadExpr:
	case *ast.Ident:
	case *ast.SelectorExpr:
		_, isIdent := t.X.(*ast.Ident)
		return isIdent
	case *ast.ArrayType:
	case *ast.StructType:
	case *ast.MapType:
	default:
		return false // all other nodes are not legal composite literal types
	}
	return true
}

// If x is of the form *T, deref returns T, otherwise it returns x.
func deref(x ast.Expr) ast.Expr {
	if p, isPtr := x.(*ast.StarExpr); isPtr {
		x = p.X
	}
	return x
}

// If x is of the form (T), unparen returns unparen(T), otherwise it returns x.
func unparen(x ast.Expr) ast.Expr {
	if p, isParen := x.(*ast.ParenExpr); isParen {
		x = unparen(p.X)
	}
	return x
}

// checkExprOrType checks that x is an expression or a type
// (and not a raw type such as [...]T).
//
func (p *parser) checkExprOrType(x ast.Expr) ast.Expr {
	switch t := unparen(x).(type) {
	case *ast.ParenExpr:
		panic("unreachable")
	case *ast.UnaryExpr:
	case *ast.ArrayType:
		if len, isEllipsis := t.Len.(*ast.Ellipsis); isEllipsis {
			p.error(len.Pos(), "expected array length, found '...'")
			x = &ast.BadExpr{From: x.Pos(), To: p.safePos(x.End())}
		}
	}

	// all other nodes are expressions or types
	return x
}

// If lhs is set and the result is an identifier, it is not resolved.
func (p *parser) parsePrimaryExpr(lhs bool) ast.Expr {
	x := p.parseOperand(lhs)
L:
	for {
		switch p.tok {
		case token.PERIOD:
			p.next()
			if lhs {
				p.resolve(x)
			}
			switch p.tok {
			case token.IDENT:
				x = p.parseSelector(p.checkExprOrType(x))
			case token.LPAREN:
				x = p.parseTypeAssertion(p.checkExpr(x))
			default:
				pos := p.pos
				p.errorExpected(pos, "selector or type assertion")
				p.next() // make progress
				sel := &ast.Ident{NamePos: pos, Name: "_"}
				x = &ast.SelectorExpr{X: x, Sel: sel}
			}
		case token.LBRACK:
			if lhs {
				p.resolve(x)
			}
			x = p.parseIndexOrSlice(p.checkExpr(x))
		case token.LPAREN:
			if lhs {
				p.resolve(x)
			}
			x = p.parseCallOrConversion(p.checkExprOrType(x))
		case token.LBRACE:
			if isLiteralType(x) && (p.exprLev >= 0 || !isTypeName(x)) {
				if lhs {
					p.resolve(x)
				}
				x = p.parseLiteralValue(x)
			} else {
				break L
			}
		default:
			break L
		}
		lhs = false // no need to try to resolve again
	}

	return x
}

// If lhs is set and the result is an identifier, it is not resolved.
func (p *parser) parseUnaryExpr(lhs bool) ast.Expr {
	switch p.tok {
	case token.ADD, token.SUB, token.NOT, token.XOR, token.AND:
		pos, op := p.pos, p.tok
		p.next()
		x := p.parseUnaryExpr(false)
		return &ast.UnaryExpr{OpPos: pos, Op: op, X: p.checkExpr(x)}
	case token.MUL:
		// pointer type or unary "*" expression
		pos := p.pos
		p.next()
		x := p.parseUnaryExpr(false)
		return &ast.StarExpr{Star: pos, X: p.checkExprOrType(x)}
	}

	return p.parsePrimaryExpr(lhs)
}

func (p *parser) tokPrec() (token.Token, int) {
	tok := p.tok
	if p.inRhs && tok == token.ASSIGN {
		tok = token.EQL
	}
	return tok, tok.Precedence()
}

// If lhs is set and the result is an identifier, it is not resolved.
func (p *parser) parseBinaryExpr(lhs bool, prec1 int) ast.Expr {
	x := p.parseUnaryExpr(lhs)
	for {
		op, oprec := p.tokPrec()
		if oprec < prec1 {
			return x
		}
		pos := p.expect(op)
		if lhs {
			p.resolve(x)
			lhs = false
		}
		y := p.parseBinaryExpr(false, oprec+1)
		x = &ast.BinaryExpr{X: p.checkExpr(x), OpPos: pos, Op: op, Y: p.checkExpr(y)}
	}
}

// If lhs is set and the result is an identifier, it is not resolved.
// The result may be a type or even a raw type ([...]int). Callers must
// check the result (using checkExpr or checkExprOrType), depending on
// context.
func (p *parser) parseExpr(lhs bool) ast.Expr {
	return p.parseBinaryExpr(lhs, token.LowestPrec+1)
}

func (p *parser) parseRhs() ast.Expr {
	old := p.inRhs
	p.inRhs = true
	x := p.checkExpr(p.parseExpr(false))
	p.inRhs = old
	return x
}

func (p *parser) parseRhsOrType() ast.Expr {
	old := p.inRhs
	p.inRhs = true
	x := p.checkExprOrType(p.parseExpr(false))
	p.inRhs = old
	return x
}

// ----------------------------------------------------------------------------
// Statements

// Parsing modes for parseSimpleStmt.
const (
	basic = iota
	labelOk
	rangeOk
)

// parseSimpleStmt returns true as 2nd result if it parsed the assignment
// of a range clause (with mode == rangeOk). The returned statement is an
// assignment with a right-hand side that is a single unary expression of
// the form "range x". No guarantees are given for the left-hand side.
func (p *parser) parseSimpleStmt(mode int) (ast.Stmt, bool) {
	x := p.parseLhsList()

	switch p.tok {
	case
		token.DEFINE, token.ASSIGN, token.ADD_ASSIGN,
		token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN,
		token.REM_ASSIGN, token.AND_ASSIGN, token.OR_ASSIGN,
		token.XOR_ASSIGN, token.SHL_ASSIGN, token.SHR_ASSIGN, token.AND_NOT_ASSIGN:
		// assignment statement, possibly part of a range clause
		pos, tok := p.pos, p.tok
		p.next()
		var y []ast.Expr
		isRange := false
		if mode == rangeOk && p.tok == token.RANGE && (tok == token.DEFINE || tok == token.ASSIGN) {
			pos := p.pos
			p.next()
			y = []ast.Expr{&ast.UnaryExpr{OpPos: pos, Op: token.RANGE, X: p.parseRhs()}}
			isRange = true
		} else {
			y = p.parseRhsList()
		}
		as := &ast.AssignStmt{Lhs: x, TokPos: pos, Tok: tok, Rhs: y}
		return as, isRange
	}

	if len(x) > 1 {
		p.errorExpected(x[0].Pos(), "1 expression")
		// continue with first expression
	}

	switch p.tok {
	case token.COLON:
		// labeled statement
		colon := p.pos
		p.next()
		if label, isIdent := x[0].(*ast.Ident); mode == labelOk && isIdent {
			// Go spec: The scope of a label is the body of the function
			// in which it is declared and excludes the body of any nested
			// function.
			stmt := &ast.LabeledStmt{Label: label, Colon: colon, Stmt: p.parseStmt()}
			return stmt, false
		}
		// The label declaration typically starts at x[0].Pos(), but the label
		// declaration may be erroneous due to a token after that position (and
		// before the ':'). If SpuriousErrors is not set, the (only) error re-
		// ported for the line is the illegal label error instead of the token
		// before the ':' that caused the problem. Thus, use the (latest) colon
		// position for error reporting.
		p.error(colon, "illegal label declaration")
		return &ast.BadStmt{From: x[0].Pos(), To: colon + 1}, false

	case token.INC, token.DEC:
		// increment or decrement
		s := &ast.IncDecStmt{X: x[0], TokPos: p.pos, Tok: p.tok}
		p.next()
		return s, false
	}

	// expression
	return &ast.ExprStmt{X: x[0]}, false
}

func (p *parser) parseCallExpr(callType string) *ast.CallExpr {
	x := p.parseRhsOrType() // could be a conversion: (some type)(x)
	if call, isCall := x.(*ast.CallExpr); isCall {
		return call
	}
	if _, isBad := x.(*ast.BadExpr); !isBad {
		// only report error if it's a new one
		p.error(p.safePos(x.End()), fmt.Sprintf("function must be invoked in %s statement", callType))
	}
	return nil
}

func (p *parser) parseGoStmt() ast.Stmt {
	pos := p.expect(token.GO)
	call := p.parseCallExpr("go")
	p.expectSemi()
	if call == nil {
		return &ast.BadStmt{From: pos, To: pos + 2} // len("go")
	}

	return &ast.GoStmt{Go: pos, Call: call}
}

func (p *parser) makeExpr(s ast.Stmt, kind string) ast.Expr {
	if s == nil {
		return nil
	}
	if es, isExpr := s.(*ast.ExprStmt); isExpr {
		return p.checkExpr(es.X)
	}
	p.error(s.Pos(), fmt.Sprintf("expected %s, found simple statement (missing parentheses around composite literal?)", kind))
	return &ast.BadExpr{From: s.Pos(), To: p.safePos(s.End())}
}

func (p *parser) parseTypeList() (list []ast.Expr) {
	list = append(list, p.parseType())
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.parseType())
	}

	return
}

func isTypeSwitchAssert(x ast.Expr) bool {
	a, ok := x.(*ast.TypeAssertExpr)
	return ok && a.Type == nil
}

func (p *parser) isTypeSwitchGuard(s ast.Stmt) bool {
	switch t := s.(type) {
	case *ast.ExprStmt:
		// x.(type)
		return isTypeSwitchAssert(t.X)
	case *ast.AssignStmt:
		// v := x.(type)
		if len(t.Lhs) == 1 && len(t.Rhs) == 1 && isTypeSwitchAssert(t.Rhs[0]) {
			switch t.Tok {
			case token.ASSIGN:
				// permit v = x.(type) but complain
				p.error(t.TokPos, "expected ':=', found '='")
				fallthrough
			case token.DEFINE:
				return true
			}
		}
	}
	return false
}

func (p *parser) parseStmt() (s ast.Stmt) {
	switch p.tok {
	case token.CONST, token.TYPE, token.VAR:
		s = &ast.DeclStmt{Decl: p.parseDecl(syncStmt)}
	case
		// tokens that may start an expression
		token.IDENT, token.INT, token.FLOAT, token.IMAG, token.CHAR, token.STRING, token.FUNC, token.LPAREN, // operands
		token.LBRACK, token.STRUCT, token.MAP, token.CHAN, token.INTERFACE, // composite types
		token.ADD, token.SUB, token.MUL, token.AND, token.XOR, token.NOT: // unary operators
		s, _ = p.parseSimpleStmt(labelOk)
		// because of the required look-ahead, labeled statements are
		// parsed by parseSimpleStmt - don't expect a semicolon after
		// them
		if _, isLabeledStmt := s.(*ast.LabeledStmt); !isLabeledStmt {
			p.expectSemi()
		}
	case token.GO:
		s = p.parseGoStmt()
	case token.LBRACE:
		s = p.parseBlockStmt()
		p.expectSemi()
	case token.SEMICOLON:
		// Is it ever possible to have an implicit semicolon
		// producing an empty statement in a valid program?
		// (handle correctly anyway)
		s = &ast.EmptyStmt{Semicolon: p.pos, Implicit: p.lit == "\n"}
		p.next()
	case token.RBRACE:
		// a semicolon may be omitted before a closing "}"
		s = &ast.EmptyStmt{Semicolon: p.pos, Implicit: true}
	default:
		// no statement found
		pos := p.pos
		p.errorExpected(pos, "statement")
		syncStmt(p)
		s = &ast.BadStmt{From: pos, To: p.pos}
	}

	return
}

// ----------------------------------------------------------------------------
// Declarations

type parseSpecFunction func(doc *ast.CommentGroup, keyword token.Token, iota int) ast.Spec

func isValidImport(lit string) bool {
	const illegalChars = `!"#$%&'()*,:;<=>?[\]^{|}` + "`\uFFFD"
	s, _ := strconv.Unquote(lit) // go/scanner returns a legal string literal
	for _, r := range s {
		if !unicode.IsGraphic(r) || unicode.IsSpace(r) || strings.ContainsRune(illegalChars, r) {
			return false
		}
	}
	return s != ""
}

func (p *parser) parseValueSpec(doc *ast.CommentGroup, keyword token.Token, iota int) ast.Spec {
	pos := p.pos
	idents := p.parseIdentList()
	typ := p.tryType()
	var values []ast.Expr
	// always permit optional initialization for more tolerant parsing
	if p.tok == token.ASSIGN {
		p.next()
		values = p.parseRhsList()
	}
	p.expectSemi() // call before accessing p.linecomment

	switch keyword {
	case token.VAR:
		if typ == nil && values == nil {
			p.error(pos, "missing variable type or initialization")
		}
	case token.CONST:
		if values == nil && (iota == 0 || typ != nil) {
			p.error(pos, "missing constant value")
		}
	}

	// Go spec: The scope of a constant or variable identifier declared inside
	// a function begins at the end of the ConstSpec or VarSpec and ends at
	// the end of the innermost containing block.
	// (Global identifiers are resolved in a separate phase after parsing.)
	spec := &ast.ValueSpec{
		Doc:    doc,
		Names:  idents,
		Type:   typ,
		Values: values,
	}
	return spec
}

func (p *parser) parseGenDecl(keyword token.Token, f parseSpecFunction) *ast.GenDecl {
	pos := p.expect(keyword)
	var lparen, rparen token.Pos
	var list []ast.Spec
	if p.tok == token.LPAREN {
		lparen = p.pos
		p.next()
		for iota := 0; p.tok != token.RPAREN && p.tok != token.EOF; iota++ {
			list = append(list, f(nil, keyword, iota))
		}
		rparen = p.expect(token.RPAREN)
		p.expectSemi()
	} else {
		list = append(list, f(nil, keyword, 0))
	}

	return &ast.GenDecl{
		TokPos: pos,
		Tok:    keyword,
		Lparen: lparen,
		Specs:  list,
		Rparen: rparen,
	}
}

func (p *parser) parseDecl(sync func(*parser)) ast.Decl {
	var f parseSpecFunction
	switch p.tok {
	case token.CONST, token.VAR:
		f = p.parseValueSpec

	default:
		pos := p.pos
		p.errorExpected(pos, "declaration")
		sync(p)
		return &ast.BadDecl{From: pos, To: p.pos}
	}

	return p.parseGenDecl(p.tok, f)
}
