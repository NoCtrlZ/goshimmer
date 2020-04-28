package sctransaction

// loads last start transaction of the smart contract
func LoadStateTx(scid ScId) (*Transaction, error) {
	// find all outputs to the address scid.Address()
	// among those outputs must be exactly the only one with the color equal to scid.Color() and balance 1
	// the transactions which holds that output is the last state transactio of the SC
	// If can't find this balance, scid is wrong and does not represent a valid SC
	panic("implement me")
}
