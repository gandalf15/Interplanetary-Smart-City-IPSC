package main

import (
	"encoding/json"
	"fmt"
	"runtime"

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
	DataEntryId  string // unique id of the entry
	Description  string // human readable description
	Value        string // data value
	Unit         string // optional units for the data value
	CreationTime string // Time when the data was created. It can differ from the blockchain entry time
	Publisher    string // publisher of the data
}

//////////
// Main
//////////
func main() {
	// increase max CPU
	runtime.GOMAXPROCS(runtime.NumCPU())
	err := shim.Start(new(Chaincode))
	if err != nil {
		shim.Error(err.Error())
	}
}

//////////////////////////////
// Init initializes chaincode
//////////////////////////////
func (cc *Chaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

////////////////////////////////////////////
// Invoke - Our entry point for Invocations
////////////////////////////////////////////
func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	// Handle functions
	if function == "createDataEntry" { //create a new data entry
		return cc.createDataEntry(stub, args)
	} else if function == "getDataEntryById" { //read specific data by DataEntryId
		return cc.getDataEntryById(stub, args)
	} else if function == "queryDataEntryByPublisher" { //find data created by publisher using compound key
		return cc.queryDataEntryByPublisher(stub, args)
	} else if function == "revealDataValue" { // invoke other chaincode and reveal values
		return cc.revealDataValue(stub, args)
	}

	return shim.Error("Received unknown function invocation")
}

/////////////////////////////////////////////////////////////////////////
// createDataEntry - create a new data entry, store into chaincode state
/////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) createDataEntry(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//        0           1             2        3           4             5
	// "DataEntryId", "Description", "Value", "Unit", "CreationTime", "Publisher"
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
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
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6th argument must be a non-empty string")
	}

	dataEntryId := args[0]
	description := args[1]
	value := args[2]
	unit := args[3]
	creationTime := args[4]
	publisher := args[5]

	// ==== Check if data entry already exists ====
	dataEntryAsBytes, err := stub.GetState(dataEntryId)
	if err != nil {
		return shim.Error("Failed to get data entry: " + err.Error())
	} else if dataEntryAsBytes != nil {
		return shim.Error("This data entry already exists: " + dataEntryId)
	}

	// Create data entry object and marshal to JSON
	recordType := "DATA_ENTRY"
	dataEntry := &DataEntry{recordType, dataEntryId, description, value, unit, creationTime, publisher}
	dataEntryJSONasBytes, err := json.Marshal(dataEntry)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Save data entry to state
	err = stub.PutState(dataEntryId, dataEntryJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Index the data to enable publisher-based range queries
	// An 'index' is a normal key/value entry in state.
	// The key is a composite key, with the elements that you want to range query on listed first.
	indexName := "Publisher~DataEntryId"
	pubIdIndexKey, err := stub.CreateCompositeKey(indexName, []string{dataEntry.Publisher, dataEntry.DataEntryId})
	if err != nil {
		return shim.Error(err.Error())
	}
	// Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
	// Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value1 := []byte{0x00}
	stub.PutState(pubIdIndexKey, value1)

	// Data entry saved and indexed
	return shim.Success([]byte("Data entry is created"))
}

////////////////////////////////////////////////////////////////////////
// getDataEntryById - read data entry from chaincode state based its Id
////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) getDataEntryById(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting data entry Id to query")
	}

	dataEntryId := args[0]
	dataAsBytes, err := stub.GetState(dataEntryId) //get the data entry from chaincode state
	if err != nil {
		return shim.Error(err.Error())
	} else if dataAsBytes == nil {
		return shim.Error(err.Error())
	}

	return shim.Success(dataAsBytes)
}

//////////////////////////////////////////////////////////////////////////////////
// queryDataEntryByPublisher - query data entry from chaincode state by publisher
//////////////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) queryDataEntryByPublisher(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting publisher name")
	}

	publisher := args[0]
	// Query the Publisher~DataEntryId index by publisher
	// This will execute a key range query on all keys starting with 'Publisher'
	pubIdResultsIterator, err := stub.GetStateByPartialCompositeKey("Publisher~DataEntryId", []string{publisher})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer pubIdResultsIterator.Close()

	// Iterate through result set and for each DataEntry found
	var dataAsBytes []byte
	var i int
	for i = 0; pubIdResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the DataEntryId name from the composite key
		responseRange, err := pubIdResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// get the publisher and dataEntryId from Publisher~DataEntryId composite key
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedPublisher := compositeKeyParts[0]
		returnedDataEntryId := compositeKeyParts[1]
		fmt.Printf("- found a data entry from index:%s Publisher:%s DataEntryId:%s\n", objectType, returnedPublisher, returnedDataEntryId)

		response := cc.getDataEntryById(stub, []string{returnedDataEntryId})
		if response.Status != shim.OK {
			return shim.Error("Retrieval of data entry failed: " + response.Message)
		}
		dataAsBytes = append(dataAsBytes, response.Payload...)
		dataAsBytes = append(dataAsBytes, []byte("\n")...)
	}
	summary := fmt.Sprintf("\nFound %d data entry from publisher %s", i, publisher)
	dataAsBytes = append(dataAsBytes, summary...)
	return shim.Success(dataAsBytes)
}

////////////////////////////////////////////////////////////
// invokeChaincode - invokes chaincode in different channel
////////////////////////////////////////////////////////////
func (cc *Chaincode) revealDataValue(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	var dataEntry DataEntry

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting chaincode name, data entry Id, channel name")
	}
	chaincodeName := args[0]
	dataEntryId := args[1]
	channel := args[2]
	f := []byte("getDataEntryById")
	argsToChaincode := [][]byte{f, []byte(dataEntryId)}

	// check if the dataEntryId is present in channel2 ledger
	dataAsBytes, err := stub.GetState(dataEntryId) //get the data entry from chaincode state
	if err != nil {
		return shim.Error(err.Error())
	} else if dataAsBytes == nil {
		return shim.Error(err.Error())
	}

	err = json.Unmarshal(dataAsBytes, &dataEntry)
	if err != nil {
		return shim.Error(err.Error())
	}

	response := stub.InvokeChaincode(chaincodeName, argsToChaincode, channel)
	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(response.Payload))
		return shim.Error(errStr)
	}

	err = stub.PutState(dataEntryId, response.Payload)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Data are already indexed.

	return response
}
