package gormextraclauseplugin

import (
	"slices"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestInstall(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db, _ := gorm.Open(mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	}))
	err = db.Use(New())
	if err != nil {
		t.Fatalf("an error '%s' was not expected when registering the plugin", err)
	}

	_, ok := db.Plugins["ExtraClausePlugin"]
	if !ok {
		t.Errorf("Could not find ExtraClausePlugin after registration")
	}
}

func TestQueryClauses_Default(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db, _ := gorm.Open(mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	}))
	db.Use(New())
	got := db.Callback().Query().Clauses
	want := []string{"WITH", "SELECT", "FROM", "WHERE", "GROUP BY", "UNION", "INTERSECT", "EXCEPT", "ORDER BY", "LIMIT", "FOR"}
	if !slices.Equal(got, want) {
		t.Errorf("Query clauses is %v, want %v", got, want)
	}
}
func TestQueryClauses_Customized(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db, _ := gorm.Open(mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	}))
	db.Callback().Query().Clauses = []string{"FOO", "SELECT", "FROM", "WHERE", "BAR", "GROUP BY", "ORDER BY", "LIMIT", "FOR"}
	db.Use(New())
	got := db.Callback().Query().Clauses
	want := []string{"FOO", "WITH", "SELECT", "FROM", "WHERE", "BAR", "GROUP BY", "UNION", "INTERSECT", "EXCEPT", "ORDER BY", "LIMIT", "FOR"}
	if !slices.Equal(got, want) {
		t.Errorf("Query clauses is %v, want %v", got, want)
	}
}

func TestRowClauses_Default(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db, _ := gorm.Open(mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	}))
	db.Use(New())
	got := db.Callback().Row().Clauses
	want := []string{"WITH", "SELECT", "FROM", "WHERE", "GROUP BY", "UNION", "INTERSECT", "EXCEPT", "ORDER BY", "LIMIT", "FOR"}
	if !slices.Equal(got, want) {
		t.Errorf("Row clauses is %v, want %v", got, want)
	}
}
func TestRowClauses_Customized(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db, _ := gorm.Open(mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	}))
	db.Callback().Row().Clauses = []string{"FOO", "SELECT", "FROM", "WHERE", "BAR", "GROUP BY", "ORDER BY", "LIMIT", "FOR"}
	db.Use(New())
	got := db.Callback().Row().Clauses
	want := []string{"FOO", "WITH", "SELECT", "FROM", "WHERE", "BAR", "GROUP BY", "UNION", "INTERSECT", "EXCEPT", "ORDER BY", "LIMIT", "FOR"}
	if !slices.Equal(got, want) {
		t.Errorf("Row clauses is %v, want %v", got, want)
	}
}

func TestUpdateClauses_Default(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db, _ := gorm.Open(mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	}))
	db.Use(New())
	got := db.Callback().Update().Clauses
	want := []string{"WITH", "UPDATE", "SET", "WHERE", "UNION", "INTERSECT", "EXCEPT", "ORDER BY", "LIMIT"}
	if !slices.Equal(got, want) {
		t.Errorf("Update clauses is %v, want %v", got, want)
	}
}
func TestUpdateClauses_Customized(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db, _ := gorm.Open(mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	}))
	db.Callback().Update().Clauses = []string{"FOO", "WITH", "UPDATE", "SET", "WHERE", "BAR", "ORDER BY", "BAZ", "LIMIT", "FOR"}
	db.Use(New())
	got := db.Callback().Update().Clauses
	want := []string{"FOO", "WITH", "UPDATE", "SET", "WHERE", "BAR", "UNION", "INTERSECT", "EXCEPT", "ORDER BY", "BAZ", "LIMIT", "FOR"}
	if !slices.Equal(got, want) {
		t.Errorf("Update clauses is %v, want %v", got, want)
	}
}
