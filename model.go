package py

import (
	"reflect"
	"strings"
)

// ModelField 泛型模型字段，用于将 Go 结构体字段映射到数据库列
//   - Value: 字段值
//   - Name:  数据库列名（由 NewModel 自动填充）
type ModelField[T any] struct {
	Value T
	Name  string
}

// Set 设置字段值
func (f *ModelField[T]) Set(value T) {
	f.Value = value
}

// NewModel 初始化模型，自动解析结构体中的 ModelField 字段并填充其 Name 为 db 标签值
//   - M: 模型结构体，包含 ModelField[T] 类型的字段
//   - 每个 ModelField 字段需标注 `db:"column_name"` 标签
//
// 用法:
//
//	type User struct {
//	    Id   py.ModelField[int64]  `db:"id"`
//	    Name py.ModelField[string] `db:"name"`
//	}
//
//	user := py.NewModel[User]()
//	user.Id.Set(1)
//	user.Name.Set("Alice")
func NewModel[M any]() *M {
	var m M

	v := reflect.ValueOf(&m).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !strings.HasPrefix(f.Type().Name(), "ModelField") {
			continue
		}

		column := t.Field(i).Tag.Get("db")
		f.FieldByName("Name").SetString(column)
	}

	return &m
}

// TableField 构建 table.field 格式的列名
//
//	py.TableField("users", "id") // "users.id"
func TableField(table, field string) string {
	return table + "." + field
}
