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

func GetBalance(name string) (int, error) {
	var balanceStr string
	row := db.QueryRow(`SELECT balance FROM Users WHERE name = ?`, name)
	err := row.Scan(&balanceStr)
	if err != nil {
		fmt.Println(err)
	}

	// convert balance from string to int
	balance, err := strconv.Atoi(balanceStr)
    if err != nil {
        // handle error
        fmt.Println(err)
    }

    return balance, err
}

func UpdateBalance(name string, balanceDifference int) error {
	balanceOld, err := GetBalance(name)
	if err != nil {
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
