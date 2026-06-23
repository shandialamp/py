package py

// SqlType SQL 语句类型
type SqlType int

const (
	SqlTypeSelect SqlType = iota
	SqlTypeInsert
	SqlTypeUpdate
	SqlTypeDelete
)

// JoinType 连接类型
type JoinType int

const (
	JoinTypeInner JoinType = iota
	JoinTypeLeft
	JoinTypeRight
	JoinTypeCross
)

// OrderDir 排序方向
type OrderDir int

const (
	OrderAsc OrderDir = iota
	OrderDesc
)

// RawExpr 原始表达式，实现 Expr 接口
type RawExpr struct {
	SQL  string
	Args []any
}

func (e *RawExpr) Build() (string, []any) {
	return e.SQL, e.Args
}

// Raw 创建原始表达式
func Raw(sql string, args ...any) *RawExpr {
	return &RawExpr{SQL: sql, Args: args}
}
