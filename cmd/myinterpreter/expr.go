package main

// Expr is the base interface for all expression types
type Expr interface {
	Accept(visitor Visitor) any
}

// Visitor interface with methods for each expression type
type Visitor interface {
	VisitLiteralExpr(literal *Literal) any
	VisitBinaryExpr(binary *Binary) any
}

// Literal expression
type Literal struct {
	Value LoxLiteral
}

func (l *Literal) Accept(visitor Visitor) any {
	return visitor.VisitLiteralExpr(l)
}

// Binary expression
type Binary struct {
	Left  Expr
	Right Expr
	Op    string
}

func (b *Binary) Accept(visitor Visitor) any {
	return visitor.VisitBinaryExpr(b)
}

// Evaluator implements the Visitor interface
type Evaluator struct{}

func (e *Evaluator) VisitLiteralExpr(literal *Literal) any {
	return literal.Value
}

func (e *Evaluator) VisitBinaryExpr(binary *Binary) any {

	// Replace me later with something that actually use type inspection
	left := binary.Left.Accept(e).(float64)
	right := binary.Right.Accept(e).(float64)

	switch binary.Op {
	case "+":
		return left + right
	case "-":
		return left - right
	case "*":
		return left * right
	case "/":
		return left / right
	default:
		panic("unknown operator")
	}
}
