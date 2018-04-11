package main

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func checkInit(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed", string(res.Message))
		t.Fail()
	}
}

func checkInvoke(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.Fail()
	}
}

func checkInvokeFail(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInvoke("1", args)
	if res.Status == shim.OK {
		fmt.Println("Invoke", args, "should fail but did not", string(res.Payload))
		t.Fail()
	}
}

func checkInvokeResponse(t *testing.T, stub *shim.MockStub, args [][]byte, expectedPayload string) {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.Fail()
	}
	if string(res.Payload) != expectedPayload {
		fmt.Println("Expected payload:", expectedPayload)
		fmt.Println("Instead got this:", string(res.Payload))
		t.Fail()
	}
}

func checkInvokeResponseFail(t *testing.T, stub *shim.MockStub, args [][]byte, expectedMessage string) {
	res := stub.MockInvoke("1", args)
	if res.Status == shim.OK {
		fmt.Println("Invoke", args, "should fail")
		fmt.Println("Instead got payload:", string(res.Payload))
		t.Fail()
	}
	if res.Message != expectedMessage {
		fmt.Println("Expected message:", expectedMessage)
		fmt.Println("Instead got this:", res.Message)
		t.Fail()
	}
}

func Test_Init(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init should always success
	checkInit(t, stub, [][]byte{[]byte("1")})
}

func Test_InvokeFail(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("invoke_fail_test", cc)
	args := [][]byte{[]byte("NoFunction"), []byte("test")}
	expectedMessage := "Received unknown function invocation"
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}

func Test_createDataEntryAd(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("20")})
	// create test data entry ad
	args := [][]byte{[]byte("createDataEntryAd"),
		[]byte("1"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("1"), []byte("2")}
	checkInvokeResponse(t, stub, args, "")

	// Check if it is in the state
	args = [][]byte{[]byte("getDataAdByIDAndTime"), []byte("1"), []byte("20181212152030")}
	expectedPayload := "{\"RecordType\":\"DATA_ENTRY_AD\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"???\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":20181212152030,\"Publisher\":\"pub_name\"," +
		"\"Price\":1,\"AccountNo\":\"2\"}"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// create test data entry ad that has the same ID and creationTime Shoult fail
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("1"), []byte("test_data"), []byte("50"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	expectedMessage := "This data entry already exists: 1~20181212152030"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte(""), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should not save to the state and it should fail
	expectedMessage = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte(""), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should not save to the state and it should fail
	expectedMessage = "Argument at position 2 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte(""), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should not save to the state and it should fail
	expectedMessage = "Argument at position 3 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte(""),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should not save to the state and it should fail
	expectedMessage = "Argument at position 4 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte(""), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should not save to the state and it should fail
	expectedMessage = "Argument at position 5 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte(""), []byte("10"), []byte("2")}
	// it should not save to the state and it should fail
	expectedMessage = "Argument at position 6 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte(""), []byte("2")}
	// it should not save to the state and it should fail
	expectedMessage = "Argument at position 7 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("")}
	// it should not save to the state and it should fail
	expectedMessage = "Argument at position 8 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createDataEntryAd that have less than 8 args
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10")}
	// it should not save to the state and it should fail
	expectedMessage = "Incorrect number of arguments. Expecting 8"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createDataEntryAd that have more than 8 args
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2"), []byte("2")}
	// it should not save to the state and it should fail
	expectedMessage = "Incorrect number of arguments. Expecting 8"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createDataEntryAd for negative price
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("-10"), []byte("2")}
	// it should not save to the state and it should fail
	expectedMessage = "Price cannot be negative number."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createDataEntryAd for negative creationTime
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("-20181212152030"), []byte("pub_name"), []byte("-10"), []byte("2")}
	// it should not save to the state and it should fail
	expectedMessage = "Expecting positiv integer or zero as creation time."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createDataEntryAd if creationTime is not uint
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("lol"), []byte("pub_name"), []byte("-10"), []byte("2")}
	// it should not save to the state and it should fail
	expectedMessage = "Expecting positiv integer or zero as creation time."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createDataEntryAd if price is not int
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("lol"), []byte("2")}
	// it should not save to the state and it should fail
	expectedMessage = "Expecting positiv integer or zero as price."
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}

func Test_getDataAdByIDAndTime(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("1")})
	// create test data entry ad
	args := [][]byte{[]byte("createDataEntryAd"),
		[]byte("1"), []byte("test_data"), []byte("10"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("0"), []byte("2")}
	// it should save to the state
	checkInvokeResponse(t, stub, args, "")

	expectedPayload := "{\"RecordType\":\"DATA_ENTRY_AD\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"10\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":20181212152030,\"Publisher\":\"pub_name\"," +
		"\"Price\":0,\"AccountNo\":\"2\"}"

	// It should get the same expected payload
	args = [][]byte{[]byte("getDataAdByIDAndTime"), []byte("1"), []byte("20181212152030")}
	checkInvokeResponse(t, stub, args, expectedPayload)

	/*
		// This cannot be tested because of the MockStub implementation limitations
		// It should fail to get ID that is not in the ledger
		args = [][]byte{[]byte("getDataAdByID"),
			[]byte("2")}
		checkInvokeFail(t, stub, args)
	*/

	// It should fail to getDataAdByIDAndTime if arg is empty string
	args = [][]byte{[]byte("getDataAdByIDAndTime"), []byte(""), []byte("20181212152030")}
	expectedPayload = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedPayload)

	// It should fail to getDataAdByIDAndTime if arg is empty string
	args = [][]byte{[]byte("getDataAdByIDAndTime"), []byte("1"), []byte("")}
	expectedPayload = "Argument at position 2 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedPayload)

	// It should fail to getDataAdByIDAndTime if less than 2 args provided
	args = [][]byte{[]byte("getDataAdByIDAndTime"), []byte("1")}
	expectedPayload = "Incorrect number of arguments. Expecting data entry Id and creationTime"
	checkInvokeResponseFail(t, stub, args, expectedPayload)

	// It should fail to getDataAdByIDAndTime if more than 2 args provided
	args = [][]byte{[]byte("getDataAdByIDAndTime"), []byte("1"), []byte("20181212152030"), []byte("2")}
	expectedPayload = "Incorrect number of arguments. Expecting data entry Id and creationTime"
	checkInvokeResponseFail(t, stub, args, expectedPayload)

	// It should fail to getDataAdByIDAndTime if creationTime is not uint
	args = [][]byte{[]byte("getDataAdByIDAndTime"), []byte("1"), []byte("-20181212152030")}
	expectedPayload = "Expecting positiv integer or zero as creation time."
	checkInvokeResponseFail(t, stub, args, expectedPayload)

	// It should fail to getDataAdByIDAndTime if creationTime is not uint
	args = [][]byte{[]byte("getDataAdByIDAndTime"), []byte("1"), []byte("lol")}
	expectedPayload = "Expecting positiv integer or zero as creation time."
	checkInvokeResponseFail(t, stub, args, expectedPayload)

}

func Test_getDataAdByPub(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("1")})
	// create test data entry ad
	args := [][]byte{[]byte("createDataEntryAd"),
		[]byte("1"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should save to the state
	checkInvoke(t, stub, args)
	expectedPayload := "{\"RecordType\":\"DATA_ENTRY_AD\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"???\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":20181212152030,\"Publisher\":\"pub_name\"," +
		"\"Price\":10,\"AccountNo\":\"2\"}"

	// It should get the same expected payload
	args = [][]byte{[]byte("getDataAdByPub"), []byte("pub_name")}
	checkInvokeResponse(t, stub, args, "["+expectedPayload+"]")

	// create second test data entry ad
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("100"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should save to the state
	checkInvoke(t, stub, args)
	expectedPayload2 := "{\"RecordType\":\"DATA_ENTRY_AD\",\"DataEntryID\":\"2\"" +
		",\"Description\":\"test_data\",\"Value\":\"100\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":20181212152030,\"Publisher\":\"pub_name\"," +
		"\"Price\":10,\"AccountNo\":\"2\"}"

	// It should get both entry as JSON array
	args = [][]byte{[]byte("getDataAdByPub"), []byte("pub_name")}
	expectedPayload3 := "[" + expectedPayload + "," + expectedPayload2 + "]"
	checkInvokeResponse(t, stub, args, expectedPayload3)

	// It should not find entry that is not in state
	args = [][]byte{[]byte("getDataAdByPub"), []byte("pub_name2")}
	// expected only empty array
	expectedPayload = "[]"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should fail with empty arg
	args = [][]byte{[]byte("getDataAdByPub"), []byte("")}
	expectedPayload = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedPayload)

	// It should fail with empty more than one arg
	args = [][]byte{[]byte("getDataAdByPub"), []byte("pub_name"), []byte("pub_name")}
	expectedPayload = "Incorrect number of arguments. Expecting publisher to get"
	checkInvokeResponseFail(t, stub, args, expectedPayload)
}

func Test_getAllDataAdByID(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("1")})
	// create test data entry ad
	args := [][]byte{[]byte("createDataEntryAd"),
		[]byte("1"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}

	expectedPayload := "{\"RecordType\":\"DATA_ENTRY_AD\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"???\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":20181212152030,\"Publisher\":\"pub_name\"," +
		"\"Price\":10,\"AccountNo\":\"2\"}"
	// it should save to the state
	checkInvoke(t, stub, args)
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("1"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152031"), []byte("pub_name"), []byte("10"), []byte("2")}

	expectedPayload2 := "{\"RecordType\":\"DATA_ENTRY_AD\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"???\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":20181212152031,\"Publisher\":\"pub_name\"," +
		"\"Price\":10,\"AccountNo\":\"2\"}"
	// it should save to the state
	checkInvoke(t, stub, args)

	// It should return JSON array with both entries
	args = [][]byte{[]byte("getAllDataAdByID"), []byte("1")}
	expectedPayload3 := "[" + expectedPayload + "," + expectedPayload2 + "]"
	checkInvokeResponse(t, stub, args, expectedPayload3)

	// It should fail with empty arg
	args = [][]byte{[]byte("getAllDataAdByID"), []byte("")}
	expectedPayload = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedPayload)

	// It should fail with empty more than one arg
	args = [][]byte{[]byte("getAllDataAdByID"), []byte("1"), []byte("1")}
	expectedPayload = "Incorrect number of arguments. Expecting data entry Id to get"
	checkInvokeResponseFail(t, stub, args, expectedPayload)
}

func Test_getLatestDataAdByID(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("1")})
	// create test data entry ad
	args := [][]byte{[]byte("createDataEntryAd"),
		[]byte("1"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should save to the state
	checkInvoke(t, stub, args)

	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("1"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152031"), []byte("pub_name"), []byte("10"), []byte("2")}

	// it should save to the state
	checkInvoke(t, stub, args)

	// It should return latest data entry by creationTime
	args = [][]byte{[]byte("getLatestDataAdByID"), []byte("1")}
	expectedPayload := "{\"RecordType\":\"DATA_ENTRY_AD\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"???\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":20181212152031,\"Publisher\":\"pub_name\"," +
		"\"Price\":10,\"AccountNo\":\"2\"}"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should fail with empty arg
	args = [][]byte{[]byte("getLatestDataAdByID"), []byte("")}
	expectedPayload = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedPayload)

	// It should fail with empty more than one arg
	args = [][]byte{[]byte("getLatestDataAdByID"), []byte("1"), []byte("1")}
	expectedPayload = "Incorrect number of arguments. Expecting data entry Id to get"
	checkInvokeResponseFail(t, stub, args, expectedPayload)
}

// For this function we cannot test more because of the MockStub limitations
func Test_revealPaidData(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("1")})

	// It should fail to revealPaidData that have less than 7 args
	args := [][]byte{[]byte("revealPaidData"),
		[]byte("channel1"), []byte("chaincode_data"), []byte("1"), []byte("20181212152030"),
		[]byte("channel3"), []byte("chaincode_tokens")}
	expectedMessage := "Incorrect number of arguments. Expecting 7"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to revealPaidData that have more than 7 args
	args = [][]byte{[]byte("revealPaidData"),
		[]byte("channel1"), []byte("chaincode_data"), []byte("1"), []byte("20181212152030"),
		[]byte("channel3"), []byte("chaincode_tokens"), []byte("TxID-1"), []byte("extra_arg")}
	expectedMessage = "Incorrect number of arguments. Expecting 7"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to revealPaidData that have one empty string arg
	args = [][]byte{[]byte("revealPaidData"),
		[]byte(""), []byte("chaincode_data"), []byte("1"), []byte("20181212152030"),
		[]byte("channel3"), []byte("chaincode_tokens"), []byte("TxID-1")}
	expectedMessage = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to revealPaidData that have one empty string arg
	args = [][]byte{[]byte("revealPaidData"),
		[]byte("channel1"), []byte(""), []byte("1"), []byte("20181212152030"),
		[]byte("channel3"), []byte("chaincode_tokens"), []byte("TxID-1")}
	expectedMessage = "Argument at position 2 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to revealPaidData that have one empty string arg
	args = [][]byte{[]byte("revealPaidData"),
		[]byte("channel1"), []byte("chaincode_data"), []byte(""), []byte("20181212152030"),
		[]byte("channel3"), []byte("chaincode_tokens"), []byte("TxID-1")}
	expectedMessage = "Argument at position 3 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to revealPaidData that have one empty string arg
	args = [][]byte{[]byte("revealPaidData"),
		[]byte("channel1"), []byte("chaincode_data"), []byte("1"), []byte(""),
		[]byte("channel3"), []byte("chaincode_tokens"), []byte("TxID-1")}
	expectedMessage = "Argument at position 4 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to revealPaidData that have one empty string arg
	args = [][]byte{[]byte("revealPaidData"),
		[]byte("channel1"), []byte("chaincode_data"), []byte("1"), []byte("20181212152030"),
		[]byte(""), []byte("chaincode_tokens"), []byte("TxID-1")}
	expectedMessage = "Argument at position 5 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to revealPaidData that have one empty string arg
	args = [][]byte{[]byte("revealPaidData"),
		[]byte("channel1"), []byte("chaincode_data"), []byte("1"), []byte("20181212152030"),
		[]byte("channel3"), []byte(""), []byte("TxID-1")}
	expectedMessage = "Argument at position 6 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to revealPaidData that have one empty string arg
	args = [][]byte{[]byte("revealPaidData"),
		[]byte("channel1"), []byte("chaincode_data"), []byte("1"), []byte("20181212152030"),
		[]byte("channel3"), []byte("chaincode_tokens"), []byte("")}
	expectedMessage = "Argument at position 7 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to revealPaidData that negative creationTime
	args = [][]byte{[]byte("revealPaidData"),
		[]byte("channel1"), []byte("chaincode_data"), []byte("1"), []byte("-20181212152030"),
		[]byte("channel3"), []byte("chaincode_tokens"), []byte("TxID-1")}
	expectedMessage = "Expecting positiv integer or zero as creation time."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to revealPaidData that creationTime is not uint
	args = [][]byte{[]byte("revealPaidData"),
		[]byte("channel1"), []byte("chaincode_data"), []byte("1"), []byte("lol"),
		[]byte("channel3"), []byte("chaincode_tokens"), []byte("TxID-1")}
	expectedMessage = "Expecting positiv integer or zero as creation time."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

}

// For this function we cannot test more because of the MockStub limitations
func Test_checkTXState(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("1")})

	// It should fail to checkTXState that more than 1 arg
	args := [][]byte{[]byte("checkTXState"), []byte("TxID-1"), []byte("extra_arg")}
	expectedMessage := "Incorrect number of arguments. Expecting TxID"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to checkTXState with empty string arg
	args = [][]byte{[]byte("checkTXState"), []byte("")}
	expectedMessage = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}
