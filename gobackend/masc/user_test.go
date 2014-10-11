package masc

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rogerwalt/GambleWithCoins/gobackend/bitcoin"
)

func TestUser(t *testing.T) {
	dbName := "./test.db"
	os.Remove(dbName)

	db, _ := sql.Open("sqlite3", dbName)
	//checkError(err)
	defer db.Close()

	SetupDb(db)

	Register("bla", "baz")
	if !Login("bla", "baz") {
		t.Error("Expected true, got false")
	}
	if Login("bla", "bas") {
		t.Error("Expected false, got true")
	}
	Register("bla", "bar")

	bitcoin.Setup("../bitcoin/blockchain-conf.test.json")
	// test deposit address
	addr1, err1 := getDepositAddress("bla")
	if err1 != nil {
		fmt.Println(err1.Error())
	}
	addr2, err2 := getDepositAddress("bla")
	if err2 != nil {
		fmt.Println(err2.Error())
	}
	if addr1 != addr2 {
		t.Error("Expected deposit addresses to be the same")
	}

	// test withdraw (just from db perspective)
	UpdateBalance("bla", 100)
	err := Withdraw("bla", 101, "addr")
	if err == nil {
		t.Error("Expected returning insufficient funds")
	}

	err = Withdraw("bla", 100, "addr")
	if err == nil {
		t.Error("Shouldn't have funds here")
	}

}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
