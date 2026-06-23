package py

// ==================== 条件表达式快捷函数 ====================

// Eq 生成 field = value 条件
func Eq(field string, value any) *SimpleExpr {
	return &SimpleExpr{Field: field, Value: value, Op: "="}
}

// Neq 生成 field != value 条件
func Neq(field string, value any) *SimpleExpr {
	return &SimpleExpr{Field: field, Value: value, Op: "!="}
}

// Gt 生成 field > value 条件
func Gt(field string, value any) *SimpleExpr {
	return &SimpleExpr{Field: field, Value: value, Op: ">"}
}

// Gte 生成 field >= value 条件
func Gte(field string, value any) *SimpleExpr {
	return &SimpleExpr{Field: field, Value: value, Op: ">="}
}

// Lt 生成 field < value 条件
func Lt(field string, value any) *SimpleExpr {
	return &SimpleExpr{Field: field, Value: value, Op: "<"}
}

// Lte 生成 field <= value 条件
func Lte(field string, value any) *SimpleExpr {
	return &SimpleExpr{Field: field, Value: value, Op: "<="}
}

// Like 生成 field LIKE value 条件
func Like(field string, value any) *SimpleExpr {
	return &SimpleExpr{Field: field, Value: value, Op: "LIKE"}
}

// NotLike 生成 field NOT LIKE value 条件
func NotLike(field string, value any) *SimpleExpr {
	return &SimpleExpr{Field: field, Value: value, Op: "NOT LIKE"}
}

// In 生成 field IN (values...) 条件
func In(field string, values ...any) *InExpr {
	return &InExpr{Field: field, Vals: values}
}

// NotIn 生成 field NOT IN (values...) 条件
func NotIn(field string, values ...any) *InExpr {
	return &InExpr{Field: field, Vals: values, Not: true}
}

// Between 生成 field BETWEEN lo AND hi 条件
func Between(field string, lo, hi any) *BetweenExpr {
	return &BetweenExpr{Field: field, Lo: lo, Hi: hi}
}

// NotBetween 生成 field NOT BETWEEN lo AND hi 条件
func NotBetween(field string, lo, hi any) *BetweenExpr {
	return &BetweenExpr{Field: field, Lo: lo, Hi: hi, Not: true}
}

// IsNull 生成 field IS NULL 条件
func IsNull(field string) *NullExpr {
	return &NullExpr{Field: field}
}

// IsNotNull 生成 field IS NOT NULL 条件
func IsNotNull(field string) *NullExpr {
	return &NullExpr{Field: field, Not: true}
}

// Exists 生成 EXISTS (subQuery) 条件
func Exists(subQuery *QueryBuilder) *ExistsExpr {
	return &ExistsExpr{Query: subQuery}
}

// NotExists 生成 NOT EXISTS (subQuery) 条件
func NotExists(subQuery *QueryBuilder) *ExistsExpr {
	return &ExistsExpr{Query: subQuery, Not: true}
}

// And 将多个条件用 AND 连接
func And(exprs ...Expr) *GroupExpr {
	return &GroupExpr{LogicOp: "AND", Exprs: exprs}
}

// Or 将多个条件用 OR 连接
func Or(exprs ...Expr) *GroupExpr {
	return &GroupExpr{LogicOp: "OR", Exprs: exprs}
}

// Not 对条件取反
func Not(expr Expr) *GroupExpr {
	return &GroupExpr{LogicOp: "AND", Exprs: []Expr{expr}, Not: true}
}
