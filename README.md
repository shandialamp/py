# py — Go SQL 语句生成器

> 一个轻量级的 Go SQL 语句生成器，支持链式调用和常见 SQL 语法。**只生成语句，不执行查询。**

## 安装

```bash
go get github.com/shandialamp/py
```

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/shandialamp/py"
)

func main() {
    sql, args := py.Table("users").
        Select("id", "name", "email").
        Where(py.Eq("status", 1), py.Gt("age", 18)).
        OrderBy("created_at", py.OrderDesc).
        Limit(10).
        Build()

    fmt.Println(sql)  // SELECT id, name, email FROM users WHERE (status = ? AND age > ?) ORDER BY created_at DESC LIMIT 10
    fmt.Println(args) // [1 18]
}
```

## 核心概念

### 链式调用

所有方法返回 `*QueryBuilder`，支持无限链式调用：

```go
py.Table("users").
    Select("id", "name").
    Where(py.Eq("status", 1)).
    OrderBy("created_at", py.OrderDesc).
    Limit(10).
    Build()
```

### 参数占位符

生成使用 `?` 占位符的 SQL，参数值以 `[]any` 形式返回，可直接用于 `database/sql`：

```go
sql, args := py.Table("users").Where(py.Eq("id", 1)).Build()
db.Query(sql, args...) // 直接使用
```

## 条件表达式

### 比较运算符

| 函数 | SQL 运算符 | 示例 |
|------|-----------|------|
| `Eq` | `=` | `py.Eq("name", "Alice")` → `name = ?` |
| `Neq` | `!=` | `py.Neq("status", 0)` → `status != ?` |
| `Gt` | `>` | `py.Gt("age", 18)` → `age > ?` |
| `Gte` | `>=` | `py.Gte("age", 18)` → `age >= ?` |
| `Lt` | `<` | `py.Lt("price", 100)` → `price < ?` |
| `Lte` | `<=` | `py.Lte("price", 100)` → `price <= ?` |

### 模糊匹配

```go
py.Like("name", "%john%")     // name LIKE ?
py.NotLike("name", "%test%")  // name NOT LIKE ?
```

### 范围查询

```go
py.In("id", 1, 2, 3)              // id IN (?, ?, ?)
py.NotIn("id", 1, 2)              // id NOT IN (?, ?)
py.Between("age", 18, 60)         // age BETWEEN ? AND ?
py.NotBetween("age", 0, 17)       // age NOT BETWEEN ? AND ?
```

### NULL 判断

```go
py.IsNull("deleted_at")       // deleted_at IS NULL
py.IsNotNull("deleted_at")    // deleted_at IS NOT NULL
```

### 子查询

```go
sub := py.Table("orders").Select("1").Where(py.Eq("user_id", py.Raw("users.id")))
py.Exists(sub)     // EXISTS (SELECT 1 FROM orders WHERE user_id = users.id)
py.NotExists(sub)  // NOT EXISTS (...)
```

### 逻辑组合

```go
py.And(py.Eq("status", 1), py.Gt("age", 18))  // (status = ? AND age > ?)
py.Or(py.Eq("role", "admin"), py.Eq("role", "manager"))  // (role = ? OR role = ?)
py.Not(py.Eq("status", 0))  // NOT (status = ?)
```

### 原始 SQL

```go
py.Raw("DATE(created_at) = CURDATE()")              // 无参数
py.Raw("FIND_IN_SET(?, tags)", "hot")               // 带参数
```

## 查询构建

### SELECT

```go
// 查询所有列
py.Table("users").Build()
// SELECT * FROM users

// 指定列
py.Table("users").Select("id", "name", "email").Build()
// SELECT id, name, email FROM users

// DISTINCT
py.Table("users").Select("city").Distinct().Build()
// SELECT DISTINCT city FROM users
```

### JOIN

```go
// INNER JOIN
py.Table("users").
    Join("orders", py.Eq("users.id", py.Raw("orders.user_id"))).
    Build()
// SELECT * FROM users INNER JOIN orders ON users.id = orders.user_id

// LEFT JOIN
py.Table("users").
    LeftJoin("orders", py.Eq("users.id", py.Raw("orders.user_id"))).
    Build()

// RIGHT JOIN / CROSS JOIN 同理
```

### WHERE

```go
// AND 连接多个条件
py.Table("users").
    Where(py.Eq("status", 1), py.Gt("age", 18)).
    Build()
// SELECT * FROM users WHERE (status = ? AND age > ?)

// OR 连接
py.Table("users").
    Where(py.Eq("status", 1)).
    OrWhere(py.Eq("role", "admin"), py.Eq("role", "manager")).
    Build()
// SELECT * FROM users WHERE (status = ? AND (role = ? OR role = ?))
```

### GROUP BY / HAVING

```go
py.Table("orders").
    Select("user_id", "COUNT(*) as total").
    GroupBy("user_id").
    Having(py.Gt("total", 5)).
    Build()
// SELECT user_id, COUNT(*) as total FROM orders GROUP BY user_id HAVING (total > ?)
```

### ORDER BY / LIMIT / OFFSET

```go
py.Table("users").
    OrderBy("created_at", py.OrderDesc).
    OrderBy("id").
    Limit(10).
    Offset(20).
    Build()
// SELECT * FROM users ORDER BY created_at DESC, id ASC LIMIT 10 OFFSET 20
```

### 聚合函数

```go
py.Table("users").Where(py.Eq("status", 1)).Count()
// SELECT COUNT(*) FROM users WHERE (status = ?)

py.Table("orders").Where(py.Eq("user_id", 42)).Sum("amount")
// SELECT SUM(amount) FROM orders WHERE (user_id = ?)

py.Table("orders").Avg("price")   // AVG
py.Table("orders").Max("price")   // MAX
py.Table("orders").Min("price")   // MIN
```

### 条件分支 When

```go
isAdmin := true
keyword := "alice"

py.Table("users").
    Select("id", "name").
    Where(py.Eq("status", 1)).
    When(isAdmin, func(q *py.QueryBuilder) *py.QueryBuilder {
        return q.Where(py.Eq("role", "admin"))
    }).
    When(keyword != "", func(q *py.QueryBuilder) *py.QueryBuilder {
        return q.Where(py.Like("name", "%"+keyword+"%"))
    }).
    Build()
// SELECT id, name FROM users WHERE (status = ? AND role = ? AND name LIKE ?)
// [1 admin %alice%]
```

### SQL 注释

```go
py.Table("users").
    Comment("Get active users").
    Where(py.Eq("status", 1)).
    Build()
// /* Get active users */ SELECT * FROM users WHERE (status = ?)
```

## INSERT / UPDATE / DELETE

### INSERT

```go
py.InsertInto("users", "name", "email", "age").
    Values("Alice", "alice@example.com", 18).
    Values("Bob", "bob@example.com", 20).
    Build()
// INSERT INTO users (name, email, age) VALUES (?, ?, ?), (?, ?, ?)
// [Alice alice@example.com 18 Bob bob@example.com 20]
```

### UPDATE

```go
py.Update("users").
    Set("name", "Alice").
    Set("age", 20).
    Where(py.Eq("id", 1)).
    Build()
// UPDATE users SET name = ?, age = ? WHERE (id = ?)
// [Alice 20 1]
```

### DELETE

```go
py.DeleteFrom("users").Where(py.Eq("id", 1)).Build()
// DELETE FROM users WHERE (id = ?)
// [1]
```

## 综合示例

```go
status := 1
keyword := "john"

sql, args := py.Table("users u").
    Select("u.id", "u.name", "u.email", "COUNT(o.id) as order_count").
    LeftJoin("orders o", py.Eq("u.id", py.Raw("o.user_id"))).
    Where(py.Eq("u.status", status)).
    When(keyword != "", func(q *py.QueryBuilder) *py.QueryBuilder {
        return q.Where(py.Or(
            py.Like("u.name", "%"+keyword+"%"),
            py.Like("u.email", "%"+keyword+"%"),
        ))
    }).
    GroupBy("u.id").
    Having(py.Gt("order_count", 0)).
    OrderBy("order_count", py.OrderDesc).
    Limit(20).
    Offset(0).
    Comment("Get active users with orders").
    Build()
```

## API 速查

### 入口函数

| 函数 | 说明 |
|------|------|
| `Table(name)` | SELECT 查询入口 |
| `InsertInto(table, cols...)` | INSERT 入口 |
| `Update(table)` | UPDATE 入口 |
| `DeleteFrom(table)` | DELETE 入口 |

### QueryBuilder 方法

| 方法 | 说明 |
|------|------|
| `Select(cols...)` | 指定查询列 |
| `From(table)` | 设置表名 |
| `Distinct()` | 添加 DISTINCT |
| `Where(exprs...)` | AND 连接 WHERE 条件 |
| `OrWhere(exprs...)` | OR 连接 WHERE 条件 |
| `When(cond, fn)` | 条件分支 |
| `Join / LeftJoin / RightJoin / CrossJoin` | JOIN 子句 |
| `GroupBy(cols...)` | GROUP BY |
| `Having(exprs...)` | HAVING 条件 |
| `OrderBy(field, dir...)` | ORDER BY |
| `Limit(n)` | LIMIT |
| `Offset(n)` | OFFSET |
| `Set(col, val)` | UPDATE SET |
| `Values(vals...)` | INSERT VALUES |
| `Comment(text)` | SQL 注释 |
| `Count / Sum / Avg / Max / Min` | 聚合函数 |
| `Build()` | 生成 SQL 和参数 |

## 许可证

MIT
