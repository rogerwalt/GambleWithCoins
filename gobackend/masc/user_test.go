package masc

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestUser(t *testing.T) {
	dbName := "./test.db"
	os.Remove(dbName)

	db, _ := sql.Open("sqlite3", dbName)
	//checkError(err)
	defer db.Close()

	setup(db)

	register("bla", "baz")
	if !login("bla", "baz") {
		t.Error("Expected true, got false")
	}
	if login("bla", "bas") {
		t.Error("Expected false, got true")
	}
	register("bla", "bar")
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
