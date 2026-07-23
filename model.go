package py

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// ModelField 泛型模型字段，用于将 Go 结构体字段映射到数据库列
//   - _Value: 字段值（内部，通过 Get()/Set() 访问）
//   - Name:   数据库列名（由 NewModel 自动填充）
//
// 支持特性：
//   - database/sql.Scanner：支持 sqlx.Get、sqlx.Select、StructScan
//   - database/sql/driver.Valuer：支持数据库写入
//   - JSON 序反序列化
//   - 泛型类型安全转换
type ModelField[T any] struct {
	_Value T
	Name   string
}

// Set 设置字段值
func (f *ModelField[T]) Set(value T) {
	f._Value = value
}

// Get 获取字段值
// 推荐用此方法替代直接访问 ._Value
//
//	user.Id.Get()   // 推荐
func (f ModelField[T]) Get() T {
	return f._Value
}

// UnmarshalJSON 反序列化 JSON 数据到字段值
// 使用指针 receiver，支持修改原对象
func (f *ModelField[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &f._Value)
}

// MarshalJSON 序列化字段值为 JSON
func (f ModelField[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(f._Value)
}

// Scan 实现 sql.Scanner 接口，支持 sqlx 扫描数据库值
//
// 支持类型转换：
//   - MySQL bigint/int -> int64
//   - varchar/text -> []byte/string
//   - datetime -> time.Time
//   - NULL -> 零值
//
// 使用指针 receiver，确保修改原对象
func (f *ModelField[T]) Scan(src any) error {
	if src == nil {
		var zero T
		f._Value = zero
		return nil
	}

	// 目标类型（T 的实际类型）
	targetType := reflect.TypeOf((*T)(nil)).Elem()
	srcValue := reflect.ValueOf(src)

	// 情况 1：直接兼容赋值（无需转换）
	if srcValue.Type().AssignableTo(targetType) {
		f._Value = srcValue.Interface().(T)
		return nil
	}

	// 情况 2：类型可转换（如 int32 -> int64）
	if srcValue.Type().ConvertibleTo(targetType) {
		converted := reflect.New(targetType).Elem()
		converted.Set(srcValue.Convert(targetType))
		f._Value = converted.Interface().(T)
		return nil
	}

	// 情况 2.5：值类型转指针类型（如 int64 -> *int64）
	if targetType.Kind() == reflect.Ptr {
		elemType := targetType.Elem()

		// 源类型是否直接兼容或可转换到指向类型
		if srcValue.Type().AssignableTo(elemType) {
			// 直接赋值：src -> *elem
			ptr := reflect.New(elemType)
			ptr.Elem().Set(srcValue)
			f._Value = ptr.Interface().(T)
			return nil
		}

		if srcValue.Type().ConvertibleTo(elemType) {
			// 先转换再创建指针
			ptr := reflect.New(elemType)
			ptr.Elem().Set(srcValue.Convert(elemType))
			f._Value = ptr.Interface().(T)
			return nil
		}
	}

	// 情况 3：特殊处理 []byte 和 string
	switch v := src.(type) {
	case []byte:
		if err := f.scanBytes(v, targetType); err != nil {
			return err
		}
		return nil

	case string:
		if err := f.scanString(v, targetType); err != nil {
			return err
		}
		return nil
	}

	// 情况 4：通用反射转换
	return fmt.Errorf("cannot scan %v (type %v) into type %v", src, srcValue.Type(), targetType)
}

// scanBytes 处理 []byte -> T 的转换
// 支持 []byte -> string、[]byte -> *string 等指针类型
func (f *ModelField[T]) scanBytes(b []byte, targetType reflect.Type) error {
	s := string(b)

	// 情况 1：目标类型是 string
	if targetType.Kind() == reflect.String {
		var result any = s
		f._Value = result.(T)
		return nil
	}

	// 情况 2：目标类型是指针
	if targetType.Kind() == reflect.Ptr {
		elem := targetType.Elem()

		// 如果指向的是 string
		if elem.Kind() == reflect.String {
			result := &s
			f._Value = any(result).(T)
			return nil
		}
	}

	// 情况 3：通用反射尝试
	converted := reflect.New(targetType).Elem()
	srcValue := reflect.ValueOf(s)

	if srcValue.Type().AssignableTo(targetType) {
		converted.Set(srcValue)
		f._Value = converted.Interface().(T)
		return nil
	}

	return fmt.Errorf("cannot convert []byte to type %v", targetType)
}

// scanString 处理 string -> T 的转换
// 支持 string -> string、string -> *string 等指针类型
func (f *ModelField[T]) scanString(s string, targetType reflect.Type) error {
	// 情况 1：目标类型是 string
	if targetType.Kind() == reflect.String {
		var result any = s
		f._Value = result.(T)
		return nil
	}

	// 情况 2：目标类型是指针
	if targetType.Kind() == reflect.Ptr {
		elem := targetType.Elem()

		// 如果指向的是 string
		if elem.Kind() == reflect.String {
			result := &s
			f._Value = any(result).(T)
			return nil
		}
	}

	return fmt.Errorf("cannot convert string to type %v", targetType)
}

// Value 实现 driver.Valuer 接口，支持数据库写入
// 返回字段值给数据库驱动程序
func (f ModelField[T]) Value() (driver.Value, error) {
	return f._Value, nil
}

// NewModel 初始化模型，自动解析结构体中的 ModelField 字段并填充其 Name 为 db 标签值
//
// 识别规则：
//   - 字段类型必须是 struct
//   - 必须包含 Value 字段（任意类型）
//   - 必须包含 Name 字段（string 类型）
//   - 类型名称必须包含 "ModelField"
//   - 读取 `db:"column_name"` 标签并填充到 Name 字段
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
//	id := user.Id.Get()  // 1
func NewModel[M any]() *M {
	var m M

	v := reflect.ValueOf(&m).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := t.Field(i)

		// 判断是否是 ModelField 类型
		if !isModelField(fieldType.Type) {
			continue
		}

		// 获取 db 标签
		column := fieldType.Tag.Get("db")
		if column != "" {
			// 设置 Name 字段值
			nameField := fieldValue.FieldByName("Name")
			if nameField.IsValid() && nameField.CanSet() {
				nameField.SetString(column)
			}
		}
	}

	return &m
}

// isModelField 判断字段类型是否为 ModelField[T]
// 通过结构成员判断，而非简单 Type.Name() 判断，确保泛型安全
func isModelField(fieldType reflect.Type) bool {
	// 必须是 struct 类型
	if fieldType.Kind() != reflect.Struct {
		return false
	}

	// 必须包含 _Value 字段
	_, hasValue := fieldType.FieldByName("_Value")
	if !hasValue {
		return false
	}

	// 必须包含 Name 字段，且类型为 string
	nameField, hasName := fieldType.FieldByName("Name")
	if !hasName || nameField.Type.Kind() != reflect.String {
		return false
	}

	// 验证类型名称包含 "ModelField"（额外验证）
	if !strings.Contains(fieldType.String(), "ModelField") {
		return false
	}

	return true
}

// TableField 构建 table.field 格式的列名
//
//	py.TableField("users", "id") // "users.id"
func TableField(table, field string) string {
	return table + "." + field
}

// ModelColumns 获取模型中定义的所有 db 标签列名（以逗号分隔）
// 用于生成 SELECT 语句，避免 SELECT * 导致的字段不匹配问题
//
// 用法:
//
//	query := fmt.Sprintf("SELECT %s FROM table WHERE id = ?", py.ModelColumns[Goods]())
//	// 输出: SELECT id, name, price FROM table WHERE id = ?
func ModelColumns[M any]() string {
	var m M
	v := reflect.ValueOf(&m).Elem()
	t := v.Type()

	var columns []string

	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)

		// 判断是否是 ModelField 类型
		if !isModelField(fieldType.Type) {
			continue
		}

		// 获取 db 标签
		column := fieldType.Tag.Get("db")
		if column != "" {
			columns = append(columns, column)
		}
	}

	return strings.Join(columns, ", ")
}

// ModelColumnsWithTable 获取模型中定义的所有 db 标签列名，带表名前缀（以逗号分隔）
// 用于 JOIN 查询等需要限定列名的场景
//
// 用法:
//
//	query := fmt.Sprintf("SELECT %s FROM goods WHERE id = ?", py.ModelColumnsWithTable[Goods]("goods"))
//	// 输出: SELECT goods.id, goods.name, goods.price FROM goods WHERE id = ?
func ModelColumnsWithTable[M any](table string) string {
	var m M
	v := reflect.ValueOf(&m).Elem()
	t := v.Type()

	var columns []string

	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)

		// 判断是否是 ModelField 类型
		if !isModelField(fieldType.Type) {
			continue
		}

		// 获取 db 标签
		column := fieldType.Tag.Get("db")
		if column != "" {
			columns = append(columns, table+"."+column)
		}
	}

	return strings.Join(columns, ", ")
}
