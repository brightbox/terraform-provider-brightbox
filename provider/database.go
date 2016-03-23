package brightbox

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

type MysqlDatabase struct {
	Statements []string
}

func (d *MysqlDatabase) DropDatabase(db_name string) {
	d.Statements = append(d.Statements,
		fmt.Sprintf("DROP DATABASE IF EXISTS %s", db_name),
	)
}

func (d *MysqlDatabase) CreateDatabase(db_name string) {
	d.Statements = append(d.Statements,
		fmt.Sprintf("CREATE DATABASE %s", db_name),
	)
}

func (d *MysqlDatabase) DropUser(user_name string) {
	d.Statements = append(d.Statements,
		fmt.Sprintf("DROP USER '%s'@'%%'", user_name),
	)
}
func (d *MysqlDatabase) CreateUser(user_name, password string) {
	d.Statements = append(d.Statements,
		fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s'", user_name, password),
	)
}

func (d *MysqlDatabase) SetPassword(user_name, password string) {
	d.Statements = append(d.Statements,
		fmt.Sprintf("SET PASSWORD FOR '%s'@'%%' = PASSWORD('%s')", user_name, password),
	)
}

func (d *MysqlDatabase) GrantPrivileges(db_name, user_name string) {
	d.Statements = append(d.Statements,
		fmt.Sprintf(
			"GRANT Alter, Alter routine, Create, Create routine, Create temporary tables, Create view, Delete, Drop, Index, Insert, Lock tables, References, Select, Show view, Update on %s.* TO '%s'@'%%'",
			db_name,
			user_name,
		),
	)
}

func (d *MysqlDatabase) OpenExec(username, password, host string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/?timeout=30s&strict=true&multiStatements=true", username, password, host)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("Failed to connect to Cloud SQL server: %s", err.Error())
	}
	defer db.Close()
	_, err = db.Exec(strings.Join(d.Statements, ";"))
	if err != nil {
		return fmt.Errorf("Failed to Exec SQL statements: %s", err.Error())
	}
	return nil
}
