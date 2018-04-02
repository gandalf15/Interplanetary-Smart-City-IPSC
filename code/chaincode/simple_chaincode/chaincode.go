package main

import (
	"encoding/json"
	"fmt"

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

// Main
////////
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

	// Handle different functions
	if function == "createData" { //create a new data entry
		return cc.createData(stub, args)
	} else if function == "getDataByID" { //read specific data by DataEntryID
		return cc.getDataByID(stub, args)
	} else if function == "queryDataByPub" { //find data created by publisher using rich query
		return cc.queryDataByPub(stub, args)
	}

	return shim.Error("Received unknown function invocation")
}

// createData - create a new data entry, store into chaincode state
/////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) createData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//        0           1             2        3           4             5
	// "DataEntryID", "Description", "Value", "Unit", "CreationTime", "Publisher",
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	// ==== Input sanitization ====
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

	dataEntryID := args[0]
	description := args[1]
	value := args[2]
	unit := args[3]
	creationTime := args[4]
	publisher := args[5]

	// Check if data entry already exists
	dataAsBytes, err := stub.GetState(dataEntryID)
	if err != nil {
		return shim.Error("Failed to get data entry: " + err.Error())
	} else if dataAsBytes != nil {
		fmt.Println("This data entry already exists: " + dataEntryID)
		return shim.Error("This data entry already exists: " + dataEntryID)
	}

	// Create data entry object and marshal to JSON
	recordType := "DATA_ENTRY"
	dataEntry := &DataEntry{recordType, dataEntryID, description, value, unit, creationTime, publisher}
	dataEntryJSONasBytes, err := json.Marshal(dataEntry)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Save data entry to state
	err = stub.PutState(dataEntryID, dataEntryJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Index the data to enable publisher-based range queries
	// An 'index' is a normal key/value entry in state.
	// The key is a composite key, with the elements that you want to range query on listed first.
	indexName := "Publisher~DataEntryID"
	pubIdIndexKey, err := stub.CreateCompositeKey(indexName, []string{dataEntry.Publisher, dataEntry.DataEntryID})
	if err != nil {
		return shim.Error(err.Error())
	}
	// Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
	// Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value1 := []byte{0x00}
	stub.PutState(pubIdIndexKey, value1)

	// Data entry saved and indexed
	return shim.Success([]byte("Data entry created."))
}

// getDataByID - read data entry from chaincode state based its Id
////////////////////////////////////////////////////////////////////////
func (cc *Chaincode) getDataByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	//    0
	// "DataEntryID"

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting data entry Id to query")
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
	//    0
	// "Publisher"

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting publisher to query")
	}

	publisher := args[0]
	// Query the Publisher~DataEntryID index by publisher
	// This will execute a key range query on all keys starting with 'Publisher'
	pubIDResultsIterator, err := stub.GetStateByPartialCompositeKey("Publisher~DataEntryID", []string{publisher})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer pubIDResultsIterator.Close()

	// Iterate through result set and for each DataEntry found
	var i int
	var dataAsBytes []byte
	for i = 0; pubIDResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the DataEntryID name from the composite key
		responseRange, err := pubIDResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// get the publisher and dataEntryID from Publisher~DataEntryID composite key
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedPublisher := compositeKeyParts[0]
		returnedDataEntryID := compositeKeyParts[1]
		fmt.Printf("- found a data entry from index:%s Publisher:%s DataEntryID:%s\n", objectType, returnedPublisher, returnedDataEntryID)

		response := cc.getDataByID(stub, []string{returnedDataEntryID})
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
