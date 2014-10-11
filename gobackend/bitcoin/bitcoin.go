package bitcoin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

var password, guid, callback_secret string

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
	callback_secret = m["callback_secret"].(string)

	return nil
}

type RecvTransaction struct {
	txhash  string
	address string
	amount  int
}

// function takes two channels over which it sends the transactions
// and returns a callback secret to include in the URL to listen for callbacks
// and returns a handler for that URL
// note that confirmed channel might return transaction several times
func ReceiveCallback(unconfirmed, confirmed chan *RecvTransaction) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		//TODO: check if response comes from blockchain domain, https
		// sanitize parameters
		params := r.URL.Query()
		confirmations, _ := strconv.Atoi(params["confirmations"][0])
		value, _ := strconv.Atoi(params["value"][0])
		tx := &RecvTransaction{params["transaction_hash"][0],
			params["input_address"][0],
			value}

		if confirmations == 0 {
			unconfirmed <- tx
		} else if confirmations == 2 {
			confirmed <- tx
			fmt.Fprintf(w, "*ok*")
		} else if confirmations <= 4 {
			log.Println("Callback from received despite having sent ok")

			confirmed <- tx
			fmt.Fprintf(w, "*ok*")
		}
	}
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
