package masc

import (
	"database/sql"
	"fmt"
	"strconv"
)

//TODO: use transactions

var db *sql.DB

// call before using any function in this file
func SetupDb(adb *sql.DB) error {
	db = adb
	_, err := db.Exec(`CREATE TABLE Users(
							id INTEGER PRIMARY KEY AUTOINCREMENT,
							name TEXT UNIQUE,
							password TEXT,
							balance INTEGER,
							depositAddress TEXT);`)
	return err
}

func Register(name, password string) error {
	_, err := db.Exec(`INSERT INTO Users 
							(name, password, balance) 
							VALUES (?, ?, 0);`, name, password)
	return err
}

func UpdateBalance(name string, balanceDifference int) error {
	var balanceOldStr string
	row := db.QueryRow(`SELECT balance FROM Users WHERE name = ?`, name)
	err := row.Scan(&balanceOldStr)
	if err != nil {
		fmt.Println(err)
	}

	// convert balance from string to int
	balanceOld, err := strconv.Atoi(balanceOldStr)
    if err != nil {
        // handle error
        fmt.Println(err)
    }

	balanceNew := balanceOld + balanceDifference

	_, err = db.Exec(`UPDATE Users SET balance = ? WHERE name = ?`, balanceNew, name)
	return err
}

func getAddress() {

}

func Login(name, password string) bool {
	var expected string
	row := db.QueryRow("SELECT password FROM Users WHERE name = ?", name)
	err := row.Scan(&expected)
	if err != nil {
		fmt.Println(err)
	}
	return expected == password
}
