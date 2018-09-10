package neoutils

import (
	"log"

	"github.com/o3labs/neo-utils/neoutils/smartcontract"
	"encoding/json"
)

type NativeAssetInterface interface {
	SendNativeAssetRawTransaction(wallet Wallet, asset smartcontract.NativeAsset, amount float64, to smartcontract.NEOAddress, unspent smartcontract.Unspent, attributes map[smartcontract.TransactionAttribute][]byte) ([]byte, string, error)
	GenerateRawTx(fromAddress string, asset smartcontract.NativeAsset, amount float64, to smartcontract.NEOAddress, unspent smartcontract.Unspent, attributes map[smartcontract.TransactionAttribute][]byte) ([]byte, error)
	SignRawTransaction(wallet Wallet, tx []byte) ([]byte, string, error)
}

type NativeAsset struct {
	NetworkFeeAmount smartcontract.NetworkFeeAmount //allow users to override the network fee here
}

func UseNativeAsset(networkFeeAmount smartcontract.NetworkFeeAmount) NativeAsset {
	return NativeAsset{
		NetworkFeeAmount: networkFeeAmount,
	}
}

var _ NativeAssetInterface = (*NativeAsset)(nil)

func (n *NativeAsset) SendNativeAssetRawTransaction(wallet Wallet, asset smartcontract.NativeAsset, amount float64, to smartcontract.NEOAddress, unspent smartcontract.Unspent, attributes map[smartcontract.TransactionAttribute][]byte) ([]byte, string, error) {
	txByte, err := n.GenerateRawTx(wallet.Address, asset, amount, to, unspent, attributes)
	if err != nil {
		return nil, "", err
	}

	return n.SignRawTransaction(wallet, txByte)
}

func (n *NativeAsset) GenerateRawTx(fromAddress string, asset smartcontract.NativeAsset, amount float64, to smartcontract.NEOAddress, unspent smartcontract.Unspent, attributes map[smartcontract.TransactionAttribute][]byte) ([]byte, error) {
	//New invocation transaction struct and fill with all necessary data
	tx := smartcontract.NewContractTransaction()

	//generate transaction inputs
	txInputs, err := smartcontract.NewScriptBuilder().GenerateTransactionInput(unspent, asset, amount, n.NetworkFeeAmount)
	if err != nil {
		return nil, err
	}
	tx.Inputs = txInputs

	txAttributes, err := smartcontract.NewScriptBuilder().GenerateTransactionAttributes(attributes)
	if err != nil {
		return nil, err
	}

	tx.Attributes = txAttributes

	sender := smartcontract.ParseNEOAddress(fromAddress)

	txOutputs, err := smartcontract.NewScriptBuilder().GenerateTransactionOutput(sender, to, unspent, asset, amount, n.NetworkFeeAmount)
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}

	tx.Outputs = txOutputs

	b, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (n *NativeAsset) SignRawTransaction(wallet Wallet, txBytes []byte) ([]byte, string, error) {
	tx := smartcontract.NewContractTransaction()
	err := json.Unmarshal(txBytes, &tx)
	if err != nil {
		return nil, "", err
	}

	//begin signing
	privateKeyInHex := bytesToHex(wallet.PrivateKey)
	signedData, err := Sign(tx.ToBytes(), privateKeyInHex)
	if err != nil {
		log.Printf("err signing %v", err)
		return nil, "", err
	}

	signature := smartcontract.TransactionSignature{
		SignedData: signedData,
		PublicKey:  wallet.PublicKey,
	}
	// try to verify it
	// hash := sha256.Sum256(tx.ToBytes())
	// valid := Verify(wallet.PublicKey, signedData, hash[:])
	// log.Printf("verify tx %v", valid)

	scripts := []interface{}{signature}
	txScripts := smartcontract.NewScriptBuilder().GenerateVerificationScripts(scripts)

	//concat data
	endPayload := []byte{}
	endPayload = append(endPayload, tx.ToBytes()...)
	endPayload = append(endPayload, txScripts...)

	return endPayload, tx.ToTXID(), nil
}