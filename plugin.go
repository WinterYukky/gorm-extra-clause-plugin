package gormextraclauseplugin

import "gorm.io/gorm"

// ExtraClausePlugin support plugin that not supported clause by gorm
type ExtraClausePlugin struct{}

// Name return plugin name
func (e *ExtraClausePlugin) Name() string {
	return "ExtraCalusePlugin"
}

// Initialize register BuildClauses
func (e *ExtraClausePlugin) Initialize(db *gorm.DB) error {
	db.Callback().Query().Clauses = []string{"WITH", "SELECT", "FROM", "WHERE", "GROUP BY", "UNION", "ORDER BY", "LIMIT", "FOR"}
	db.Callback().Row().Clauses = []string{"WITH", "SELECT", "FROM", "WHERE", "GROUP BY", "UNION", "ORDER BY", "LIMIT", "FOR"}
	db.Callback().Update().Clauses = []string{"WITH", "UPDATE", "SET", "FROM", "WHERE"}
	return nil
}

// New create new ExtraClausePlugin
//  // example
//  db.Use(extraClausePlugin.New())
func New() *ExtraClausePlugin {
	return &ExtraClausePlugin{}
}
