package bitcoin

import (
	"fmt"
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
	fmt.Println("New transaction txhash: ", txhash)

	//test receiving coins

}

func checkError(t *testing.T, err error) {
	if err != nil {
		t.Errorf(err.Error())
	}
}
