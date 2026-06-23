package py

import (
	"fmt"
	"strings"
)

// QueryBuilder SQL 查询构建器
type QueryBuilder struct {
	sqlType    SqlType
	table      string
	columns    []string
	wheres     []Expr
	groups     []string
	havings    []Expr
	orders     []orderClause
	limit      *int
	offset     *int
	joins      []joinClause
	distinct   bool
	comment    string

	// insert/update
	insertCols []string
	insertVals [][]any
	setMap     map[string]any
}

type orderClause struct {
	field string
	dir   OrderDir
}

type joinClause struct {
	joinType JoinType
	table    string
	on       Expr
}

// Table 设置表名，开始构建查询
func Table(name string) *QueryBuilder {
	return &QueryBuilder{
		sqlType: SqlTypeSelect,
		table:   name,
	}
}

// Select 设置查询列
func (q *QueryBuilder) Select(columns ...string) *QueryBuilder {
	q.sqlType = SqlTypeSelect
	q.columns = append(q.columns, columns...)
	return q
}

// From 设置表名
func (q *QueryBuilder) From(table string) *QueryBuilder {
	q.table = table
	return q
}

// Distinct 设置 DISTINCT
func (q *QueryBuilder) Distinct() *QueryBuilder {
	q.distinct = true
	return q
}

// InsertInto 设置为 INSERT 语句
func InsertInto(table string, cols ...string) *QueryBuilder {
	return &QueryBuilder{
		sqlType:    SqlTypeInsert,
		table:      table,
		insertCols: cols,
	}
}

// Values 设置 INSERT 的值
func (q *QueryBuilder) Values(vals ...any) *QueryBuilder {
	cloned := make([]any, len(vals))
	copy(cloned, vals)
	q.insertVals = append(q.insertVals, cloned)
	return q
}

// Update 设置为 UPDATE 语句
func Update(table string) *QueryBuilder {
	return &QueryBuilder{
		sqlType: SqlTypeUpdate,
		table:   table,
		setMap:  make(map[string]any),
	}
}

// Set 设置 UPDATE 的字段值
func (q *QueryBuilder) Set(col string, val any) *QueryBuilder {
	if q.setMap == nil {
		q.setMap = make(map[string]any)
	}
	q.setMap[col] = val
	return q
}

// DeleteFrom 设置为 DELETE 语句
func DeleteFrom(table string) *QueryBuilder {
	return &QueryBuilder{
		sqlType: SqlTypeDelete,
		table:   table,
	}
}

// ==================== JOIN ====================

// Join INNER JOIN
func (q *QueryBuilder) Join(table string, on Expr) *QueryBuilder {
	q.joins = append(q.joins, joinClause{joinType: JoinTypeInner, table: table, on: on})
	return q
}

// LeftJoin LEFT JOIN
func (q *QueryBuilder) LeftJoin(table string, on Expr) *QueryBuilder {
	q.joins = append(q.joins, joinClause{joinType: JoinTypeLeft, table: table, on: on})
	return q
}

// RightJoin RIGHT JOIN
func (q *QueryBuilder) RightJoin(table string, on Expr) *QueryBuilder {
	q.joins = append(q.joins, joinClause{joinType: JoinTypeRight, table: table, on: on})
	return q
}

// CrossJoin CROSS JOIN
func (q *QueryBuilder) CrossJoin(table string) *QueryBuilder {
	q.joins = append(q.joins, joinClause{joinType: JoinTypeCross, table: table})
	return q
}

// ==================== WHERE ====================

// Where 添加 WHERE 条件 (AND 连接)
func (q *QueryBuilder) Where(exprs ...Expr) *QueryBuilder {
	for _, expr := range exprs {
		if expr != nil {
			q.wheres = append(q.wheres, expr)
		}
	}
	return q
}

// OrWhere 添加 OR WHERE 条件
func (q *QueryBuilder) OrWhere(exprs ...Expr) *QueryBuilder {
	if len(exprs) == 0 {
		return q
	}
	group := &GroupExpr{LogicOp: "OR", Exprs: exprs}
	q.wheres = append(q.wheres, group)
	return q
}

// When 条件分支: 当 condition 为 true 时执行 fn
func (q *QueryBuilder) When(condition bool, fn func(q *QueryBuilder) *QueryBuilder) *QueryBuilder {
	if condition && fn != nil {
		return fn(q)
	}
	return q
}

// ==================== GROUP BY / HAVING ====================

// GroupBy 添加 GROUP BY
func (q *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	q.groups = append(q.groups, columns...)
	return q
}

// Having 添加 HAVING 条件
func (q *QueryBuilder) Having(exprs ...Expr) *QueryBuilder {
	for _, expr := range exprs {
		if expr != nil {
			q.havings = append(q.havings, expr)
		}
	}
	return q
}

// ==================== ORDER BY ====================

// OrderBy 添加 ORDER BY
func (q *QueryBuilder) OrderBy(field string, dir ...OrderDir) *QueryBuilder {
	d := OrderAsc
	if len(dir) > 0 {
		d = dir[0]
	}
	q.orders = append(q.orders, orderClause{field: field, dir: d})
	return q
}

// ==================== LIMIT / OFFSET ====================

// Limit 设置 LIMIT
func (q *QueryBuilder) Limit(n int) *QueryBuilder {
	q.limit = &n
	return q
}

// Offset 设置 OFFSET
func (q *QueryBuilder) Offset(n int) *QueryBuilder {
	q.offset = &n
	return q
}

// ==================== Comment ====================

// Comment 添加 SQL 注释
func (q *QueryBuilder) Comment(comment string) *QueryBuilder {
	q.comment = comment
	return q
}

// ==================== 聚合函数 ====================

// Count 生成 COUNT 查询
func (q *QueryBuilder) Count(columns ...string) (string, []any) {
	col := "*"
	if len(columns) > 0 {
		col = columns[0]
	}
	return q.aggregate("COUNT", col)
}

// Sum 生成 SUM 查询
func (q *QueryBuilder) Sum(column string) (string, []any) {
	return q.aggregate("SUM", column)
}

// Avg 生成 AVG 查询
func (q *QueryBuilder) Avg(column string) (string, []any) {
	return q.aggregate("AVG", column)
}

// Max 生成 MAX 查询
func (q *QueryBuilder) Max(column string) (string, []any) {
	return q.aggregate("MAX", column)
}

// Min 生成 MIN 查询
func (q *QueryBuilder) Min(column string) (string, []any) {
	return q.aggregate("MIN", column)
}

func (q *QueryBuilder) aggregate(fn, column string) (string, []any) {
	q.columns = []string{fmt.Sprintf("%s(%s)", fn, column)}
	return q.Build()
}

// ==================== Build ====================

// Build 生成最终的 SQL 语句和参数
func (q *QueryBuilder) Build() (string, []any) {
	switch q.sqlType {
	case SqlTypeSelect:
		return q.buildSelect()
	case SqlTypeInsert:
		return q.buildInsert()
	case SqlTypeUpdate:
		return q.buildUpdate()
	case SqlTypeDelete:
		return q.buildDelete()
	default:
		return "", nil
	}
}

func (q *QueryBuilder) buildSelect() (string, []any) {
	var b strings.Builder
	var args []any

	// Comment
	if q.comment != "" {
		b.WriteString(fmt.Sprintf("/* %s */ ", q.comment))
	}

	b.WriteString("SELECT ")

	if q.distinct {
		b.WriteString("DISTINCT ")
	}

	if len(q.columns) > 0 {
		b.WriteString(strings.Join(q.columns, ", "))
	} else {
		b.WriteString("*")
	}

	if q.table != "" {
		b.WriteString(fmt.Sprintf(" FROM %s", q.table))
	}

	// JOIN
	for _, j := range q.joins {
		b.WriteString(fmt.Sprintf(" %s", joinTypeString(j.joinType)))
		b.WriteString(fmt.Sprintf(" %s", j.table))
		if j.on != nil {
			sql, a := j.on.Build()
			b.WriteString(fmt.Sprintf(" ON %s", sql))
			args = append(args, a...)
		}
	}

	// WHERE
	if len(q.wheres) > 0 {
		group := &GroupExpr{LogicOp: "AND", Exprs: q.wheres}
		sql, a := group.Build()
		b.WriteString(fmt.Sprintf(" WHERE %s", sql))
		args = append(args, a...)
	}

	// GROUP BY
	if len(q.groups) > 0 {
		b.WriteString(fmt.Sprintf(" GROUP BY %s", strings.Join(q.groups, ", ")))
	}

	// HAVING
	if len(q.havings) > 0 {
		group := &GroupExpr{LogicOp: "AND", Exprs: q.havings}
		sql, a := group.Build()
		b.WriteString(fmt.Sprintf(" HAVING %s", sql))
		args = append(args, a...)
	}

	// ORDER BY
	if len(q.orders) > 0 {
		var parts []string
		for _, o := range q.orders {
			dir := "ASC"
			if o.dir == OrderDesc {
				dir = "DESC"
			}
			parts = append(parts, fmt.Sprintf("%s %s", o.field, dir))
		}
		b.WriteString(fmt.Sprintf(" ORDER BY %s", strings.Join(parts, ", ")))
	}

	// LIMIT
	if q.limit != nil {
		b.WriteString(fmt.Sprintf(" LIMIT %d", *q.limit))
	}

	// OFFSET
	if q.offset != nil {
		b.WriteString(fmt.Sprintf(" OFFSET %d", *q.offset))
	}

	return b.String(), args
}

func (q *QueryBuilder) buildInsert() (string, []any) {
	if len(q.insertCols) == 0 || len(q.insertVals) == 0 {
		return "", nil
	}

	var b strings.Builder
	var args []any

	if q.comment != "" {
		b.WriteString(fmt.Sprintf("/* %s */ ", q.comment))
	}

	b.WriteString(fmt.Sprintf("INSERT INTO %s", q.table))
	b.WriteString(fmt.Sprintf(" (%s)", strings.Join(q.insertCols, ", ")))

	var rows []string
	for _, vals := range q.insertVals {
		rows = append(rows, fmt.Sprintf("(%s)", strings.Join(repeat("?", len(vals)), ", ")))
		args = append(args, vals...)
	}

	b.WriteString(fmt.Sprintf(" VALUES %s", strings.Join(rows, ", ")))
	return b.String(), args
}

func (q *QueryBuilder) buildUpdate() (string, []any) {
	if len(q.setMap) == 0 {
		return "", nil
	}

	var b strings.Builder
	var args []any

	if q.comment != "" {
		b.WriteString(fmt.Sprintf("/* %s */ ", q.comment))
	}

	b.WriteString(fmt.Sprintf("UPDATE %s", q.table))

	var sets []string
	for col, val := range q.setMap {
		sets = append(sets, fmt.Sprintf("%s = ?", col))
		args = append(args, val)
	}
	b.WriteString(fmt.Sprintf(" SET %s", strings.Join(sets, ", ")))

	// WHERE
	if len(q.wheres) > 0 {
		group := &GroupExpr{LogicOp: "AND", Exprs: q.wheres}
		sql, a := group.Build()
		b.WriteString(fmt.Sprintf(" WHERE %s", sql))
		args = append(args, a...)
	}

	return b.String(), args
}

func (q *QueryBuilder) buildDelete() (string, []any) {
	var b strings.Builder
	var args []any

	if q.comment != "" {
		b.WriteString(fmt.Sprintf("/* %s */ ", q.comment))
	}

	b.WriteString(fmt.Sprintf("DELETE FROM %s", q.table))

	// WHERE
	if len(q.wheres) > 0 {
		group := &GroupExpr{LogicOp: "AND", Exprs: q.wheres}
		sql, a := group.Build()
		b.WriteString(fmt.Sprintf(" WHERE %s", sql))
		args = append(args, a...)
	}

	return b.String(), args
}

func joinTypeString(t JoinType) string {
	switch t {
	case JoinTypeInner:
		return "INNER JOIN"
	case JoinTypeLeft:
		return "LEFT JOIN"
	case JoinTypeRight:
		return "RIGHT JOIN"
	case JoinTypeCross:
		return "CROSS JOIN"
	default:
		return "JOIN"
	}
}
