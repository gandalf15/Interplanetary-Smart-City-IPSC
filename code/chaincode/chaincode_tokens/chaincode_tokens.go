package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Chaincode implements Chaincode interface
type Chaincode struct {
}

// Variable names in a struct must be capitalised. Otherwise they are not exported (also to JSON)

// Account represents account of a member
type Account struct {
	RecordType    string // RecordType is used to distinguish the various types of objects in state database
	AccountID     string // unique id of the account
	Name          string // name of the account holder
	Tokens        int64  // amount of tokens (money)
	PendingTokens int64  // pending tokens from pending transactions
}

// limitTokens - limits the highest number of tokens that can be transfered
// from account without immediate verification of available tokens.
// This provides high throughput required for IoT data an many transactions per sec
var limitTokens int64 = 1

// Main function
/////////////////
func main() {
	// increase max CPU
	// runtime.GOMAXPROCS(runtime.NumCPU())
	err := shim.Start(new(Chaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode - Creates initial amount of tokens in two accounts
/////////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	// create initial ammount of tokens
	var err error
	//    		0                	1                     2
	// "NumberOfAccount" "Initial amount of tokens" "limitTokens"
	args := stub.GetStringArgs()
	if len(args) != 3 {
		return shim.Error(`Incorect number of arguments.
			Expectiong number of accounts and tokens for each account to create`)
	}
	// Input sanitation
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}

	noOfAccounts, err := strconv.Atoi(args[0])
	if err != nil || noOfAccounts < 0 {
		return shim.Error("Expecting positiv integer or zero as number of accounts to create.")
	}
	if noOfAccounts == 0 {
		return shim.Success(nil)
	}
	tokens, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil || tokens < 0 {
		return shim.Error("Expecting positiv integer or zero as number of tokens to init.")
	}
	limitTokens, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil || limitTokens < 0 {
		return shim.Error("Expecting positiv integer or zero as number of limit tokens to init.")
	}

	accounts := make([]*Account, noOfAccounts)
	for i := 0; i < noOfAccounts; i++ {
		accounts[i] = &Account{"ACCOUNT", strconv.Itoa(i + 1), "Init_Account", tokens, 0}
	}
	var accountJSONasBytes []byte
	for i := 0; i < noOfAccounts; i++ {
		accountJSONasBytes, err = json.Marshal(accounts[i])
		if err != nil {
			return shim.Error(err.Error())
		}
		err = stub.PutState(accounts[i].AccountID, accountJSONasBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	//  Index the account to enable name-based range queries
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	for i := 0; i < noOfAccounts; i++ {
		txID := stub.GetTxID()
		txRecipientIDCompositeKey, err := stub.CreateCompositeKey("Account~op~Tok~TxID", []string{strconv.Itoa(i), "+", strconv.FormatInt(tokens, 10), txID})
		if err != nil {
			return shim.Error(err.Error())
		}

		// Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
		// Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
		value := []byte{0x00}
		stub.PutState(txRecipientIDCompositeKey, value)
		if err != nil {
			return shim.Error(err.Error())
		}
		// Tx entry saved and indexed

		nameIDIndexKey, err := stub.CreateCompositeKey("Name~AccountID", []string{accounts[i].Name, accounts[i].AccountID})
		if err != nil {
			return shim.Error(err.Error())
		}
		stub.PutState(nameIDIndexKey, value)
	}

	return shim.Success(nil)
}

// Invoke - Entry point for Invocations
////////////////////////////////////////////
func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	// Handle different functions
	if function == "createAccount" { //create a new account
		return cc.createAccount(stub, args)
	} else if function == "deleteAccountByID" { // delete an account by account Id
		return cc.deleteAccountByID(stub, args)
	} else if function == "getAccountByID" { // get an account by its Id
		return cc.getAccountByID(stub, args)
	} else if function == "queryAccountByName" { // find an account base on name of account holder
		return cc.queryAccountByName(stub, args)
	} else if function == "sendTokensFast" { // transfer tokens from one account to another without check
		return cc.sendTokensFast(stub, args)
	} else if function == "sendTokensSafe" { // transfer tokens from one account to another with check
		return cc.sendTokensSafe(stub, args)
	} else if function == "updateAccountTokens" { // update state of account (value of tokens)
		return cc.updateAccountTokens(stub, args)
	} else if function == "getAccountTokens" { // get the current value of tokens on account
		return cc.getAccountTokens(stub, args)
	} else if function == "getAccountHistoryByID" { // get history for an account by its Id
		return cc.getAccountHistoryByID(stub, args)
		/*
			} else if function == "getTxParticipants" { // get recipient ID based on TxID
				return cc.getTxParticipants(stub, args)
			} else if function == "addTxAsUsed" { // Add TxID as used to state and index it.
				return cc.addTxAsUsed(stub, args)
		*/
	}

	return shim.Error("Received unknown function invocation")
}

// createAccount - create a new account and store into chaincode state
///////////////////////////////////////////////////////////////////////
func (cc *Chaincode) createAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//      0         1
	// "AccountID", "Name"
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting account Id and name")
	}

	// Input sanitation
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}

	accountID := args[0]
	name := args[1]

	// Check if an account already exists
	accountAsBytes, err := stub.GetState(accountID)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	} else if accountAsBytes != nil {
		return shim.Error("This account already exists: " + accountID)
	}

	// Create Account object and marshal to JSON
	recordType := "ACCOUNT"
	accountEntry := &Account{recordType, accountID, name, 0, 0}
	accountEntryJSONasBytes, err := json.Marshal(accountEntry)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Save account entry to state
	err = stub.PutState(accountID, accountEntryJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Index the account to enable name-based range queries
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	indexName := "Name~AccountID"
	nameIDIndexKey, err := stub.CreateCompositeKey(indexName, []string{accountEntry.Name, accountEntry.AccountID})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(nameIDIndexKey, value)

	// Account saved and indexed. Return success
	return shim.Success([]byte("Account created"))
}

// deleteAccountByID - deletes the account if number of tokens is 0 and no pending Tx or pending tokens
///////////////////////////////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) deleteAccountByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	//       0
	// deleteAccountId
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting AccountID.")
	}
	// Input sanitation
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	accountID := args[0]
	accountAsBytes, err := stub.GetState(accountID) //get the account entry from chaincode state
	var account Account
	err = json.Unmarshal(accountAsBytes, &account)
	if err != nil {
		return shim.Error(err.Error())
	}
	if (account.Tokens == 0) && (account.PendingTokens == 0) {
		// Delete the account state
		err = stub.DelState(accountID)
		if err != nil {
			return shim.Error("Failed to delete state:" + err.Error())
		}

		// maintain the index
		indexName := "Name~AccountID"
		nameIDIndexKey, err := stub.CreateCompositeKey(indexName, []string{account.Name, account.AccountID})
		if err != nil {
			return shim.Error(err.Error())
		}

		//  Delete index entry to state.
		err = stub.DelState(nameIDIndexKey)
		if err != nil {
			return shim.Error("Failed to delete state:" + err.Error())
		}

		return shim.Success(nil)
	}

	return shim.Error("Account cannot be deleted. Amount of tokens or pending tokens is not 0.")

}

// getAccountByID - read account entry from chaincode state based on its Id
////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) getAccountByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting account ID")
	}

	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	accountID := args[0]
	// Get all deltas for the variable

	accountAsBytes, err := stub.GetState(accountID) //get the account entry from chaincode state
	if err != nil {
		return shim.Error(err.Error())
	} else if accountAsBytes == nil {
		return shim.Error(err.Error())
	}

	return shim.Success(accountAsBytes)
}

// queryAccountByName - query data entry from chaincode state by publisher
///////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) queryAccountByName(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of account holder")
	}

	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	name := args[0]
	// Query the Name~AccountID index by publisher
	// This will execute a key range query on all keys starting with 'Publisher'
	nameIDResultsIterator, err := stub.GetStateByPartialCompositeKey("Name~AccountID", []string{name})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer nameIDResultsIterator.Close()

	// Iterate through result set
	var accountsAsBytes []byte
	var i int
	for i = 0; nameIDResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable).
		responseRange, err := nameIDResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// get the Name and AccountID from Name~AccountID composite key
		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		// returnedName := compositeKeyParts[0]
		returnedAccountID := compositeKeyParts[1]

		response := cc.getAccountByID(stub, []string{returnedAccountID})
		if response.Status != shim.OK {
			return shim.Error("Retrieval of account entry failed: " + response.Message)
		}
		accountsAsBytes = append(accountsAsBytes, response.Payload...)
		if nameIDResultsIterator.HasNext() {
			accountsAsBytes = append(accountsAsBytes, []byte(",")...)
		}
	}
	accountsAsBytes = append([]byte("["), accountsAsBytes...)
	accountsAsBytes = append(accountsAsBytes, []byte("]")...)
	// It returns results as JSON array
	return shim.Success(accountsAsBytes)
}

// sendTokensFast - transfer tokens from one account to another without check of sender's tokens
////////////////////////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) sendTokensFast(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	//       0              1            2          3
	// "fromAccountId" "toAccountId" "Amount" "dataPurchase"
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting FromAccountId, ToAccountId, Amount, dataPurchase")
	}
	// Input sanitation
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}

	fromAccountID := args[0]
	toAccountID := args[1]
	if fromAccountID == toAccountID {
		return shim.Error("From account and to account cannot be the same.")
	}
	tokensToSend, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return shim.Error("Expecting integer as number of tokens to transfer.")
	}
	if tokensToSend > limitTokens {
		return shim.Error("Exceeded max number of tokens for fast transaction. Use safe token transfer instead.")
	}

	dataPurchase, err := strconv.ParseBool(args[3])
	if err != nil {
		return shim.Error("Expecting boolean value. If this transfer is for data purchase or not.")
	}

	// Index txID and sender accounts ID
	// this is required for quick lookup and transaction aggregation.
	txID := stub.GetTxID()
	var txSenderIDCompositeKey, txRecipientIDCompositeKey string
	if dataPurchase {
		txSenderIDCompositeKey, err = stub.CreateCompositeKey("Account~op~Tok~TxID", []string{fromAccountID, "-", strconv.FormatInt(tokensToSend, 10), txID})
		if err != nil {
			return shim.Error(err.Error())
		}
		txRecipientIDCompositeKey, err = stub.CreateCompositeKey("Account~op~PendingTok~TxID", []string{toAccountID, "+", strconv.FormatInt(tokensToSend, 10), txID})
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		txSenderIDCompositeKey, err = stub.CreateCompositeKey("Account~op~Tok~TxID", []string{fromAccountID, "-", strconv.FormatInt(tokensToSend, 10), txID})
		if err != nil {
			return shim.Error(err.Error())
		}
		txRecipientIDCompositeKey, err = stub.CreateCompositeKey("Account~op~Tok~TxID", []string{toAccountID, "+", strconv.FormatInt(tokensToSend, 10), txID})
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	// Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
	// Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(txSenderIDCompositeKey, value)
	if err != nil {
		return shim.Error(err.Error())
	}

	stub.PutState(txRecipientIDCompositeKey, value)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Tx entry saved and indexed

	return shim.Success([]byte(txID))

}

// sendTokensSafe - transfer tokens from one account to another with check of sender's tokens
////////////////////////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) sendTokensSafe(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	//       0              1            2          3
	// "fromAccountId" "toAccountId" "Amount" "dataPurchase"
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting FromAccountId, ToAccountId, Amount")
	}
	// Input sanitation
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}

	fromAccountID := args[0]
	toAccountID := args[1]
	if fromAccountID == toAccountID {
		return shim.Error("From account and to account cannot be the same.")
	}
	tokensToSend, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return shim.Error("Expecting integer as number of tokens to transfer.")
	}

	dataPurchase, err := strconv.ParseBool(args[3])
	if err != nil {
		return shim.Error("Expecting boolean value. If this transfer is for data purchase or not.")
	}

	// If it does not fail then accounts exist
	// no need to unmarshal
	fromAccountAsBytes, err := stub.GetState(fromAccountID) //get the account entry from chaincode state
	if err != nil {
		return shim.Error(err.Error())
	} else if fromAccountAsBytes == nil {
		return shim.Error(err.Error())
	}
	toAccountAsBytes, err := stub.GetState(toAccountID)
	if err != nil {
		return shim.Error(err.Error())
	} else if toAccountAsBytes == nil {
		return shim.Error(err.Error())
	}

	// get the latest state of tokens for sender's account
	argsTok := []string{fromAccountID}
	fromAccTokResponse := cc.getAccountTokens(stub, argsTok)
	if fromAccTokResponse.Status != shim.OK {
		return shim.Error("Retrieval of account tokens failed: " + fromAccTokResponse.Message)
	}

	fromAccTokStr := string(fromAccTokResponse.Payload)
	// index 0 holds usable tokens and index 1 holds pending tokens ":" is delimiter
	fromAccTokArr := strings.Split(fromAccTokStr, ":")
	fromAccTok, err := strconv.ParseInt(fromAccTokArr[0], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	if fromAccTok < tokensToSend {
		return shim.Error("Not enough tokens on the sender's account")
	}

	// Index txID and sender accounts ID
	// this is required for quick lookup and transaction aggregation.
	txID := stub.GetTxID()
	var txSenderIDCompositeKey, txRecipientIDCompositeKey string
	if dataPurchase {
		txSenderIDCompositeKey, err = stub.CreateCompositeKey("Account~op~PendingTok~TxID", []string{fromAccountID, "-", strconv.FormatInt(tokensToSend, 10), txID})
		if err != nil {
			return shim.Error(err.Error())
		}
		txRecipientIDCompositeKey, err = stub.CreateCompositeKey("Account~op~PendingTok~TxID", []string{toAccountID, "+", strconv.FormatInt(tokensToSend, 10), txID})
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		txSenderIDCompositeKey, err = stub.CreateCompositeKey("Account~op~Tok~TxID", []string{fromAccountID, "-", strconv.FormatInt(tokensToSend, 10), txID})
		if err != nil {
			return shim.Error(err.Error())
		}
		txRecipientIDCompositeKey, err = stub.CreateCompositeKey("Account~op~Tok~TxID", []string{toAccountID, "+", strconv.FormatInt(tokensToSend, 10), txID})
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	// Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
	// Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(txSenderIDCompositeKey, value)
	if err != nil {
		return shim.Error(err.Error())
	}

	stub.PutState(txRecipientIDCompositeKey, value)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Tx entry saved and indexed

	return shim.Success([]byte(txID))
}

// updateAccountTokens - updates the account entry in state with the latest values of tokens and pending tokens
/////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) updateAccountTokens(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	//       0
	// "accountID"
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting account ID")
	}
	// Input sanitation
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	accountID := args[0]

	//get the account entry from chaincode state
	accountAsBytes, err := stub.GetState(accountID)
	if err != nil {
		return shim.Error(err.Error())
	} else if accountAsBytes == nil {
		return shim.Error(err.Error())
	}

	var account Account
	err = json.Unmarshal(accountAsBytes, &account)
	if err != nil {
		return shim.Error("Some error: " + err.Error())
	}

	// get the latest state of tokens for sender's account
	argsTok := []string{accountID}
	accTokResponse := cc.getAccountTokens(stub, argsTok)
	if accTokResponse.Status != shim.OK {
		return shim.Error("Retrieval of account tokens failed: " + accTokResponse.Message)
	}

	accTokStr := string(accTokResponse.Payload)
	// index 0 holds usable tokens and index 1 holds pending tokens ":" is delimiter
	accTokArr := strings.Split(accTokStr, ":")
	accTok, err := strconv.ParseInt(accTokArr[0], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	accPendingTok, err := strconv.ParseInt(accTokArr[1], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	// update values of tokens
	account.Tokens = accTok
	account.PendingTokens = accPendingTok

	// Marshal objects back
	accountAsBytesNew, err := json.Marshal(&account)
	if err != nil {
		return shim.Error("Some error: " + err.Error())
	}
	// Write state back to the ledger
	err = stub.PutState(accountID, accountAsBytesNew)
	if err != nil {
		return shim.Error("Some error: " + err.Error())
	}
	// return JSON object Account with updated token values
	return shim.Success(accountAsBytesNew)
}

// getAccountTokens - returns current state of tokens in a specific account
////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) getAccountTokens(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	//    0
	// "accountID"

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting account ID")
	}
	// Input sanitation
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	accountID := args[0]
	// Get all account transactions for the account ID
	accountTxIterator, err := stub.GetStateByPartialCompositeKey("Account~op~Tok~TxID", []string{accountID})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer accountTxIterator.Close()

	// Iterate through result set and compute final amount of tokens
	var finalTok int64
	for i := 0; accountTxIterator.HasNext(); i++ {
		// Get the next row
		responseRange, err := accountTxIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// Split the composite key into its component parts
		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		// Retrieve the amount of tokens and operation
		operation := compositeKeyParts[1]
		tokensStr := compositeKeyParts[2]

		// Convert the tokensStr string and perform the operation
		tokens, err := strconv.ParseInt(tokensStr, 10, 64)
		if err != nil {
			return shim.Error(err.Error())
		}
		// calculate the delta
		switch operation {
		case "+":
			finalTok += tokens
		case "-":
			finalTok -= tokens
		default:
			return shim.Error(fmt.Sprintf("Unrecognized operation %s", operation))
		}
	}

	// Get all account pending transactions for the account ID (for data purchase that did not happen yet)
	accountPendingTxIterator, err := stub.GetStateByPartialCompositeKey("Account~op~PendingTok~TxID", []string{accountID})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer accountPendingTxIterator.Close()

	// Iterate through result set and compute final amount of tokens
	var finalPendingTok int64
	for i := 0; accountPendingTxIterator.HasNext(); i++ {
		// Get the next row
		responseRange2, err := accountPendingTxIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// Split the composite key into its component parts
		_, compositeKeyParts2, err := stub.SplitCompositeKey(responseRange2.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		// Retrieve the amount of tokens and operation
		pendingOperation := compositeKeyParts2[1]
		PendingTokensStr := compositeKeyParts2[2]

		// Convert the tokensStr string and perform the operation
		pendingTokens, err := strconv.ParseInt(PendingTokensStr, 10, 64)
		if err != nil {
			return shim.Error(err.Error())
		}
		// calculate the delta
		switch pendingOperation {
		case "+":
			finalPendingTok += pendingTokens
		case "-":
			finalPendingTok -= pendingTokens
		default:
			return shim.Error(fmt.Sprintf("Unrecognized operation %s", pendingOperation))
		}
	}
	// format int64 to string and separate tokens and pending tokens with ":"
	res := fmt.Sprint(finalTok) + ":" + fmt.Sprint(finalPendingTok)
	return shim.Success([]byte(res))
}

// getAccountHistoryByID - get the whole history of specific account number even if it was deleted from state.
/////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) getAccountHistoryByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//    0
	// "accountID"

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting AccountID")
	}
	// Input sanitation
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	accountID := args[0]

	resultsIterator, err := stub.GetHistoryForKey(accountID)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the account
	var buffer bytes.Buffer
	buffer.WriteString("[")

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we set the
		// value to null. Else, we will write the response.Value
		// as-is (as the Value itself a JSON)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		// was it delete transaction?
		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		// Add a comma in front of an array member
		if resultsIterator.HasNext() {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}

/*
// getTxParticipants - returns recepient's account ID of transaction
//////////////////////////////////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) getTxParticipants(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	//    0
	// "txID"

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting TxID")
	}
	// Input sanitation
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	txID := args[0]
	txIDResultsIterator, err := stub.GetStateByPartialCompositeKey("TxID~Sender~Recipient", []string{txID})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer txIDResultsIterator.Close()

	if !txIDResultsIterator.HasNext() {
		return shim.Error("Transaction was not found.")
	}
	responseRange, err := txIDResultsIterator.Next()
	if err != nil {
		return shim.Error(err.Error())
	}
	// get the recipient account ID
	_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
	if err != nil {
		return shim.Error(err.Error())
	}
	participantsAccountsID := compositeKeyParts[1] + "->" + compositeKeyParts[2]

	if txIDResultsIterator.HasNext() {
		return shim.Error("Two TxID are same? Impossible!")
	}

	return shim.Success([]byte(participantsAccountsID))

}
*/
/*
// checkUsedTx - check if specific transaction is already used. If not then it adds it to the index
////////////////////////////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) addTxAsUsed(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	//    0
	// "txID"

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting TxID")
	}
	// Input sanitation
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	indexName := "TxID~Used"
	txID := args[0]
	txIDResultsIterator, err := stub.GetStateByPartialCompositeKey(indexName, []string{txID})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer txIDResultsIterator.Close()

	if txIDResultsIterator.HasNext() {
		return shim.Error("Transaction was already used for data purchase.")
	}

	txIDUsedIndexKey, err := stub.CreateCompositeKey(indexName, []string{txID, "TRUE"})
	if err != nil {
		return shim.Error(err.Error())
	}
	// Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
	// Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(txIDUsedIndexKey, value)
	// txId entry saved and indexed

	return shim.Success([]byte("Added as used Tx"))
}
*/
