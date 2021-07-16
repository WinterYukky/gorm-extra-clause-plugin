package exclause

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// With with clause
type With struct {
	CTEs []CTE
}

// CTE common table expressions
type CTE struct {
	Recursive   bool
	Alias       string
	Columns     []string
	Expressions []clause.Expression
}

// Name with clause name
func (with With) Name() string {
	return "WITH"
}

// Build build with clause
func (with With) Build(builder clause.Builder) {
	for _, cte := range with.CTEs {
		if cte.Recursive {
			builder.WriteString("RECURSIVE ")
			break
		}
	}
	for index, cte := range with.CTEs {
		if index > 0 {
			builder.WriteByte(',')
		}
		cte.Build(builder)
	}
}

// Build build CTE
func (cte CTE) Build(builder clause.Builder) {
	builder.WriteQuoted(cte.Alias)
	if len(cte.Columns) > 0 {
		builder.WriteString(" (")
		for index, column := range cte.Columns {
			if index > 0 {
				builder.WriteByte(',')
			}
			builder.WriteQuoted(column)
		}
		builder.WriteByte(')')
	}

	builder.WriteString(" AS ")

	builder.WriteByte('(')
	for _, expression := range cte.Expressions {
		expression.Build(builder)
	}
	builder.WriteByte(')')
}

// MergeClause merge With clauses
func (with With) MergeClause(clause *clause.Clause) {
	if w, ok := clause.Expression.(With); ok {
		ctes := make([]CTE, len(w.CTEs)+len(with.CTEs))
		copy(ctes, w.CTEs)
		copy(ctes[len(w.CTEs):], with.CTEs)
		with.CTEs = ctes
	}

	clause.Expression = with
}

// NewWith is create new With
//
//  // examples
//  // WITH `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`
//  db.Clauses(exclause.NewWith(db, "cte", "SELECT * FROM `users`")).Table("cte").Scan(&users)
//
//  // WITH `cte` AS (SELECT * FROM `users` WHERE `name` = 'WinterYukky') SELECT * FROM `cte`
//  db.Clauses(exclause.NewWith(db, "cte", "SELECT * FROM `users` WHERE `name` = ?", "WinterYukky")).Table("cte").Scan(&users)
//
//  // WITH `cte` AS (SELECT * FROM `users` WHERE `name` = 'WinterYukky') SELECT * FROM `cte`
//  db.Clauses(exclause.NewWith(db, "cte", db.Table("users").Where("`name` = ?", "WinterYukky"))).Table("cte").Scan(&users)
//
//  // WITH `cte`(`id`,`name`) AS (SELECT * FROM `users`) SELECT * FROM `cte`
//  db.Clauses(exclause.NewWith(db, exclause.CTE{Alias: "cte", Columns: []string{"id", "name"}}, db.Table("users"))).Table("cte").Scan(&users)
//
//  // WITH RECURSIVE `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`
//  db.Clauses(exclause.NewWith(db, exclause.CTE{Recursive: true, Alias: "cte"}, db.Table("users"))).Table("cte").Scan(&users)
func NewWith(db *gorm.DB, alias interface{}, query interface{}, args ...interface{}) With {
	cte := CTE{}
	switch v := alias.(type) {
	case string:
		cte.Alias = v
	case CTE:
		cte = v
	default:
		return With{}
	}

	switch v := query.(type) {
	case *gorm.DB:
		if conds := db.Statement.BuildCondition("?", v); len(conds) > 0 {
			cte.Expressions = conds
			return With{CTEs: []CTE{cte}}
		}
	default:
		if conds := db.Statement.BuildCondition(query, args...); len(conds) > 0 {
			cte.Expressions = conds
			return With{CTEs: []CTE{cte}}
		}
	}
	return With{}
}
