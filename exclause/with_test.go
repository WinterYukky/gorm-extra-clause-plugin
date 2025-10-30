package exclause

import (
	"database/sql/driver"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	extraClausePlugin "github.com/WinterYukky/gorm-extra-clause-plugin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestWith_Query(t *testing.T) {
	tests := []struct {
		name      string
		operation func(db *gorm.DB) *gorm.DB
		want      string
		wantArgs  []driver.Value
	}{
		{
			name: "When Subquery is clause.Expr, then should be used as subquery",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{{Name: "cte", Subquery: clause.Expr{SQL: "SELECT * FROM `users` WHERE `name` = ?", Vars: []interface{}{"WinterYukky"}}}}}).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` AS (SELECT * FROM `users` WHERE `name` = ?) SELECT * FROM `cte`",
			wantArgs: []driver.Value{"WinterYukky"},
		},
		{
			name: "When Subquery is exclause.Subquery, then should be used as subquery",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{{Name: "cte", Subquery: Subquery{DB: db.Table("users").Where("`name` = ?", "WinterYukky")}}}}).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` AS (SELECT * FROM `users` WHERE `name` = ?) SELECT * FROM `cte`",
			wantArgs: []driver.Value{"WinterYukky"},
		},
		{
			name: "When has specific fields, then should be used with columns specified",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{{Name: "cte", Columns: []string{"id", "name"}, Subquery: Subquery{DB: db.Table("users")}}}}).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` (`id`,`name`) AS (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "When contains recursive even once, then should be used RECURSIVE keyword",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.
					Clauses(With{Recursive: true, CTEs: []CTE{{Name: "cte1", Subquery: Subquery{DB: db.Table("users")}}}}).
					Clauses(With{Recursive: false, CTEs: []CTE{{Name: "cte2", Subquery: Subquery{DB: db.Table("users")}}}}).
					Table("cte").Scan(nil)
			},
			want:     "WITH RECURSIVE `cte1` AS (SELECT * FROM `users`),`cte2` AS (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "When Materialized is CTEMaterialize, then should use MATERIALIZED keyword",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{{Name: "cte", Subquery: Subquery{DB: db.Table("users")}, Materialized: CTEMaterialize}}}).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` AS MATERIALIZED (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "When Materialized is CTENotMaterialize, then should use NOT MATERIALIZED keyword",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{{Name: "cte", Subquery: Subquery{DB: db.Table("users")}, Materialized: CTENotMaterialize}}}).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` AS NOT MATERIALIZED (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "When Materialized is CTEMaterializeUnspecified, then should not use any materialization keyword",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{{Name: "cte", Subquery: Subquery{DB: db.Table("users")}, Materialized: CTEMaterializeUnspecified}}}).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "When multiple CTEs with different materialization options",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{
					CTEs: []CTE{
						{Name: "cte1", Subquery: Subquery{DB: db.Table("users")}, Materialized: CTEMaterialize},
						{Name: "cte2", Subquery: Subquery{DB: db.Table("products")}, Materialized: CTENotMaterialize},
						{Name: "cte3", Subquery: Subquery{DB: db.Table("orders")}, Materialized: CTEMaterializeUnspecified},
					},
				}).Table("cte1").Scan(nil)
			},
			want:     "WITH `cte1` AS MATERIALIZED (SELECT * FROM `users`),`cte2` AS NOT MATERIALIZED (SELECT * FROM `products`),`cte3` AS (SELECT * FROM `orders`) SELECT * FROM `cte1`",
			wantArgs: []driver.Value{},
		},
		{
			name: "When using NewMaterializedCTE helper",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{NewMaterializedCTE("cte", Subquery{DB: db.Table("users")})}}).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` AS MATERIALIZED (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "When using NewNotMaterializedCTE helper",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{NewNotMaterializedCTE("cte", Subquery{DB: db.Table("users")})}}).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` AS NOT MATERIALIZED (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "When using NewCTE helper (unspecified materialization)",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{NewCTE("cte", Subquery{DB: db.Table("users")})}}).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` AS (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "When RECURSIVE with MATERIALIZED",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{
					Recursive: true,
					CTEs:      []CTE{{Name: "cte", Subquery: Subquery{DB: db.Table("users")}, Materialized: CTEMaterialize}},
				}).Table("cte").Scan(nil)
			},
			want:     "WITH RECURSIVE `cte` AS MATERIALIZED (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "When RECURSIVE with NOT MATERIALIZED",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{
					Recursive: true,
					CTEs:      []CTE{{Name: "cte", Subquery: Subquery{DB: db.Table("users")}, Materialized: CTENotMaterialize}},
				}).Table("cte").Scan(nil)
			},
			want:     "WITH RECURSIVE `cte` AS NOT MATERIALIZED (SELECT * FROM `users`) SELECT * FROM `cte`",
			wantArgs: []driver.Value{},
		},
		{
			name: "When materialized with clause.Expr subquery",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{{Name: "cte", Subquery: clause.Expr{SQL: "SELECT * FROM `users` WHERE `name` = ?", Vars: []interface{}{"WinterYukky"}}, Materialized: CTEMaterialize}}}).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` AS MATERIALIZED (SELECT * FROM `users` WHERE `name` = ?) SELECT * FROM `cte`",
			wantArgs: []driver.Value{"WinterYukky"},
		},
		{
			name: "When materialized with columns specified",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{{Name: "cte", Columns: []string{"id", "name"}, Subquery: Subquery{DB: db.Table("users")}, Materialized: CTEMaterialize}}}).Table("cte").Scan(nil)
			},
			want:     "WITH `cte` (`id`,`name`) AS MATERIALIZED (SELECT * FROM `users`) SELECT * FROM `cte`",
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

func TestWith_Update(t *testing.T) {
	tests := []struct {
		name      string
		operation func(db *gorm.DB) *gorm.DB
		want      string
		wantArgs  []driver.Value
	}{
		{
			name: "When Subquery is clause.Expr, then should be used as subquery",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{{Name: "cte", Subquery: clause.Expr{SQL: "SELECT * FROM `users` WHERE `name` = ?", Vars: []interface{}{"WinterYukky"}}}}}).Table("users").Where("`users`.`id` IN (SELECT `id` FROM `cte`)").Update("name", "new_name")
			},
			want:     "WITH `cte` AS (SELECT * FROM `users` WHERE `name` = ?) UPDATE `users` SET `name`=? WHERE `users`.`id` IN (SELECT `id` FROM `cte`)",
			wantArgs: []driver.Value{"WinterYukky", "new_name"},
		},
		{
			name: "When Subquery is exclause.Subquery, then should be used as subquery",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{{Name: "cte", Subquery: Subquery{DB: db.Table("users").Where("`name` = ?", "WinterYukky")}}}}).Table("users").Where("`users`.`id` IN (SELECT `id` FROM `cte`)").Update("name", "new_name")
			},
			want:     "WITH `cte` AS (SELECT * FROM `users` WHERE `name` = ?) UPDATE `users` SET `name`=? WHERE `users`.`id` IN (SELECT `id` FROM `cte`)",
			wantArgs: []driver.Value{"WinterYukky", "new_name"},
		},
		{
			name: "When has specific fields, then should be used with columns specified",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{{Name: "cte", Columns: []string{"id", "name"}, Subquery: Subquery{DB: db.Table("users")}}}}).Table("users").Where("`users`.`id` IN (SELECT `id` FROM `cte`)").Update("name", "new_name")
			},
			want:     "WITH `cte` (`id`,`name`) AS (SELECT * FROM `users`) UPDATE `users` SET `name`=? WHERE `users`.`id` IN (SELECT `id` FROM `cte`)",
			wantArgs: []driver.Value{"new_name"},
		},
		{
			name: "When contains recursive even once, then should be used RECURSIVE keyword",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.
					Clauses(With{Recursive: true, CTEs: []CTE{{Name: "cte1", Subquery: Subquery{DB: db.Table("users")}}}}).
					Clauses(With{Recursive: false, CTEs: []CTE{{Name: "cte2", Subquery: Subquery{DB: db.Table("users")}}}}).
					Table("users").Where("`users`.`id` IN (SELECT `id` FROM `cte`)").Update("name", "new_name")
			},
			want:     "WITH RECURSIVE `cte1` AS (SELECT * FROM `users`),`cte2` AS (SELECT * FROM `users`) UPDATE `users` SET `name`=? WHERE `users`.`id` IN (SELECT `id` FROM `cte`)",
			wantArgs: []driver.Value{"new_name"},
		},
		{
			name: "When Materialized is CTEMaterialize in update",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{{Name: "cte", Subquery: Subquery{DB: db.Table("users").Where("`name` = ?", "WinterYukky")}, Materialized: CTEMaterialize}}}).Table("users").Where("`users`.`id` IN (SELECT `id` FROM `cte`)").Update("name", "new_name")
			},
			want:     "WITH `cte` AS MATERIALIZED (SELECT * FROM `users` WHERE `name` = ?) UPDATE `users` SET `name`=? WHERE `users`.`id` IN (SELECT `id` FROM `cte`)",
			wantArgs: []driver.Value{"WinterYukky", "new_name"},
		},
		{
			name: "When Materialized is CTENotMaterialize in update",
			operation: func(db *gorm.DB) *gorm.DB {
				return db.Clauses(With{CTEs: []CTE{{Name: "cte", Subquery: Subquery{DB: db.Table("users").Where("`name` = ?", "WinterYukky")}, Materialized: CTENotMaterialize}}}).Table("users").Where("`users`.`id` IN (SELECT `id` FROM `cte`)").Update("name", "new_name")
			},
			want:     "WITH `cte` AS NOT MATERIALIZED (SELECT * FROM `users` WHERE `name` = ?) UPDATE `users` SET `name`=? WHERE `users`.`id` IN (SELECT `id` FROM `cte`)",
			wantArgs: []driver.Value{"WinterYukky", "new_name"},
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
			mock.ExpectBegin()
			mock.ExpectExec(regexp.QuoteMeta(tt.want)).WithArgs(tt.wantArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			if tt.operation != nil {
				db = tt.operation(db)
			}
			if db.Error != nil {
				t.Errorf(db.Error.Error())
			}
		})
	}
}

func TestNewWith(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db, _ := gorm.Open(mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	}))
	db = db.Table("users")
	type args struct {
		name     string
		subquery interface{}
		args     []interface{}
	}
	tests := []struct {
		name string
		args args
		want With
	}{
		{
			name: "When subquery is *gorm.DB, then CTE's Subquery is exclause.Subquery",
			args: args{
				name:     "cte",
				subquery: db,
			},
			want: With{
				Recursive: false,
				CTEs: []CTE{
					{
						Name:     "cte",
						Subquery: Subquery{DB: db},
					},
				},
			},
		},
		{
			name: "When subquery is string, then CTE's Subquery is clause.Expr",
			args: args{
				name:     "cte",
				subquery: "SELECT * FROM `users` WHERE `name` = ?",
				args:     []interface{}{"WinterYukky"},
			},
			want: With{
				Recursive: false,
				CTEs: []CTE{
					{
						Name: "cte",
						Subquery: clause.Expr{
							SQL:  "SELECT * FROM `users` WHERE `name` = ?",
							Vars: []interface{}{"WinterYukky"},
						},
					},
				},
			},
		},
		{
			name: "When subquery is else, then CTE's Subquery is empty With",
			args: args{
				name:     "cte",
				subquery: 0,
			},
			want: With{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewWith(tt.args.name, tt.args.subquery, tt.args.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWith() = %v, want %v", got, tt.want)
			}
		})
	}
}
