package sechat

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"
)

// astCall stores basic information about a function call.
type astCall struct {
	Name      string
	Arguments []ast.Expression
}

// astMap stores a simple object as a map.
type astMap map[string]interface{}

// parseJavaScript attempts to parse the JavaScript embedded on a page and
// create an abstract syntax tree (AST) from it.
func (c *Conn) parseJavaScript(urlStr string) (*ast.Program, error) {
	req, err := c.newRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(forceRedirect, "1")
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return parser.ParseFile(nil, "", doc.Find("script").Text(), 0)
}

// parseIdentifier converts identifiers (and DotExpressions) into their
// string representation.
func (c *Conn) parseIdentifier(exp ast.Expression) (string, bool) {
	switch t := exp.(type) {
	case *ast.Identifier:
		return t.Name, true
	case *ast.DotExpression:
		id, ok := c.parseIdentifier(t.Left)
		if !ok {
			return "", false
		}
		return id + "." + t.Identifier.Name, true
	default:
		return "", false
	}
}

// parseArray returns a list of expressions in an array.
func (c *Conn) parseArray(exp ast.Expression) ([]ast.Expression, bool) {
	arr, ok := exp.(*ast.ArrayLiteral)
	if !ok {
		return nil, false
	}
	return arr.Value, true
}

// parseMap parses the provided expression as a simple map.
func (c *Conn) parseMap(exp ast.Expression) astMap {
	obj, ok := exp.(*ast.ObjectLiteral)
	if !ok {
		return nil
	}
	m := make(astMap)
	for _, v := range obj.Value {
		switch t := v.Value.(type) {
		case *ast.BooleanLiteral:
			m[v.Key] = t.Value
		case *ast.NumberLiteral:
			m[v.Key] = t.Value
		case *ast.StringLiteral:
			m[v.Key] = t.Value
		}
	}
	return m
}

// parseFunctionCall attempts to parse a statement as a function call.
func (c *Conn) parseFunctionCall(stm ast.Statement) *astCall {
	exp, ok := stm.(*ast.ExpressionStatement)
	if !ok {
		return nil
	}
	call, ok := exp.Expression.(*ast.CallExpression)
	if !ok {
		return nil
	}
	name, ok := c.parseIdentifier(call.Callee)
	if !ok {
		return nil
	}
	return &astCall{
		Name:      name,
		Arguments: call.ArgumentList,
	}
}

// findOnReadyStatements searches for $(function) calls and returns a list of
// statements that they execute.
func (c *Conn) findOnReadyStatements(program *ast.Program) []ast.Statement {
	statements := []ast.Statement{}
	for _, stm := range program.Body {
		call := c.parseFunctionCall(stm)
		if call == nil {
			continue
		}
		if call.Name != "$" {
			continue
		}
		if len(call.Arguments) != 1 {
			continue
		}
		fn, ok := call.Arguments[0].(*ast.FunctionLiteral)
		if !ok {
			continue
		}
		blk, ok := fn.Body.(*ast.BlockStatement)
		if !ok {
			continue
		}
		statements = append(statements, blk.List...)
	}
	return statements
}
