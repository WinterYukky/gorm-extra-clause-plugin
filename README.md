# gorm-extra-clause-plugin

The clause support plugin for gorm, that not supported by gorm.

[![test status](https://github.com/WinterYukky/gorm-extra-clause-plugin/actions/workflows/go.yml/badge.svg?branch=main "test status")](https://github.com/WinterYukky/gorm-extra-clause-plugin/actions)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)

## Support clauses

- [x] WITH (CTE)
- [x] UNION
- [x] INTERSECT
- [x] EXCEPT

## Install
```shell
go get github.com/mamazinho/gorm-extra-clause-plugin
```

## Get Started

> [!IMPORTANT]
> Do not forget the plugin initialisation code `db.Use(extraClausePlugin.New()`.

```go
package main

import (
  "gorm.io/gorm"
  "gorm.io/driver/sqlite"
  // Add plugin package
  extraClausePlugin "github.com/mamazinho/gorm-extra-clause-plugin"
  "github.com/mamazinho/gorm-extra-clause-plugin/exclause"
)

func main() {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    // Insert this line
    db.Use(extraClausePlugin.New())
    // Use exclauses
    db.Clauses(exclause.NewWith("cte", db.Table("users"))).Table("cte").Scan(&users)
}
```


## Examples

### WITH (CTE)

```go
// WITH `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`
db.Clauses(exclause.NewWith("cte", "SELECT * FROM `users`")).Table("cte").Scan(&users)

// WITH `cte` AS (SELECT * FROM `users` WHERE `name` = 'WinterYukky') SELECT * FROM `cte`
db.Clauses(exclause.NewWith("cte", "SELECT * FROM `users` WHERE `name` = ?", "WinterYukky")).Table("cte").Scan(&users)

// WITH `cte` AS (SELECT * FROM `users` WHERE `name` = 'WinterYukky') SELECT * FROM `cte`
db.Clauses(exclause.NewWith("cte", db.Table("users").Where("`name` = ?", "WinterYukky"))).Table("cte").Scan(&users)

// WITH `cte` (`id`,`name`) AS (SELECT * FROM `users`) SELECT * FROM `cte`
db.Clauses(exclause.With{CTEs: []exclause.CTE{{Name: "cte", Columns: []string{"id", "name"}, Subquery: exclause.Subquery{DB: db.Table("users")}}}}).Table("cte").Scan(&users)

// WITH RECURSIVE `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`
db.Clauses(exclause.With{Recursive: true, CTEs: []exclause.CTE{{Name: "cte", Subquery: exclause.Subquery{DB: db.Table("users")}}}}).Table("cte").Scan(&users)
```

### UNION

```go
// SELECT * FROM `general_users` UNION SELECT * FROM `admin_users`
db.Table("general_users").Clauses(exclause.NewUnion("SELECT * FROM `admin_users`")).Scan(&users)

// SELECT * FROM `general_users` UNION SELECT * FROM `admin_users`
db.Table("general_users").Clauses(exclause.NewUnion(db.Table("admin_users"))).Scan(&users)

// SELECT * FROM `general_users` UNION ALL SELECT * FROM `admin_users`
db.Table("general_users").Clauses(exclause.NewUnion("ALL ?", db.Table("admin_users"))).Scan(&users)
```

### INTERSECT

```go
// SELECT * FROM `general_users` INTERSECT SELECT * FROM `admin_users`
db.Table("general_users").Clauses(exclause.NewIntersect("SELECT * FROM `admin_users`")).Scan(&users)

// SELECT * FROM `general_users` INTERSECT SELECT * FROM `admin_users`
db.Table("general_users").Clauses(exclause.NewIntersect(db.Table("admin_users"))).Scan(&users)

// SELECT * FROM `general_users` INTERSECT ALL SELECT * FROM `admin_users`
db.Table("general_users").Clauses(exclause.NewIntersect("ALL ?", db.Table("admin_users"))).Scan(&users)
```

### EXCEPT

```go
// SELECT * FROM `general_users` EXCEPT SELECT * FROM `admin_users`
db.Table("general_users").Clauses(exclause.NewExcept("SELECT * FROM `admin_users`")).Scan(&users)

// SELECT * FROM `general_users` EXCEPT SELECT * FROM `admin_users`
db.Table("general_users").Clauses(exclause.NewExcept(db.Table("admin_users"))).Scan(&users)

// SELECT * FROM `general_users` EXCEPT ALL SELECT * FROM `admin_users`
db.Table("general_users").Clauses(exclause.NewExcept("ALL ?", db.Table("admin_users"))).Scan(&users)
```
