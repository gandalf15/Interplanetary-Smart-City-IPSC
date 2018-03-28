package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"

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
	AccountId  string // unique id of the account
	Name       string // name of the account holder
	Tokens     int64  // amount of tokens (money)
}

/////////////////
// Main function
/////////////////
func main() {
	// increase max CPU
	runtime.GOMAXPROCS(runtime.NumCPU())
	err := shim.Start(new(Chaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

/////////////////////////////////////////////////////////////////////////////////
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
		err = stub.PutState(accounts[i].AccountId, accountJSONasBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	//  Index the account to enable name-based range queries
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	indexName := "Name~AccountId"
	var nameIdIndexKey string
	for i := 0; i < noOfAccounts; i++ {
		nameIdIndexKey, err = stub.CreateCompositeKey(indexName, []string{accounts[i].Name, accounts[i].AccountId})
		if err != nil {
			return shim.Error(err.Error())
		}
		//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
		//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
		value := []byte{0x00}
		stub.PutState(nameIdIndexKey, value)
	}

	return shim.Success(nil)
}

////////////////////////////////////////////
// Invoke - Our entry point for Invocations
////////////////////////////////////////////
func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "createAccount" { //create a new account
		return cc.createAccount(stub, args)
		/*
			} else if function == "deleteAccountById" { // delete an account by account Id
				return cc.deleteAccountById(stub, args)
		*/
	} else if function == "getAccountById" { // get an account by its Id
		return cc.getAccountById(stub, args)
	} else if function == "queryAccountByName" { // find an account base on name of account holder
		return cc.queryAccountByName(stub, args)
	} else if function == "transferTokens" { // transfer tokens from one account to another
		return cc.transferTokens(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

///////////////////////////////////////////////////////////////////////
// createAccount - create a new account and store into chaincode state
///////////////////////////////////////////////////////////////////////
func (cc *Chaincode) createAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//      0         1
	// "AccountId", "Name"
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

	accountId := args[0]
	name := args[1]

	// Check if an account already exists
	accountAsBytes, err := stub.GetState(accountId)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	} else if accountAsBytes != nil {
		fmt.Println("This account already exists: " + accountId)
		return shim.Error("This account already exists: " + accountId)
	}

	// Create Account object and marshal to JSON
	recordType := "ACCOUNT"
	accountEntry := &Account{recordType, accountId, name, 0}
	accountEntryJSONasBytes, err := json.Marshal(accountEntry)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Save account entry to state
	err = stub.PutState(accountId, accountEntryJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Index the account to enable name-based range queries
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	indexName := "Name~AccountId"
	nameIdIndexKey, err := stub.CreateCompositeKey(indexName, []string{accountEntry.Name, accountEntry.AccountId})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(nameIdIndexKey, value)

	// Account saved and indexed. Return success
	return shim.Success([]byte("Account created"))
}

/*
func (cc *Chaincode) deleteAccountById(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Success(nil)
}
*/

////////////////////////////////////////////////////////////////////////////
// getAccountById - read account entry from chaincode state based on its Id
////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) getAccountById(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting account Id")
	}

	accountId := args[0]
	accountAsBytes, err := stub.GetState(accountId) //get the account entry from chaincode state
	if err != nil {
		return shim.Error(err.Error())
	} else if accountAsBytes == nil {
		return shim.Error(err.Error())
	}

	return shim.Success(accountAsBytes)
}

///////////////////////////////////////////////////////////////////////////
// queryAccountByName - query data entry from chaincode state by publisher
///////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) queryAccountByName(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of account holder")
	}

	name := args[0]
	// Query the Name~AccountId index by publisher
	// This will execute a key range query on all keys starting with 'Publisher'
	nameIdResultsIterator, err := stub.GetStateByPartialCompositeKey("Name~AccountId", []string{name})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer nameIdResultsIterator.Close()

	// Iterate through result set
	var accountsAsBytes []byte
	var i int
	for i = 0; nameIdResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable).
		responseRange, err := nameIdResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// get the Name and AccountId from Name~AccountId composite key
		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		// returnedName := compositeKeyParts[0]
		returnedAccountId := compositeKeyParts[1]

		response := cc.getAccountById(stub, []string{returnedAccountId})
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

	fromAccountId := args[0]
	toAccountId := args[1]
	amount, err := strconv.Atoi(args[2])
	tokens := int64(amount)
	if err != nil {
		return shim.Error("Expecting integer as number of tokens to transfer.")
	}

	fromAccountAsBytes, err := stub.GetState(fromAccountId) //get the account entry from chaincode state
	if err != nil {
		return shim.Error(err.Error())
	} else if fromAccountAsBytes == nil {
		return shim.Error(err.Error())
	}

	toAccountAsBytes, err := stub.GetState(toAccountId)
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
	err = stub.PutState(fromAccountId, fromAccountAsBytesNew)
	if err != nil {
		return shim.Error("Some error: " + err.Error())
	}
	err = stub.PutState(toAccountId, toAccountAsBytesNew)
	if err != nil {
		return shim.Error("Some error: " + err.Error())
	}

	return shim.Success(nil)

}
