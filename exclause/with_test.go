package exclause

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	extraClausePlugin "github.com/WinterYukky/gorm-extra-clause-plugin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestWith_Query(t *testing.T) {
	tests := []struct {
		name      string
		operation func(db *gorm.DB) *gorm.DB
		want      string
		wantArgs  []driver.Value
	}{
		{
			name: "String query should be used as is",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(NewWith(db, "cte", "SELECT * FROM `users`")).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "String query with args should be used as is",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(NewWith(db, "cte", "SELECT * FROM `users` WHERE `name` = ?", "WinterYukky")).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` AS (SELECT * FROM `users` WHERE `name` = ?) SELECT * FROM `cte`",
			wantArgs: []driver.Value{"WinterYukky"},
		},
		{
			name: "DB query should be built and used",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(NewWith(db, "cte", db.Table("users"))).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "DB query with args should be built and used",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(NewWith(db, "cte", db.Table("users").Where("`name` = ?", "WinterYukky"))).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` AS (SELECT * FROM `users` WHERE `name` = ?) SELECT * FROM `cte`",
			wantArgs: []driver.Value{"WinterYukky"},
		},
		{
			name: "CTE alias with columns should be used with columns specified",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(NewWith(db, CTE{Alias: "cte", Columns: []string{"id", "name"}}, db.Table("users"))).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` (`id`,`name`) AS (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "CTE alias with recursive should be used as recursive",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(NewWith(db, CTE{Recursive: true, Alias: "cte"}, db.Table("users"))).Table("cte").Scan(nil)
			},
			want:     "WITH RECURSIVE `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "Mulitiple CTEs should be used only one WITH keyword",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(NewWith(db, "cte1", db.Table("users"))).Clauses(NewWith(db, "cte2", db.Table("users"))).
					Table("cte").Scan(nil)
			},
			want:     "WITH `cte1` AS (SELECT * FROM `users`),`cte2` AS (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "Mulitiple recursive CTEs should be used only one RECURSIVE keyword",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(NewWith(db, CTE{Recursive: true, Alias: "cte1"}, db.Table("users"))).
					Clauses(NewWith(db, CTE{Recursive: true, Alias: "cte2"}, db.Table("users"))).
					Table("cte").Scan(nil)
			},
			want:     "WITH RECURSIVE `cte1` AS (SELECT * FROM `users`),`cte2` AS (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer mockDB.Close()
			db, _ := gorm.Open(mysql.New(mysql.Config{
				Conn:                      mockDB,
				SkipInitializeWithVersion: true,
			}))
			db.Use(extraClausePlugin.New())
			mock.ExpectQuery(regexp.QuoteMeta(tt.want)).WithArgs(tt.wantArgs...).WillReturnRows(sqlmock.NewRows([]string{}))
			if tt.operation != nil {
				db = tt.operation(db)
			}
			if db.Error != nil {
				t.Errorf(db.Error.Error())
			}
		})
	}
}
