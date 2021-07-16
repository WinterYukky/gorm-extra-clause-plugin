# gorm-extra-clause-plugin

The clause support plugin for gorm, that not supported by gorm.

## Support clauses

- [x] WITH (CTE)

## Install
```shell
go get github.com/WinterYukky/gorm-extra-clause-plugin
```

## Get Started

```go
package main

import (
  "gorm.io/gorm"
  "gorm.io/driver/sqlite"
  // Add plugin package
  extraClausePlugin "github.com/WinterYukky/gorm-extra-clause-plugin"
  "github.com/WinterYukky/gorm-extra-clause-plugin/exclause"
)

func main() {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    // Insert this line
    db.Use(extraClausePlugin.New())
    // Use exclauses
    db.Clauses(exclause.NewWith(db, "cte", db.Table("users"))).Table("cte").Scan(&users)
}
```

## Examples

### WITH (CTE)

```go
// WITH `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`
db.Clauses(exclause.NewWith(db, "cte", "SELECT * FROM `users`")).Table("cte").Scan(&users)

// WITH `cte` AS (SELECT * FROM `users` WHERE `name` = 'WinterYukky') SELECT * FROM `cte`
db.Clauses(exclause.NewWith(db, "cte", "SELECT * FROM `users` WHERE `name` = ?", "WinterYukky")).Table("cte").Scan(&users)

// WITH `cte` AS (SELECT * FROM `users` WHERE `name` = 'WinterYukky') SELECT * FROM `cte`
db.Clauses(exclause.NewWith(db, "cte", db.Table("users").Where("`name` = ?", "WinterYukky"))).Table("cte").Scan(&users)

// WITH `cte`(`id`,`name`) AS (SELECT * FROM `users`) SELECT * FROM `cte`
db.Clauses(exclause.NewWith(db, exclause.CTE{Alias: "cte", Columns: []string{"id", "name"}}, db.Table("users"))).Table("cte").Scan(&users)

// WITH RECURSIVE `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`
db.Clauses(exclause.NewWith(db, exclause.CTE{Recursive: true, Alias: "cte"}, db.Table("users"))).Table("cte").Scan(&users)
```