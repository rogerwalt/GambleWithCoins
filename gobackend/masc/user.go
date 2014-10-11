package masc

import "fmt"

//TODO: use transactions

func register(name, password string) {
	result, err := db.Exec(`CREATE TABLE Users(
							id INTEGER PRIMARY KEY AUTOINCREMENT,
							name TEXT UNIQUE,
							password TEXT,
							balance INTEGER,
							depositAddress TEXT);`)
	if err != nil {
		fmt.Println(err)
	}

	result, err = db.Exec(`INSERT INTO Users 
							(name, password, balance) 
							VALUES (?, ?, 0);`, name, password)
	fmt.Println(result)
	if err != nil {
		fmt.Println(err)
	}
}

func getAddress() {

}

func login(name, password string) bool {
	var expected string
	row := db.QueryRow("SELECT password FROM Users WHERE name = ?", name)
	err := row.Scan(&expected)
	if err != nil {
		fmt.Println(err)
	}
	return expected == password
}
