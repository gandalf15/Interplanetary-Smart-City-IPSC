package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Chaincode implements Chaincode interface
type Chaincode struct {
}

// Variable names in a struct must be capitalised. Otherwise they are not exported (also to JSON)

// DataEntry represents data created on IoT device
type DataEntry struct {
	RecordType   string // RecordType is used to distinguish the various types of objects in state database
	DataEntryID  string // unique id of the entry
	Description  string // human readable description
	Value        string // data value
	Unit         string // optional units for the data value
	CreationTime string // Time when the data was created. It can differ from the blockchain entry time
	Publisher    string // publisher of the data
}

// DataEntryAd - represents data created by publisher and advertised for specific price
type DataEntryAd struct {
	DataEntry        // anonymous field
	Price     int64  // Price for data value
	AccountNo string // account number where to transfer tokens
}

// Main
//////////
func main() {
	// increase max CPU
	// runtime.GOMAXPROCS(runtime.NumCPU())
	err := shim.Start(new(Chaincode))
	if err != nil {
		shim.Error(err.Error())
	}
}

// Init initializes chaincode
//////////////////////////////
func (cc *Chaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
////////////////////////////////////////////
func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	// Handle functions
	if function == "createDataEntryAd" { //create a new data entry
		return cc.createDataEntryAd(stub, args)
	} else if function == "getDataAdByID" { //read specific data by DataEntryID
		return cc.getDataAdByID(stub, args)
	} else if function == "queryDataByPub" { //find data created by publisher using compound key
		return cc.queryDataByPub(stub, args)
	} else if function == "revealFreeData" { // invoke other chaincode and reveal values
		return cc.revealFreeData(stub, args)
	} else if function == "revealPaidData" { // invoke other chaincode and reveal values
		return cc.revealPaidData(stub, args)
	}

	return shim.Error("Received unknown function invocation")
}

// createDataEntryAd - create a new data entry, store into chaincode state
/////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) createDataEntryAd(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//        0           1             2        3           4             5         6        	7
	// "DataEntryID", "Description", "Value", "Unit", "CreationTime", "Publisher", "Price", "AccountNo"
	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments. Expecting 8")
	}

	// Input sanitization
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
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6th argument must be a non-empty string")
	}
	if len(args[6]) <= 0 {
		return shim.Error("7th argument must be a non-empty string")
	}
	if len(args[7]) <= 0 {
		return shim.Error("8th argument must be a non-empty string")
	}

	dataEntryID := args[0]
	description := args[1]
	value := args[2]
	unit := args[3]
	creationTime := args[4]
	publisher := args[5]
	priceInt, err := strconv.Atoi(args[6])
	accountNo := args[7]
	if err != nil {
		return shim.Error("Expecting integer as price for the data entry.")
	}
	price := int64(priceInt)
	// ==== Check if data entry already exists ====
	dataEntryAdAsBytes, err := stub.GetState(dataEntryID)
	if err != nil {
		return shim.Error("Failed to get data entry: " + err.Error())
	} else if dataEntryAdAsBytes != nil {
		return shim.Error("This data entry Ad already exists: " + dataEntryID)
	}

	// Check if price is positive number
	if price < 0 {
		return shim.Error("Price cannot be negative number.")
	}

	// Create data entry object and marshal to JSON
	recordType := "DATA_ENTRY_AD"
	dataEntryAd := &DataEntryAd{DataEntry{recordType, dataEntryID, description, value, unit, creationTime, publisher}, price, accountNo}
	dataEntryAdJSONasBytes, err := json.Marshal(dataEntryAd)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Save data entry to state
	err = stub.PutState(dataEntryID, dataEntryAdJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Index the data to enable publisher-based range queries
	// An 'index' is a normal key/value entry in state.
	// The key is a composite key, with the elements that you want to range query on listed first.
	indexName := "Publisher~DataEntryID"
	pubIDIndexKey, err := stub.CreateCompositeKey(indexName, []string{dataEntryAd.Publisher, dataEntryAd.DataEntryID})
	if err != nil {
		return shim.Error(err.Error())
	}
	// Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
	// Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value1 := []byte{0x00}
	stub.PutState(pubIDIndexKey, value1)

	// Data entry saved and indexed
	return shim.Success([]byte("Data entry is created"))
}

// getDataAdByID - read data entry from chaincode state based its Id
////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) getDataAdByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	//        0
	// "DataEntryID"
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting data entry Id to query")
	}

	// Input sanitization
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	dataEntryID := args[0]
	dataAsBytes, err := stub.GetState(dataEntryID) //get the data entry from chaincode state
	if err != nil {
		return shim.Error(err.Error())
	} else if dataAsBytes == nil {
		return shim.Error(err.Error())
	}

	return shim.Success(dataAsBytes)
}

// queryDataByPub - query data entry from chaincode state by publisher
//////////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) queryDataByPub(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	//     0
	// "Publisher"

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting publisher name")
	}

	// Input sanitization
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	publisher := args[0]
	// Query the Publisher~DataEntryID index by publisher
	// This will execute a key range query on all keys starting with 'Publisher'
	pubIDResultsIterator, err := stub.GetStateByPartialCompositeKey("Publisher~DataEntryID", []string{publisher})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer pubIDResultsIterator.Close()

	// Iterate through result set and for each DataEntryAd found
	var dataAsBytes []byte
	var i int
	for i = 0; pubIDResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the DataEntryID name from the composite key
		responseRange, err := pubIDResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// get the publisher and dataEntryID from Publisher~DataEntryID composite key
		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedDataEntryID := compositeKeyParts[1]

		response := cc.getDataAdByID(stub, []string{returnedDataEntryID})
		if response.Status != shim.OK {
			return shim.Error("Retrieval of data entry failed: " + response.Message)
		}
		dataAsBytes = append(dataAsBytes, response.Payload...)
		if pubIDResultsIterator.HasNext() {
			dataAsBytes = append(dataAsBytes, []byte(",")...)
		}
	}
	dataAsBytes = append([]byte("["), dataAsBytes...)
	dataAsBytes = append(dataAsBytes, []byte("]")...)
	// It returns results as JSON array
	return shim.Success(dataAsBytes)
}

// revealFreeData - invokes chaincode in different channel and reveal its value
////////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) revealFreeData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//		0				  1			   2
	// "chaincodeName", "dataEntryID", "channelName"
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting chaincode name, data entry Id, channel name")
	}
	chaincodeName := args[0]
	dataEntryID := args[1]
	channel := args[2]
	f := []byte("getDataByID")
	argsToChaincode := [][]byte{f, []byte(dataEntryID)}

	// check if the dataEntryID is present in channel2 ledger
	dataAsBytes, err := stub.GetState(dataEntryID) //get the data entry from chaincode state
	if err != nil {
		return shim.Error(err.Error())
	} else if dataAsBytes == nil {
		return shim.Error(err.Error())
	}

	var dataEntryAd DataEntryAd
	err = json.Unmarshal(dataAsBytes, &dataEntryAd)
	if err != nil {
		return shim.Error(err.Error())
	}

	response := stub.InvokeChaincode(chaincodeName, argsToChaincode, channel)
	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(response.Message))
		return shim.Error(errStr)
	}

	var dataEntry DataEntry
	err = json.Unmarshal(response.Payload, &dataEntry)
	if err != nil {
		return shim.Error(err.Error())
	}

	dataEntryAd.Value = dataEntry.Value
	dataEntryAdAsBytes, err := json.Marshal(dataEntryAd)
	if err != nil {
		return shim.Error(err.Error())
	}
	// write state
	err = stub.PutState(dataEntryID, dataEntryAdAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	// No need to index. DataEntryAd already indexed.

	return shim.Success(dataEntryAdAsBytes)
}

// revealPaidData - invokes chaincode in different channel. Data entry
//                   is paid, first check transaction.
///////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) revealPaidData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	//			0				1				2					3				4			5
	// "chaincodeDataName", "dataEntryID", "channelData", "chaincodeTokensName", "txID", "channelTokens"
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	// Input sanitization
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
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6th argument must be a non-empty string")
	}

	chaincodeDataName := args[0]
	dataEntryID := args[1]
	channelData := args[2]
	chaincodeTokensName := args[3]
	txID := args[4]
	channelTokens := args[5]

	// check if the dataEntryID is present in channel2 ledger
	dataAsBytes, err := stub.GetState(dataEntryID) //get the data entry from chaincode state
	if err != nil {
		return shim.Error(err.Error())
	} else if dataAsBytes == nil {
		return shim.Error(err.Error())
	}
	// unmarshal
	var dataEntryAd DataEntryAd
	err = json.Unmarshal(dataAsBytes, &dataEntryAd)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Check if data entry is paid.
	if dataEntryAd.Price == 0 {
		return shim.Error("Data entry is free. Wrong function call.")
	}

	// Check if the txID is already in state used for some data entry purchase.
	// If not then add and index it as used transaction
	indexName := "Tx~DataEntryID"
	txIDResultsIterator, err := stub.GetStateByPartialCompositeKey(indexName, []string{txID})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer txIDResultsIterator.Close()

	if txIDResultsIterator.HasNext() {
		return shim.Error("Transaction was already used for data purchase.")
	}
	// it only indexes if this transaction is commited. Atomicity...
	// therefore this statement does not have to be at the end.
	txIDIndexKey, err := stub.CreateCompositeKey(indexName, []string{txID, dataEntryAd.DataEntryID})
	if err != nil {
		return shim.Error(err.Error())
	}
	// Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
	// Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(txIDIndexKey, value)
	// txId entry saved and indexed

	// TODO: change this for return value for both participants IDs
	// invoke chaincode and get the recipient of Tx
	fTokens := []byte("getTxDetails")
	argsToChaincode := [][]byte{fTokens, []byte(txID)}
	responseParticipantsIDs := stub.InvokeChaincode(chaincodeTokensName, argsToChaincode, channelTokens)
	if responseParticipantsIDs.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(responseParticipantsIDs.Message))
		return shim.Error(errStr)
	}

	// Check if recipient of the Tx is the data entry account No.
	participantsAccIDs := strings.Split(string(responseParticipantsIDs.Payload), "->")
	recipientAccID := participantsAccIDs[1]
	if recipientAccID != dataEntryAd.AccountNo {
		return shim.Error("This transaction does not have the same recipient account ID as required by data entry ad.")
	}

	// invoke chaincode in channel where data entry with value is
	// this prevent from indexing TxID as used if data entry is not present on another channel
	fData := []byte("getDataByID")
	argsToChaincode = [][]byte{fData, []byte(dataEntryID)}
	responseData := stub.InvokeChaincode(chaincodeDataName, argsToChaincode, channelData)
	if responseData.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(responseData.Message))
		return shim.Error(errStr)
	}

	// at this stage we know that Tx recipient is correct and data entry present
	// invoke chaincode and check/add Tx to the central index as spent
	fTokens = []byte("addTxAsUsed")
	argsToChaincode = [][]byte{fTokens, []byte(txID)}
	responseTxSpent := stub.InvokeChaincode(chaincodeTokensName, argsToChaincode, channelTokens)
	if responseTxSpent.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(responseTxSpent.Message))
		return shim.Error(errStr)
	}

	// unmarshal data entry
	var dataEntry DataEntry
	err = json.Unmarshal(responseData.Payload, &dataEntry)
	if err != nil {
		return shim.Error(err.Error())
	}

	dataEntryAd.Value = dataEntry.Value
	dataEntryAdAsBytes, err := json.Marshal(dataEntryAd)
	if err != nil {
		return shim.Error(err.Error())
	}
	// write state
	err = stub.PutState(dataEntryID, dataEntryAdAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	// No need to index. DataEntryAd already indexed.

	return shim.Success(dataEntryAdAsBytes)

}
