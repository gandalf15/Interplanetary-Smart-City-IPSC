package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode implements Chaincode interface
type SimpleChaincode struct {
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

// ===================================================================================
// Main
// ===================================================================================
func main() {
	// increase max CPU
	runtime.GOMAXPROCS(runtime.NumCPU())
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ===========================
// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// ========================================
// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "putDataEntry" { //create a new data entry
		return t.putDataEntry(stub, args)
	} else if function == "getDataEntryById" { //read specific data by DataEntryId
		return t.getDataEntryById(stub, args)
	} else if function == "queryDataEntryByPublisher" { //find data created by publisher using rich query
		return t.queryDataEntryByPublisher(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ===================================================================
// putDataEntry - create a new data entry, store into chaincode state
// ===================================================================
func (t *SimpleChaincode) putDataEntry(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//        0           1             2        3           4             5
	// "DataEntryId", "Description", "Value", "Unit", "CreationTime", "Publisher"
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	// ==== Input sanitation ====
	fmt.Println("- start create user")
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
	description := strings.ToLower(args[1])
	value := strings.ToLower(args[2])
	unit := strings.ToLower(args[3])
	creationTime := strings.ToLower(args[4])
	publisher := strings.ToLower(args[5])

	// ==== Check if data entry already exists ====
	dataEntryAsBytes, err := stub.GetState(dataEntryId)
	if err != nil {
		return shim.Error("Failed to get data entry: " + err.Error())
	} else if dataEntryAsBytes != nil {
		fmt.Println("This data entry already exists: " + dataEntryId)
		return shim.Error("This data entry already exists: " + dataEntryId)
	}

	// ==== Create data entry object and marshal to JSON ====
	recordType := "DATA_ENTRY"
	dataEntry := &DataEntry{recordType, dataEntryId, description, value, unit, creationTime, publisher}
	dataEntryJSONasBytes, err := json.Marshal(dataEntry)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save data entry to state ===
	err = stub.PutState(dataEntryId, dataEntryJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the data to enable publisher-based range queries ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	indexName := "Publisher~DataEntryId"
	pubIdIndexKey, err := stub.CreateCompositeKey(indexName, []string{dataEntry.Publisher, dataEntry.DataEntryId})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value1 := []byte{0x00}
	stub.PutState(pubIdIndexKey, value1)

	// ==== data entry saved and indexed. Return success ====
	fmt.Println("- end create data entry")
	return shim.Success(nil)
}

// ====================================================================
// getDataEntryById - read data entry from chaincode state based its Id
// ====================================================================
func (t *SimpleChaincode) getDataEntryById(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting data entry Id to query")
	}

	dataEntryId := args[0]
	valAsbytes, err := stub.GetState(dataEntryId) //get the data entry from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + dataEntryId + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Data entry does not exist: " + dataEntryId + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ===============================================
// queryDataEntryByPublisher - query data entry from chaincode state by publisher
// ===============================================
func (t *SimpleChaincode) queryDataEntryByPublisher(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting publisher to query")
	}

	publisher := strings.ToLower(args[0])
	// Query the Publisher~DataEntryId index by publisher
	// This will execute a key range query on all keys starting with 'Publisher'
	pubIdResultsIterator, err := stub.GetStateByPartialCompositeKey("Publisher~DataEntryId", []string{publisher})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer pubIdResultsIterator.Close()

	// Iterate through result set and for each DataEntry found
	var i int
	var valAsbytes []byte
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

		response := t.getDataEntryById(stub, []string{returnedDataEntryId})
		if response.Status != shim.OK {
			return shim.Error("Retrieval of data entry failed: " + response.Message)
		}
		valAsbytes = append(valAsbytes, response.Payload...)
		valAsbytes = append(valAsbytes, []byte(string('\n'))...)
	}
	responsePayload := fmt.Sprintf("\nFound %d data entry from publisher %s", i, publisher)
	fmt.Println("- end queryDataEntryByPublisher: " + responsePayload)
	return shim.Success(append(valAsbytes, []byte(responsePayload)...))
}
