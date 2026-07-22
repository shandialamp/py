package py

import (
	"fmt"
	"reflect"
	"strings"
)

// OnBuildHandler 构建成功后的回调函数类型
// sql: 生成的 SQL 语句, args: SQL 参数
type OnBuildHandler func(sql string, args []any)

// onBuildGlobals 全局构建回调列表
var onBuildGlobals []OnBuildHandler

// OnBuild 注册一个全局处理器，在每次 Build 成功后调用
// 可用于日志记录、SQL 审计、性能监控等
func OnBuild(handler OnBuildHandler) {
	onBuildGlobals = append(onBuildGlobals, handler)
}

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

	// insert
	insertCols []string
	insertVals [][]any

	// update — 用有序切片替代 map，保证 SET 列顺序
	setCols []setPair

	// insert from select
	selectQuery *QueryBuilder
}

type setPair struct {
	col string
	val any
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

// InsertInto 设置为 INSERT 语句，指定列名
//   py.InsertInto("users", "name", "email").Values("Alice", "alice@test.com")
func InsertInto(table string, cols ...string) *QueryBuilder {
	return &QueryBuilder{
		sqlType:    SqlTypeInsert,
		table:      table,
		insertCols: cols,
	}
}

// Insert 设置为 INSERT 语句，不指定列名（通过 Set 逐列添加）
//   py.Insert("users").Set("name", "Alice").Set("age", 18)
func Insert(table string) *QueryBuilder {
	return &QueryBuilder{
		sqlType: SqlTypeInsert,
		table:   table,
	}
}

// Columns 设置 INSERT 的列名（与 Values 搭配使用）
func (q *QueryBuilder) Columns(cols ...string) *QueryBuilder {
	q.insertCols = append(q.insertCols, cols...)
	return q
}

// Values 设置 INSERT 的值
func (q *QueryBuilder) Values(vals ...any) *QueryBuilder {
	cloned := make([]any, len(vals))
	copy(cloned, vals)
	q.insertVals = append(q.insertVals, cloned)
	return q
}

// SubQuery 用于 INSERT ... SELECT 子查询
//   py.InsertInto("users", "name", "email").SubQuery(
//       py.Table("temp_users").Select("name", "email").Where(py.Eq("status", 1)),
//   )
func (q *QueryBuilder) SubQuery(subQuery *QueryBuilder) *QueryBuilder {
	q.selectQuery = subQuery
	return q
}

// Update 设置为 UPDATE 语句
//   py.Update("users").Set("name", "Alice").Set("age", 20).Where(py.Eq("id", 1))
func Update(table string) *QueryBuilder {
	return &QueryBuilder{
		sqlType: SqlTypeUpdate,
		table:   table,
	}
}

// Set 设置 UPDATE 或 INSERT 的字段值
// UPDATE: py.Update("users").Set("name", "Alice").Set("age", 20)
// INSERT: py.Insert("users").Set("name", "Alice").Set("age", 18)
func (q *QueryBuilder) Set(col string, val any) *QueryBuilder {
	if q.sqlType == SqlTypeInsert {
		q.insertCols = append(q.insertCols, col)
	}
	q.setCols = append(q.setCols, setPair{col: col, val: val})
	return q
}

// SetStruct 通过结构体批量设置字段值
// 结构体中的 ModelField 字段必须标注 db 标签，SetStruct 会自动读取 Name 和 Value
//
//	user := py.NewModel[User]()
//	user.Name.Set("Alice")
//	user.Age.Set(18)
//
//	py.Insert("users").SetStruct(user).Build()
//	// INSERT INTO users (name, age) VALUES (?, ?)  -- "Alice", 18
//
//	py.Update("users").SetStruct(user).Where(py.Eq("id", 1)).Build()
//	// UPDATE users SET name = ?, age = ? WHERE (id = ?)  -- "Alice", 18, 1
func (q *QueryBuilder) SetStruct(m any) *QueryBuilder {
	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return q
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		ft := t.Field(i)

		// 只处理 ModelField 类型的字段
		if !strings.HasPrefix(f.Type().Name(), "ModelField") {
			continue
		}

		col := ft.Tag.Get("db")
		if col == "" {
			continue
		}

		// 获取 Value 字段
		valField := f.FieldByName("Value")
		if !valField.IsValid() {
			continue
		}

		q.Set(col, valField.Interface())
	}

	return q
}

// DeleteFrom 设置为 DELETE 语句
//   py.DeleteFrom("users").Where(py.Eq("id", 1))
func DeleteFrom(table string) *QueryBuilder {
	return &QueryBuilder{
		sqlType: SqlTypeDelete,
		table:   table,
	}
}

// Delete 设置为 DELETE 语句（简写）
//   py.Delete("users").Where(py.Eq("id", 1))
func Delete(table string) *QueryBuilder {
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
// 生成成功后，会依次调用所有通过 OnBuild 注册的全局处理器
func (q *QueryBuilder) Build() (string, []any) {
	var sql string
	var args []any

	switch q.sqlType {
	case SqlTypeSelect:
		sql, args = q.buildSelect()
	case SqlTypeInsert:
		sql, args = q.buildInsert()
	case SqlTypeUpdate:
		sql, args = q.buildUpdate()
	case SqlTypeDelete:
		sql, args = q.buildDelete()
	default:
		return "", nil
	}

	// 调用全局 OnBuild 处理器
	for _, handler := range onBuildGlobals {
		handler(sql, args)
	}

	return sql, args
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
	var b strings.Builder
	var args []any

	if q.comment != "" {
		b.WriteString(fmt.Sprintf("/* %s */ ", q.comment))
	}

	b.WriteString(fmt.Sprintf("INSERT INTO %s", q.table))

	// INSERT ... SELECT
	if q.selectQuery != nil {
		selectSQL, selectArgs := q.selectQuery.Build()
		if len(q.insertCols) > 0 {
			b.WriteString(fmt.Sprintf(" (%s)", strings.Join(q.insertCols, ", ")))
		}
		b.WriteString(fmt.Sprintf(" %s", selectSQL))
		args = append(args, selectArgs...)
		return b.String(), args
	}

	// 通过 Set() 收集的单行插入: Insert("t").Set("a",1).Set("b",2)
	if len(q.insertVals) == 0 && len(q.setCols) > 0 {
		var cols []string
		var vals []any
		for _, p := range q.setCols {
			cols = append(cols, p.col)
			vals = append(vals, p.val)
		}
		q.insertCols = cols
		q.insertVals = [][]any{vals}
	}

	if len(q.insertCols) == 0 || len(q.insertVals) == 0 {
		return "", nil
	}

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
	if len(q.setCols) == 0 {
		return "", nil
	}

	var b strings.Builder
	var args []any

	if q.comment != "" {
		b.WriteString(fmt.Sprintf("/* %s */ ", q.comment))
	}

	b.WriteString(fmt.Sprintf("UPDATE %s", q.table))

	var sets []string
	for _, p := range q.setCols {
		sets = append(sets, fmt.Sprintf("%s = ?", p.col))
		args = append(args, p.val)
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
