package gormextraclauseplugin

import (
	"slices"

	"gorm.io/gorm"
)

// ExtraClausePlugin support plugin that not supported clause by gorm
type ExtraClausePlugin struct{}

// Name return plugin name
func (e *ExtraClausePlugin) Name() string {
	return "ExtraClausePlugin"
}

// Initialize register BuildClauses
func (e *ExtraClausePlugin) Initialize(db *gorm.DB) error {

	db.Callback().Query().Clauses = merge(db.Callback().Query().Clauses, queryClauses)
	db.Callback().Row().Clauses = merge(db.Callback().Row().Clauses, queryClauses)
	db.Callback().Update().Clauses = merge(db.Callback().Update().Clauses, updateClauses)
	return nil
}

// New create new ExtraClausePlugin
//
//	// example
//	db.Use(extraClausePlugin.New())
func New() *ExtraClausePlugin {
	return &ExtraClausePlugin{}
}

type pluginClause struct {
	name   string
	before string
}

var (
	queryClauses = []pluginClause{
		{name: "WITH", before: "SELECT"},
		{name: "UNION", before: "ORDER BY"},
		{name: "INTERSECT", before: "ORDER BY"},
		{name: "EXCEPT", before: "ORDER BY"},
	}
	updateClauses = []pluginClause{
		{name: "WITH", before: "UPDATE"},
		{name: "UNION", before: "ORDER BY"},
		{name: "INTERSECT", before: "ORDER BY"},
		{name: "EXCEPT", before: "ORDER BY"},
	}
)

func merge(origin []string, pluginClauses []pluginClause) []string {
	collect := func(target string) []string {
		found := []string{}
		for _, clause := range pluginClauses {
			if clause.before == target {
				found = append(found, clause.name)
			}
		}
		return found
	}
	result := []string{}
	for i := 0; i < len(origin); i++ {
		clause := origin[i]
		appendClauses := collect(clause)
		result = append(result, appendClauses...)
		result = append(result, clause)
	}
	return slices.Compact(result)
}
