package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
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
	RecordType string // RecordType is used to distinguish the various types of objects in state database
	AccountID  string // unique id of the account
	Name       string // name of the account holder
	Tokens     int64  // amount of tokens (money)
}

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

	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error(`Incorect number of arguments.
			Expectiong number of accounts and tokens for each account to create`)
	}
	noOfAccounts, err := strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("Expecting integer as number of accounts to create.")
	}
	tokens, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer as number of tokens to init.")
	}
	accounts := make([]*Account, noOfAccounts)
	for i := 0; i < noOfAccounts; i++ {
		accounts[i] = &Account{"ACCOUNT", strconv.Itoa(i + 1), "Init_Account", int64(tokens)}
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
	indexName := "Name~AccountID"
	var nameIDIndexKey string
	for i := 0; i < noOfAccounts; i++ {
		nameIDIndexKey, err = stub.CreateCompositeKey(indexName, []string{accounts[i].Name, accounts[i].AccountID})
		if err != nil {
			return shim.Error(err.Error())
		}
		//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
		//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
		value := []byte{0x00}
		stub.PutState(nameIDIndexKey, value)
	}

	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
////////////////////////////////////////////
func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "createAccount" { //create a new account
		return cc.createAccount(stub, args)
	} else if function == "deleteAccountByID" { // delete an account by account Id
		return cc.deleteAccountByID(stub, args)
	} else if function == "getAccountByID" { // get an account by its Id
		return cc.getAccountByID(stub, args)
	} else if function == "getHistoryForAccount" { // get history for an account by its Id
		return cc.getHistoryForAccount(stub, args)
	} else if function == "queryAccountByName" { // find an account base on name of account holder
		return cc.queryAccountByName(stub, args)
	} else if function == "transferTokens" { // transfer tokens from one account to another
		return cc.transferTokens(stub, args)
	} else if function == "getRecipientTx" { // transfer tokens from one account to another
		return cc.getRecipientTx(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
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
		fmt.Println("This account already exists: " + accountID)
		return shim.Error("This account already exists: " + accountID)
	}

	// Create Account object and marshal to JSON
	recordType := "ACCOUNT"
	accountEntry := &Account{recordType, accountID, name, 0}
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

// deleteAccountByID - deletes the account if number of tokens is 0
////////////////////////////////////////////////////////////////////
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
	if account.Tokens == 0 {
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

	return shim.Error("Account cannot be deleted. Amount of tokens is not 0.")

}

// getAccountByID - read account entry from chaincode state based on its Id
////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) getAccountByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting account Id")
	}

	accountID := args[0]
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
		accountsAsBytes = append(accountsAsBytes, []byte("\n")...)
	}
	summary := fmt.Sprintf("\nFound %d accounts with name %s", i, name)
	accountsAsBytes = append(accountsAsBytes, summary...)
	return shim.Success(accountsAsBytes)
}

////////////////////////////////////////////////////////////////
// transferTokens - transfer tokens from one account to another
////////////////////////////////////////////////////////////////

func (cc *Chaincode) transferTokens(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	//       0              1            2
	// FromAccountId    ToAccountId    Amount
	if len(args) != 3 {
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

	fromAccountID := args[0]
	toAccountID := args[1]
	if fromAccountID == toAccountID {
		return shim.Error("From account and to account cannot be the same.")
	}
	amount, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Expecting integer as number of tokens to transfer.")
	}
	tokens := int64(amount)

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

	var fromAccount, toAccount Account
	err = json.Unmarshal(fromAccountAsBytes, &fromAccount)
	if err != nil {
		return shim.Error("Some error: " + err.Error())
	}

	err = json.Unmarshal(toAccountAsBytes, &toAccount)
	if err != nil {
		return shim.Error("Some error: " + err.Error())
	}
	if fromAccount.Tokens < tokens {
		return shim.Error("Account does not have sufficient amount of tokens.")
	}

	fromAccount.Tokens -= tokens
	toAccount.Tokens += tokens

	// Marshal objects back
	fromAccountAsBytesNew, err := json.Marshal(&fromAccount)
	if err != nil {
		return shim.Error("Some error: " + err.Error())
	}
	toAccountAsBytesNew, err := json.Marshal(&toAccount)
	if err != nil {
		return shim.Error("Some error: " + err.Error())
	}
	// Write state back to the ledger
	err = stub.PutState(fromAccountID, fromAccountAsBytesNew)
	if err != nil {
		return shim.Error("Some error: " + err.Error())
	}
	err = stub.PutState(toAccountID, toAccountAsBytesNew)
	if err != nil {
		return shim.Error("Some error: " + err.Error())
	}

	// Index txID and recepient's account ID
	// this is required for quick lookup and validation if txID was already used.
	txID := stub.GetTxID()
	txIDIndexKey, err := stub.CreateCompositeKey("TxID~RecipientAccountID", []string{txID, toAccountID})
	if err != nil {
		return shim.Error(err.Error())
	}
	// Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
	// Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(txIDIndexKey, value)
	// TxId entry saved and indexed
	return shim.Success([]byte(txID))

}

// getHistoryForAccount - get the whole history of specific account number even if it was deleted from state.
/////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) getHistoryForAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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

// getRecipientTx - returns recepient's account ID of transaction
//////////////////////////////////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) getRecipientTx(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	txIDResultsIterator, err := stub.GetStateByPartialCompositeKey("TxID~RecipientAccountID", []string{txID})
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
	recipientAccountID := compositeKeyParts[1]

	if txIDResultsIterator.HasNext() {
		return shim.Error("Two TxID are same? Impossible!")
	}

	return shim.Success([]byte(recipientAccountID))

}
