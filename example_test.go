package py

import (
	"fmt"
)

// Example 1: 基础查询 - 查询所有用户
func ExampleTable_basic() {
	sql, args := Table("users").Select("id", "name", "email").Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name, email FROM users
	// []
}

// Example 2: 带条件的查询
func ExampleTable_where() {
	sql, args := Table("users").
		Select("id", "name").
		Where(Eq("status", 1), Gt("age", 18)).
		OrderBy("created_at", OrderDesc).
		Limit(10).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name FROM users WHERE (status = ? AND age > ?) ORDER BY created_at DESC LIMIT 10
	// [1 18]
}

// Example 3: 聚合查询 - COUNT
func ExampleTable_count() {
	sql, args := Table("users").Where(Eq("status", 1)).Count()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT COUNT(*) FROM users WHERE (status = ?)
	// [1]
}

// Example 4: 聚合查询 - SUM/AVG
func ExampleTable_sum() {
	sql, args := Table("orders").Where(Eq("user_id", 42)).Sum("amount")
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT SUM(amount) FROM orders WHERE (user_id = ?)
	// [42]
}

// Example 5: 条件分支 When
func ExampleTable_when() {
	isAdmin := true
	keyword := "alice"

	sql, args := Table("users").
		Select("id", "name").
		Where(Eq("status", 1)).
		When(isAdmin, func(q *QueryBuilder) *QueryBuilder {
			return q.Where(Eq("role", "admin"))
		}).
		When(keyword != "", func(q *QueryBuilder) *QueryBuilder {
			return q.Where(Like("name", fmt.Sprintf("%%%s%%", keyword)))
		}).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name FROM users WHERE (status = ? AND role = ? AND name LIKE ?)
	// [1 admin %alice%]
}

// Example 6: 条件分支 When - 条件为 false 时跳过
func ExampleTable_when_skip() {
	isAdmin := false

	sql, args := Table("users").
		Select("id", "name").
		Where(Eq("status", 1)).
		When(isAdmin, func(q *QueryBuilder) *QueryBuilder {
			return q.Where(Eq("role", "admin"))
		}).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name FROM users WHERE (status = ?)
	// [1]
}

// Example 7: IN 查询
func ExampleTable_in() {
	sql, args := Table("users").
		Select("id", "name").
		Where(In("id", 1, 2, 3, 4, 5)).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name FROM users WHERE (id IN (?, ?, ?, ?, ?))
	// [1 2 3 4 5]
}

// Example 8: 模糊查询 LIKE
func ExampleTable_like() {
	sql, args := Table("products").
		Select("id", "name").
		Where(Like("name", "%phone%")).
		Where(Gt("price", 100)).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name FROM products WHERE (name LIKE ? AND price > ?)
	// [%phone% 100]
}

// Example 9: 范围查询 BETWEEN
func ExampleTable_between() {
	sql, args := Table("orders").
		Select("id", "total", "created_at").
		Where(Between("created_at", "2024-01-01", "2024-12-31")).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, total, created_at FROM orders WHERE (created_at BETWEEN ? AND ?)
	// [2024-01-01 2024-12-31]
}

// Example 10: NULL 查询
func ExampleTable_null() {
	sql, args := Table("users").
		Select("id", "name").
		Where(IsNull("deleted_at")).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name FROM users WHERE (deleted_at IS NULL)
	// []
}

// Example 11: 逻辑组合 And/Or
func ExampleTable_and_or() {
	sql, args := Table("users").
		Select("id", "name").
		Where(
			And(Eq("status", 1), Gt("age", 18)),
			Or(Eq("role", "admin"), Eq("role", "manager")),
		).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name FROM users WHERE ((status = ? AND age > ?) AND (role = ? OR role = ?))
	// [1 18 admin manager]
}

// Example 12: NOT 取反
func ExampleTable_not() {
	sql, args := Table("users").
		Select("id", "name").
		Where(Not(Eq("status", 0))).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name FROM users WHERE (NOT (status = ?))
	// [0]
}

// Example 13: JOIN 查询
func ExampleTable_join() {
	sql, args := Table("users").
		Select("users.name", "orders.total").
		Join("orders", Eq("users.id", Raw("orders.user_id"))).
		Where(Eq("users.status", 1)).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT users.name, orders.total FROM users INNER JOIN orders ON users.id = orders.user_id WHERE (users.status = ?)
	// [1]
}

// Example 14: LEFT JOIN 查询
func ExampleTable_left_join() {
	sql, args := Table("users").
		Select("users.name", "COUNT(orders.id) as order_count").
		LeftJoin("orders", Eq("users.id", Raw("orders.user_id"))).
		GroupBy("users.id").
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT users.name, COUNT(orders.id) as order_count FROM users LEFT JOIN orders ON users.id = orders.user_id GROUP BY users.id
	// []
}

// Example 15: 子查询 EXISTS
func ExampleTable_exists() {
	subQuery := Table("orders").Select("1").Where(Eq("user_id", Raw("users.id")))
	sql, args := Table("users").
		Select("id", "name").
		Where(Exists(subQuery)).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name FROM users WHERE (EXISTS (SELECT 1 FROM orders WHERE (user_id = users.id)))
	// []
}

// Example 16: GROUP BY + HAVING
func ExampleTable_group_having() {
	sql, args := Table("orders").
		Select("user_id", "COUNT(*) as total", "SUM(amount) as sum").
		Where(Eq("status", 1)).
		GroupBy("user_id").
		Having(Gt("total", 5), Lt("sum", 1000)).
		OrderBy("total", OrderDesc).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT user_id, COUNT(*) as total, SUM(amount) as sum FROM orders WHERE (status = ?) GROUP BY user_id HAVING (total > ? AND sum < ?) ORDER BY total DESC
	// [1 5 1000]
}

// Example 17: DISTINCT 查询
func ExampleTable_distinct() {
	sql, args := Table("users").Select("city").Distinct().Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT DISTINCT city FROM users
	// []
}

// Example 18: 分页查询
func ExampleTable_paginate() {
	page := 2
	pageSize := 20

	sql, args := Table("articles").
		Select("id", "title", "created_at").
		Where(Eq("status", 1)).
		OrderBy("created_at", OrderDesc).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, title, created_at FROM articles WHERE (status = ?) ORDER BY created_at DESC LIMIT 20 OFFSET 20
	// [1]
}

// Example 19: INSERT 语句
func ExampleInsertInto() {
	sql, args := InsertInto("users", "name", "email", "age").
		Values("Alice", "alice@example.com", 18).
		Values("Bob", "bob@example.com", 20).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// INSERT INTO users (name, email, age) VALUES (?, ?, ?), (?, ?, ?)
	// [Alice alice@example.com 18 Bob bob@example.com 20]
}

// Example 20: UPDATE 语句
func ExampleUpdate() {
	sql, args := Update("users").
		Set("name", "Alice").
		Set("age", 20).
		Where(Eq("id", 1)).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// UPDATE users SET name = ?, age = ? WHERE (id = ?)
	// [Alice 20 1]
}

// Example 21: DELETE 语句
func ExampleDeleteFrom() {
	sql, args := DeleteFrom("users").Where(Eq("id", 1)).Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// DELETE FROM users WHERE (id = ?)
	// [1]
}

// Example 22: 复杂链式调用 - 综合示例
func ExampleTable_complex() {
	status := 1
	keyword := "john"
	minAge := 18

	sql, args := Table("users").
		Select("u.id", "u.name", "u.email", "COUNT(o.id) as order_count").
		From("users u").
		LeftJoin("orders o", Eq("u.id", Raw("o.user_id"))).
		Where(Eq("u.status", status)).
		When(keyword != "", func(q *QueryBuilder) *QueryBuilder {
			return q.Where(Or(
				Like("u.name", fmt.Sprintf("%%%s%%", keyword)),
				Like("u.email", fmt.Sprintf("%%%s%%", keyword)),
			))
		}).
		When(minAge > 0, func(q *QueryBuilder) *QueryBuilder {
			return q.Where(Gte("u.age", minAge))
		}).
		GroupBy("u.id").
		Having(Gt("order_count", 0)).
		OrderBy("order_count", OrderDesc).
		OrderBy("u.name").
		Limit(20).
		Offset(0).
		Comment("Get active users with orders").
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// /* Get active users with orders */ SELECT u.id, u.name, u.email, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON u.id = o.user_id WHERE (u.status = ? AND (u.name LIKE ? OR u.email LIKE ?) AND u.age >= ?) GROUP BY u.id HAVING (order_count > ?) ORDER BY order_count DESC, u.name ASC LIMIT 20 OFFSET 0
	// [1 %john% %john% 18 0]
}

// Example 23: 使用 Raw 表达式
func ExampleTable_raw() {
	sql, args := Table("users").
		Select("id", "name").
		Where(Raw("DATE(created_at) = CURDATE()")).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name FROM users WHERE (DATE(created_at) = CURDATE())
	// []
}

// Example 24: NotIn / NotBetween / IsNotNull
func ExampleTable_negations() {
	sql, args := Table("users").
		Select("id", "name").
		Where(
			NotIn("id", 1, 2, 3),
			NotBetween("age", 0, 17),
			IsNotNull("email"),
		).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name FROM users WHERE (id NOT IN (?, ?, ?) AND age NOT BETWEEN ? AND ? AND email IS NOT NULL)
	// [1 2 3 0 17]
}

// Example 25: 多表 JOIN 链式调用
func ExampleTable_multi_join() {
	sql, args := Table("users u").
		Select("u.name", "o.id as order_id", "p.name as product_name").
		Join("orders o", Eq("u.id", Raw("o.user_id"))).
		Join("order_items oi", Eq("o.id", Raw("oi.order_id"))).
		Join("products p", Eq("oi.product_id", Raw("p.id"))).
		Where(Eq("u.status", 1)).
		Limit(10).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT u.name, o.id as order_id, p.name as product_name FROM users u INNER JOIN orders o ON u.id = o.user_id INNER JOIN order_items oi ON o.id = oi.order_id INNER JOIN products p ON oi.product_id = p.id WHERE (u.status = ?) LIMIT 10
	// [1]
}
