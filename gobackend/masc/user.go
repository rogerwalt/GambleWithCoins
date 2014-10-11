package masc

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/rogerwalt/GambleWithCoins/gobackend/bitcoin"
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
							depositAddress TEXT,
							cooperate INTEGER,
							defect INTEGER);`)
	return err
}

func Register(name, password string) error {
	_, err := db.Exec(`INSERT INTO Users 
							(name, password, balance, depositAddress, cooperate, defect) 
							VALUES (?, ?, 0, "NULL", 0, 0);`, name, password)
	return err
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

func AddAction(name, action string) bool {
	var err error
	if action == "cooperate" {
		_, err = db.Exec("UPDATE Users SET cooperate = cooperate + 1 WHERE name = ?", name)
	} else if action == "defect" {
		_, err = db.Exec("UPDATE Users SET defect = defect + 1 WHERE name = ?", name)
	} else {
		fmt.Println("Action does not exist")
	}
	
	if err != nil {
		fmt.Println(err)
	}
	return true
}

//TODO: secure against race conditions
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

// sends amount of satoshis to given address
//TODO: do proper transaction
func Withdraw(name string, amount int, address string) error {
	balance, err := GetBalance(name)
	if err != nil {
		return err
	}
	if balance < amount {
		return errors.New("Insufficient funds.")
	}

	err = UpdateBalance(name, -amount)
	if err != nil {
		return err
	}
	//TODO: store txhash
	_, err = bitcoin.SendCoins(address, amount)
	if err != nil {
		UpdateBalance(name, amount)
		return err
	}
	return nil
}

// gets deposit address from database
// if there is no deposit address yet for the user, create a new one
func GetDepositAddress(name string) (string, error) {
	var depositAddress string
	row := db.QueryRow("SELECT depositAddress FROM Users WHERE name = ?", name)
	err := row.Scan(&depositAddress)
	if err != nil {
		return "", err
	}

	if depositAddress == "NULL" {
		depositAddress, err = bitcoin.NewAddress()
		if err != nil {
			return "", err
		}
		_, err = db.Exec(`UPDATE Users SET depositAddress = ? WHERE name = ?`, depositAddress, name)
		if err != nil {
			return "", err
		}
	}
	return depositAddress, nil
}

func GetNameFromDepositAddress(address string) (string, error) {
	var name string
	//TODO: index on depositAddress
	row := db.QueryRow("SELECT name FROM Users WHERE depositAddress = ?", address)
	err := row.Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}

// TODO: store transaction id
func InsertIncomingTransactionsInDb(confirmed chan *bitcoin.RecvTransaction) {
	for {
		select {
		case tx := <-confirmed:
			name, err := GetNameFromDepositAddress(tx.Address)
			if err != nil {
				continue
			}
			UpdateBalance(name, tx.Amount)
		}
	}
}
