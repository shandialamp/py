package py

import "fmt"

// Expr 条件表达式接口
type Expr interface {
	Build() (string, []any)
}

// SimpleExpr 简单条件表达式 a = ?
type SimpleExpr struct {
	Field string
	Value any
	Op    string // =, !=, >, <, >=, <=, LIKE, IN, NOT IN, IS, IS NOT
}

func (e *SimpleExpr) Build() (string, []any) {
	// 如果值是 RawExpr，直接嵌入 SQL 而不使用占位符
	if raw, ok := e.Value.(*RawExpr); ok {
		return fmt.Sprintf("%s %s %s", e.Field, e.Op, raw.SQL), raw.Args
	}
	return fmt.Sprintf("%s %s ?", e.Field, e.Op), []any{e.Value}
}

// RawExprWrapper 包装 RawExpr 实现 Expr 接口
type RawExprWrapper struct {
	Raw *RawExpr
}

func (e *RawExprWrapper) Build() (string, []any) {
	return e.Raw.SQL, e.Raw.Args
}

// GroupExpr 分组条件表达式 (a AND b AND c) 或 (a OR b OR c)
type GroupExpr struct {
	LogicOp string // AND / OR
	Exprs   []Expr
	Not     bool // NOT 取反
}

func (e *GroupExpr) Build() (string, []any) {
	if len(e.Exprs) == 0 {
		return "", nil
	}
	var args []any
	var segs []string
	for _, expr := range e.Exprs {
		sql, a := expr.Build()
		segs = append(segs, sql)
		args = append(args, a...)
	}
	sql := fmt.Sprintf("(%s)", joinWith(segs, fmt.Sprintf(" %s ", e.LogicOp)))
	if e.Not {
		sql = fmt.Sprintf("NOT %s", sql)
	}
	return sql, args
}

// BetweenExpr BETWEEN 表达式
type BetweenExpr struct {
	Field string
	Lo    any
	Hi    any
	Not   bool
}

func (e *BetweenExpr) Build() (string, []any) {
	op := "BETWEEN"
	if e.Not {
		op = "NOT BETWEEN"
	}
	return fmt.Sprintf("%s %s ? AND ?", e.Field, op), []any{e.Lo, e.Hi}
}

// InExpr IN 表达式
type InExpr struct {
	Field string
	Vals  []any
	Not   bool
}

func (e *InExpr) Build() (string, []any) {
	if len(e.Vals) == 0 {
		if e.Not {
			return "1=1", nil
		}
		return "1=0", nil
	}
	op := "IN"
	if e.Not {
		op = "NOT IN"
	}
	placeholders := joinWith(repeat("?", len(e.Vals)), ", ")
	return fmt.Sprintf("%s %s (%s)", e.Field, op, placeholders), e.Vals
}

// NullExpr IS NULL / IS NOT NULL 表达式
type NullExpr struct {
	Field string
	Not   bool
}

func (e *NullExpr) Build() (string, []any) {
	if e.Not {
		return fmt.Sprintf("%s IS NOT NULL", e.Field), nil
	}
	return fmt.Sprintf("%s IS NULL", e.Field), nil
}

// ExistsExpr EXISTS / NOT EXISTS 表达式
type ExistsExpr struct {
	Query *QueryBuilder
	Not   bool
}

func (e *ExistsExpr) Build() (string, []any) {
	sql, args := e.Query.Build()
	op := "EXISTS"
	if e.Not {
		op = "NOT EXISTS"
	}
	return fmt.Sprintf("%s (%s)", op, sql), args
}

// helper functions
func joinWith(items []string, sep string) string {
	if len(items) == 0 {
		return ""
	}
	result := items[0]
	for i := 1; i < len(items); i++ {
		result += sep + items[i]
	}
	return result
}

func repeat(s string, n int) []string {
	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = s
	}
	return result
}
