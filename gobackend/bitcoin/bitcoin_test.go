package bitcoin

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNewAddress(t *testing.T) {
	err := Setup("./blockchain-conf.test.json")
	checkError(t, err)

	// test address generation
	address, err := NewAddress()
	checkError(t, err)
	fmt.Println("Generated new address: ", address)

	// test sending coins
	txhash, err := SendCoins(address, 100000)
	checkError(t, err)
	if err == nil {
		fmt.Println("New transaction txhash: ", txhash)
	}

	//test receiving coins
	fmt.Println("Receive coins: ", txhash)

	unconfirmed := make(chan *RecvTransaction)
	confirmed := make(chan *RecvTransaction)
	http.HandleFunc(fmt.Sprintf("/receive/%s/", Callback_secret),
		ReceiveCallback(unconfirmed, confirmed))
	go http.ListenAndServe(":8080", nil)
	for {
		select {
		case o := <-unconfirmed:
			fmt.Println("unconfirmed", o.Txhash, o.Address, o.Amount)

		case o := <-confirmed:
			fmt.Println("confirmed", o.Txhash, o.Address, o.Amount)
		}
	}
}

func checkError(t *testing.T, err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}
