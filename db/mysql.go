package db

import (
	_ "database/sql"
	_ "encoding/json"
	"log"

	"../model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	ConnectString string
}

func InitDb(cfg Config) (*sqlDb, error) {
	if dbConn, err := sqlx.Connect("mysql", cfg.ConnectString); err != nil {
		return nil, err
	} else {
		p := &sqlDb{dbConn: dbConn}
		if err := p.dbConn.Ping(); err != nil {
			return nil, err
		}
		if err := p.createTablesIfNotExist(); err != nil {
			return nil, err
		}
		if err := p.prepareSqlStatements(); err != nil {
			return nil, err
		}
		return p, nil
	}
}

type sqlDb struct {
	dbConn *sqlx.DB

	sqlSelectPeople *sqlx.Stmt
	sqlInsertUser   *sqlx.NamedStmt
	sqlLoginUser    *sqlx.Stmt
}

func (p *sqlDb) createTablesIfNotExist() error {
	create_sql := `

       CREATE TABLE IF NOT EXISTS people (
       id SERIAL NOT NULL PRIMARY KEY,
       name TEXT NOT NULL,
       login TEXT NOT NULL,
       password TEXT NOT NULL);

    `
	if rows, err := p.dbConn.Query(create_sql); err != nil {
		return err
	} else {
		rows.Close()
	}
	return nil
}

func (p *sqlDb) prepareSqlStatements() (err error) {

	if p.sqlSelectPeople, err = p.dbConn.Preparex(
		"SELECT name, login, password FROM people",
	); err != nil {
		return err
	}
	if p.sqlInsertUser, err = p.dbConn.PrepareNamed(
		"INSERT INTO people (name, login, password) VALUES (:name, :login, :password) ", //+ "RETURNING id, name, login, password",
		//"INSERT INTO people (name, login, password) VALUES ( ?, ?, ? ) ", //+ "RETURNING id, name, login, password",
		/* использование:
		_, err = p.sqlInsertUser.Exec( "1", "A", "qwe") // Insert tuples (i, i^2)
		if err != nil {	panic(err.Error()) }
		*/
	); err != nil {
		return err
	}
	//defer p.sqlInsertUser.Close()

	if p.sqlLoginUser, err = p.dbConn.Preparex(
		"SELECT id, name, login, password FROM people WHERE login = ? and password = ?",
		/* использование:
		var result
		err = stmtOut.QueryRow("Name").Scan(&result) // WHERE name = "Name"
		if err != nil {	panic(err.Error()) }
		*/

	); err != nil {
		return err
	}

	return nil
}

func (p *sqlDb) SelectPeople() ([]*model.User, error) {
	people := make([]*model.User, 0)
	if err := p.sqlSelectPeople.Select(&people); err != nil {
		return nil, err
	}
	return people, nil
}

func (p *sqlDb) LoginUser(login, password string) ([]*model.User, error) {
	people := make([]*model.User, 0)
	log.Println("func LoginUser")
	if err := p.sqlLoginUser.Select(&people, login, password); err != nil {
		log.Println(err)
		return nil, err
	}
	return people, nil
}
