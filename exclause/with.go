package exclause

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CTEMaterializeOption represents the materialization hint for a CTE
type CTEMaterializeOption int

const (
	// CTEMaterializeUnspecified means no materialization hint (database default)
	CTEMaterializeUnspecified CTEMaterializeOption = iota
	// CTEMaterialize forces the CTE to be materialized
	CTEMaterialize
	// CTENotMaterialize prevents the CTE from being materialized
	CTENotMaterialize
)

// With with clause
//
//	// examples
//	// WITH `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`
//	db.Clauses(exclause.With{CTEs: []exclause.CTE{{Name: "cte", Subquery: clause.Expr{SQL: "SELECT * FROM `users`"}}}}).Table("cte").Scan(&users)
//
//	// WITH `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`
//	db.Clauses(exclause.With{CTEs: []exclause.CTE{{Name: "cte", Subquery: exclause.Subquery{DB: db.Table("users")}}}}).Table("cte").Scan(&users)
//
//	// WITH `cte` (`id`,`name`) AS (SELECT * FROM `users`) SELECT * FROM `cte`
//	db.Clauses(exclause.With{CTEs: []exclause.CTE{{Name: "cte", Columns: []string{"id", "name"}, Subquery: exclause.Subquery{DB: db.Table("users")}}}}).Table("cte").Scan(&users)
//
//	// WITH RECURSIVE `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`
//	db.Clauses(exclause.With{Recursive: true, CTEs: []exclause.CTE{{Name: "cte", Subquery: exclause.Subquery{DB: db.Table("users")}}}}).Table("cte").Scan(&users)
//
//	// WITH `cte` AS MATERIALIZED (SELECT * FROM `users`) SELECT * FROM `cte`
//	db.Clauses(exclause.With{CTEs: []exclause.CTE{{Name: "cte", Subquery: exclause.Subquery{DB: db.Table("users")}, Materialized: exclause.CTEMaterialize}}}).Table("cte").Scan(&users)
//
//	// WITH `cte` AS NOT MATERIALIZED (SELECT * FROM `users`) SELECT * FROM `cte`
//	db.Clauses(exclause.With{CTEs: []exclause.CTE{{Name: "cte", Subquery: exclause.Subquery{DB: db.Table("users")}, Materialized: exclause.CTENotMaterialize}}}).Table("cte").Scan(&users)
//
//	// WITH `cte1` AS MATERIALIZED (...), `cte2` AS NOT MATERIALIZED (...) SELECT * FROM `cte1`
//	db.Clauses(exclause.With{CTEs: []exclause.CTE{
//		exclause.NewMaterializedCTE("cte1", exclause.Subquery{DB: db.Table("users")}),
//		exclause.NewNotMaterializedCTE("cte2", exclause.Subquery{DB: db.Table("products")}),
//	}}).Table("cte1").Scan(&users)
type With struct {
	Recursive bool
	CTEs      []CTE
}

// CTE common table expressions
type CTE struct {
	Name         string
	Columns      []string
	Subquery     clause.Expression
	Materialized CTEMaterializeOption
}

// NewCTE creates a new CTE with unspecified materialization (database default)
func NewCTE(name string, subquery clause.Expression) CTE {
	return CTE{
		Name:         name,
		Subquery:     subquery,
		Materialized: CTEMaterializeUnspecified,
	}
}

// NewMaterializedCTE creates a new CTE that will be materialized
func NewMaterializedCTE(name string, subquery clause.Expression) CTE {
	return CTE{
		Name:         name,
		Subquery:     subquery,
		Materialized: CTEMaterialize,
	}
}

// NewNotMaterializedCTE creates a new CTE that will not be materialized
func NewNotMaterializedCTE(name string, subquery clause.Expression) CTE {
	return CTE{
		Name:         name,
		Subquery:     subquery,
		Materialized: CTENotMaterialize,
	}
}

// Name with clause name
func (with With) Name() string {
	return "WITH"
}

// Build build with clause
func (with With) Build(builder clause.Builder) {
	if with.Recursive {
		builder.WriteString("RECURSIVE ")
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
	builder.WriteQuoted(cte.Name)
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

	switch cte.Materialized {
	case CTEMaterialize:
		builder.WriteString("MATERIALIZED ")
	case CTENotMaterialize:
		builder.WriteString("NOT MATERIALIZED ")
	}

	builder.WriteByte('(')
	cte.Subquery.Build(builder)
	builder.WriteByte(')')
}

// MergeClause merge With clauses
func (with With) MergeClause(clause *clause.Clause) {
	if w, ok := clause.Expression.(With); ok {
		if w.Recursive {
			with.Recursive = true
		}
		ctes := make([]CTE, len(w.CTEs)+len(with.CTEs))
		copy(ctes, w.CTEs)
		copy(ctes[len(w.CTEs):], with.CTEs)
		with.CTEs = ctes
	}

	clause.Expression = with
}

// NewWith is easy to create new With
//
//	// examples
//	// WITH `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`
//	db.Clauses(exclause.NewWith("cte", "SELECT * FROM `users`")).Table("cte").Scan(&users)
//
//	// WITH `cte` AS (SELECT * FROM `users` WHERE `name` = 'WinterYukky') SELECT * FROM `cte`
//	db.Clauses(exclause.NewWith("cte", "SELECT * FROM `users` WHERE `name` = ?", "WinterYukky")).Table("cte").Scan(&users)
//
//	// WITH `cte` AS (SELECT * FROM `users` WHERE `name` = 'WinterYukky') SELECT * FROM `cte`
//	db.Clauses(exclause.NewWith("cte", db.Table("users").Where("`name` = ?", "WinterYukky"))).Table("cte").Scan(&users)
//
// If you need more advanced WITH clause, you can see With struct.
func NewWith(name string, subquery interface{}, args ...interface{}) With {
	switch v := subquery.(type) {
	case *gorm.DB:
		return With{
			CTEs: []CTE{
				{
					Name:     name,
					Subquery: Subquery{DB: v},
				},
			},
		}
	case string:
		return With{
			CTEs: []CTE{
				{
					Name:     name,
					Subquery: clause.Expr{SQL: v, Vars: args},
				},
			},
		}
	}
	return With{}
}
