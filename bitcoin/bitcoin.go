package bitcoin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

var password, guid string

// reads global password and guid variables from file
func Setup(confname string) error {
	txt, err := ioutil.ReadFile(confname)
	if err != nil {
		return err
	}

	var m map[string]interface{}
	json.Unmarshal(txt, &m)
	password = m["password"].(string)
	guid = m["guid"].(string)

	return nil
}

type RecvTransaction struct {
	txhash  string
	address string
	amount  int
}

func SetupReceiveCallback(unconfirmed chan *RecvTransaction,
	confirmed chan *RecvTransaction) {

	//
}

// helper function to query blockchain.info
func queryBlockchain(req string) (map[string]interface{}, error) {
	resp, err := http.Get(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(contents))

	var m map[string]interface{}
	json.Unmarshal(contents, &m)
	return m, nil
}

// returns a new address
func NewAddress() (string, error) {
	req := fmt.Sprintf(
		"https://blockchain.info/merchant/%s/new_address?password=%s",
		guid, password)

	m, err := queryBlockchain(req)
	if err != nil {
		return "", err
	}

	address, ok := m["address"]
	if !ok {
		return "", errors.New("Could not find address field in blockchain return.")
	}
	return address.(string), nil
}

// Sends amount satoshis to address
func SendCoins(address string, amount int) (txhash string, err error) {
	req := fmt.Sprintf(
		"https://blockchain.info/merchant/%s/payment?password=%s&to=%s&amount=%d",
		guid, password, address, amount)

	fmt.Println(req)

	m, err := queryBlockchain(req)
	if err != nil {
		return "", err
	}

	txhash_raw, ok := m["tx_hash"]
	if !ok {
		// TODO:
		// distinguish "No free outputs to spend" from
		// "com.google.bitcoin.core.AddressFormatException: Checksum does not validate"
		return "", errors.New("Transaction was not successfully executed.")
	}

	txhash = txhash_raw.(string)
	err = nil
	return
}
