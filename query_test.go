package py

import (
	"fmt"
	"testing"
)

func assertSQL(t *testing.T, expectedSQL string, expectedArgs []any, sql string, args []any) {
	t.Helper()
	if sql != expectedSQL {
		t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", sql, expectedSQL)
	}
	if len(args) != len(expectedArgs) {
		t.Errorf("Args length mismatch: got %d, want %d\n  got:  %v\n  want: %v",
			len(args), len(expectedArgs), args, expectedArgs)
		return
	}
	for i := range args {
		if args[i] != expectedArgs[i] {
			t.Errorf("Arg[%d] mismatch: got %v, want %v", i, args[i], expectedArgs[i])
		}
	}
}

func TestData(t *testing.T) {
	status := 1
	name := ""
	// SELECT id, name, type, config, desc, status, created_at, updated_at FROM data_sources
	sql, args := Table("data_sources").
		Select("id", "name", "type", "config", "desc", "status", "created_at", "updated_at").
		When(name != "", func(q *QueryBuilder) *QueryBuilder {
			return q.Where(Like("name", name))
		}).
		When(status != 0, func(q *QueryBuilder) *QueryBuilder {
			return q.Where(Eq("status", status))
		}).
		Build()

	fmt.Printf("SQL:  %s\n", sql)
	fmt.Printf("Args: %v\n", args)
}

func TestSelectBasic(t *testing.T) {
	sql, args := Table("users").Select("id", "name").Build()
	assertSQL(t, "SELECT id, name FROM users", nil, sql, args)
}

func TestSelectAll(t *testing.T) {
	sql, args := Table("users").Build()
	assertSQL(t, "SELECT * FROM users", nil, sql, args)
}

func TestSelectDistinct(t *testing.T) {
	sql, args := Table("users").Select("name").Distinct().Build()
	assertSQL(t, "SELECT DISTINCT name FROM users", nil, sql, args)
}

func TestWhere(t *testing.T) {
	sql, args := Table("users").
		Where(Eq("name", "Alice"), Gt("age", 18)).
		Build()
	assertSQL(t, "SELECT * FROM users WHERE (name = ? AND age > ?)", []any{"Alice", 18}, sql, args)
}

func TestOrWhere(t *testing.T) {
	sql, args := Table("users").
		Where(Eq("status", 1)).
		OrWhere(Eq("role", "admin"), Eq("role", "manager")).
		Build()
	assertSQL(t, "SELECT * FROM users WHERE (status = ? AND (role = ? OR role = ?))",
		[]any{1, "admin", "manager"}, sql, args)
}

func TestGroupByHaving(t *testing.T) {
	sql, args := Table("orders").
		Select("user_id", "COUNT(*) as total").
		GroupBy("user_id").
		Having(Gt("total", 5)).
		Build()
	assertSQL(t, "SELECT user_id, COUNT(*) as total FROM orders GROUP BY user_id HAVING (total > ?)",
		[]any{5}, sql, args)
}

func TestOrderByLimitOffset(t *testing.T) {
	sql, args := Table("users").
		OrderBy("created_at", OrderDesc).
		OrderBy("id").
		Limit(10).
		Offset(20).
		Build()
	assertSQL(t, "SELECT * FROM users ORDER BY created_at DESC, id ASC LIMIT 10 OFFSET 20", nil, sql, args)
}

func TestCount(t *testing.T) {
	sql, args := Table("users").Where(Eq("status", 1)).Count()
	assertSQL(t, "SELECT COUNT(*) FROM users WHERE (status = ?)", []any{1}, sql, args)
}

func TestSum(t *testing.T) {
	sql, args := Table("orders").Where(Eq("user_id", 1)).Sum("amount")
	assertSQL(t, "SELECT SUM(amount) FROM orders WHERE (user_id = ?)", []any{1}, sql, args)
}

func TestInsert(t *testing.T) {
	sql, args := InsertInto("users", "name", "age").
		Values("Alice", 18).
		Values("Bob", 20).
		Build()
	assertSQL(t, "INSERT INTO users (name, age) VALUES (?, ?), (?, ?)",
		[]any{"Alice", 18, "Bob", 20}, sql, args)
}

func TestUpdate(t *testing.T) {
	sql, args := Update("users").
		Set("name", "Alice").
		Set("age", 20).
		Where(Eq("id", 1)).
		Build()
	assertSQL(t, "UPDATE users SET name = ?, age = ? WHERE (id = ?)",
		[]any{"Alice", 20, 1}, sql, args)
}

func TestDelete(t *testing.T) {
	sql, args := DeleteFrom("users").Where(Eq("id", 1)).Build()
	assertSQL(t, "DELETE FROM users WHERE (id = ?)", []any{1}, sql, args)
}

func TestJoin(t *testing.T) {
	sql, args := Table("users").
		Select("users.name", "orders.total").
		Join("orders", Eq("users.id", Raw("orders.user_id"))).
		Build()
	assertSQL(t, "SELECT users.name, orders.total FROM users INNER JOIN orders ON users.id = orders.user_id", nil, sql, args)
}

func TestLeftJoin(t *testing.T) {
	sql, args := Table("users").
		LeftJoin("orders", Eq("users.id", Raw("orders.user_id"))).
		Build()
	assertSQL(t, "SELECT * FROM users LEFT JOIN orders ON users.id = orders.user_id", nil, sql, args)
}

func TestIn(t *testing.T) {
	sql, args := Table("users").
		Where(In("id", 1, 2, 3)).
		Build()
	assertSQL(t, "SELECT * FROM users WHERE (id IN (?, ?, ?))", []any{1, 2, 3}, sql, args)
}

func TestNotIn(t *testing.T) {
	sql, args := Table("users").
		Where(NotIn("id", 1, 2)).
		Build()
	assertSQL(t, "SELECT * FROM users WHERE (id NOT IN (?, ?))", []any{1, 2}, sql, args)
}

func TestBetween(t *testing.T) {
	sql, args := Table("users").
		Where(Between("age", 18, 60)).
		Build()
	assertSQL(t, "SELECT * FROM users WHERE (age BETWEEN ? AND ?)", []any{18, 60}, sql, args)
}

func TestIsNull(t *testing.T) {
	sql, args := Table("users").
		Where(IsNull("deleted_at")).
		Build()
	assertSQL(t, "SELECT * FROM users WHERE (deleted_at IS NULL)", nil, sql, args)
}

func TestIsNotNull(t *testing.T) {
	sql, args := Table("users").
		Where(IsNotNull("deleted_at")).
		Build()
	assertSQL(t, "SELECT * FROM users WHERE (deleted_at IS NOT NULL)", nil, sql, args)
}

func TestLike(t *testing.T) {
	sql, args := Table("users").
		Where(Like("name", "%Alice%")).
		Build()
	assertSQL(t, "SELECT * FROM users WHERE (name LIKE ?)", []any{"%Alice%"}, sql, args)
}

func TestAndOr(t *testing.T) {
	sql, args := Table("users").
		Where(
			And(Eq("status", 1), Gt("age", 18)),
			Or(Eq("role", "admin"), Eq("role", "manager")),
		).
		Build()
	assertSQL(t, "SELECT * FROM users WHERE ((status = ? AND age > ?) AND (role = ? OR role = ?))",
		[]any{1, 18, "admin", "manager"}, sql, args)
}

func TestNot(t *testing.T) {
	sql, args := Table("users").
		Where(Not(Eq("status", 0))).
		Build()
	assertSQL(t, "SELECT * FROM users WHERE (NOT (status = ?))", []any{0}, sql, args)
}

func TestExists(t *testing.T) {
	subQuery := Table("orders").Select("1").Where(Eq("user_id", Raw("users.id")))
	sql, args := Table("users").
		Where(Exists(subQuery)).
		Build()
	assertSQL(t, "SELECT * FROM users WHERE (EXISTS (SELECT 1 FROM orders WHERE (user_id = users.id)))", nil, sql, args)
}

func TestWhen(t *testing.T) {
	// When condition is true
	isAdmin := true
	sql, args := Table("users").
		Where(Eq("status", 1)).
		When(isAdmin, func(q *QueryBuilder) *QueryBuilder {
			return q.Where(Eq("role", "admin"))
		}).
		Build()
	assertSQL(t, "SELECT * FROM users WHERE (status = ? AND role = ?)", []any{1, "admin"}, sql, args)

	// When condition is false
	isAdmin = false
	sql2, args2 := Table("users").
		Where(Eq("status", 1)).
		When(isAdmin, func(q *QueryBuilder) *QueryBuilder {
			return q.Where(Eq("role", "admin"))
		}).
		Build()
	assertSQL(t, "SELECT * FROM users WHERE (status = ?)", []any{1}, sql2, args2)
}

func TestComment(t *testing.T) {
	sql, args := Table("users").
		Comment("This is a test query").
		Where(Eq("id", 1)).
		Build()
	assertSQL(t, "/* This is a test query */ SELECT * FROM users WHERE (id = ?)", []any{1}, sql, args)
}

func TestCrossJoin(t *testing.T) {
	sql, args := Table("users").CrossJoin("orders").Build()
	assertSQL(t, "SELECT * FROM users CROSS JOIN orders", nil, sql, args)
}

func TestInsertSet(t *testing.T) {
	// Insert().Set() 风格
	sql, args := Insert("users").
		Set("name", "Alice").
		Set("age", 18).
		Build()
	assertSQL(t, "INSERT INTO users (name, age) VALUES (?, ?)", []any{"Alice", 18}, sql, args)
}

func TestInsertColumnsValues(t *testing.T) {
	// InsertInto().Columns().Values() 风格
	sql, args := InsertInto("users").
		Columns("name", "email").
		Values("Alice", "alice@test.com").
		Values("Bob", "bob@test.com").
		Build()
	assertSQL(t, "INSERT INTO users (name, email) VALUES (?, ?), (?, ?)",
		[]any{"Alice", "alice@test.com", "Bob", "bob@test.com"}, sql, args)
}

func TestInsertSelect(t *testing.T) {
	// INSERT ... SELECT
	sql, args := InsertInto("users", "name", "email").
		SubQuery(
			Table("temp_users").Select("name", "email").Where(Eq("status", 1)),
		).
		Build()
	assertSQL(t, "INSERT INTO users (name, email) SELECT name, email FROM temp_users WHERE (status = ?)",
		[]any{1}, sql, args)
}

func TestDeleteShort(t *testing.T) {
	sql, args := Delete("users").Where(Eq("id", 1)).Build()
	assertSQL(t, "DELETE FROM users WHERE (id = ?)", []any{1}, sql, args)
}

func TestChainExample(t *testing.T) {
	// 复杂链式调用示例
	sql, args := Table("users").
		Select("id", "name", "email").
		Where(Eq("status", 1)).
		Where(Gt("age", 18)).
		OrWhere(Eq("role", "admin")).
		OrderBy("created_at", OrderDesc).
		Limit(10).
		Build()

	fmt.Printf("SQL:  %s\n", sql)
	fmt.Printf("Args: %v\n", args)
}

func TestWhenExample(t *testing.T) {
	// 用户示例中的用法
	a := 1
	sql, args := Table("users").
		Select("id", "name").
		When(a == 1, func(q *QueryBuilder) *QueryBuilder {
			return q.Where(Eq("a", 1))
		}).
		Count()

	fmt.Printf("SQL:  %s\n", sql)
	fmt.Printf("Args: %v\n", args)
	assertSQL(t, "SELECT COUNT(*) FROM users WHERE (a = ?)", []any{1}, sql, args)
}

// ==================== SetStruct 测试 ====================

type testUserModel struct {
	Id     ModelField[int64]  `db:"id"`
	Name   ModelField[string] `db:"name"`
	Age    ModelField[int]    `db:"age"`
	Email  ModelField[string] `db:"email"`
	Status ModelField[int]    `db:"status"`
}

func TestSetStructInsert(t *testing.T) {
	user := NewModel[testUserModel]()
	user.Name.Set("Alice")
	user.Age.Set(18)
	user.Email.Set("alice@test.com")
	user.Status.Set(1)

	sql, args := Insert("users").SetStruct(user).Build()
	assertSQL(t, "INSERT INTO users (id, name, age, email, status) VALUES (?, ?, ?, ?, ?)",
		[]any{int64(0), "Alice", 18, "alice@test.com", 1}, sql, args)
}

func TestSetStructUpdate(t *testing.T) {
	user := NewModel[testUserModel]()
	user.Name.Set("Bob")
	user.Age.Set(25)
	user.Status.Set(1)

	sql, args := Update("users").SetStruct(user).Where(Eq("id", 1)).Build()
	assertSQL(t, "UPDATE users SET id = ?, name = ?, age = ?, email = ?, status = ? WHERE (id = ?)",
		[]any{int64(0), "Bob", 25, "", 1, 1}, sql, args)
}

func TestSetStructWithSet(t *testing.T) {
	// SetStruct 和 Set 混合使用
	user := NewModel[testUserModel]()
	user.Name.Set("Charlie")

	sql, args := Update("users").
		SetStruct(user).
		Set("updated_at", "2024-01-01").
		Where(Eq("id", 1)).
		Build()
	assertSQL(t, "UPDATE users SET id = ?, name = ?, age = ?, email = ?, status = ?, updated_at = ? WHERE (id = ?)",
		[]any{int64(0), "Charlie", 0, "", 0, "2024-01-01", 1}, sql, args)
}

func TestSetStructPtr(t *testing.T) {
	// 传入指针
	user := NewModel[testUserModel]()
	user.Name.Set("Dave")

	sql, args := Insert("users").SetStruct(user).Build()
	assertSQL(t, "INSERT INTO users (id, name, age, email, status) VALUES (?, ?, ?, ?, ?)",
		[]any{int64(0), "Dave", 0, "", 0}, sql, args)
}

func TestSetStructAllFields(t *testing.T) {
	// SetStruct 会设置所有 ModelField 字段（包括零值），这是预期行为
	user := NewModel[testUserModel]()
	user.Name.Set("Eve")

	sql, args := Insert("users").SetStruct(user).Build()
	assertSQL(t, "INSERT INTO users (id, name, age, email, status) VALUES (?, ?, ?, ?, ?)",
		[]any{int64(0), "Eve", 0, "", 0}, sql, args)
}
