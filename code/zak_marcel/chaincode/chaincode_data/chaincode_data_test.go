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

func checkInitFail(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status == shim.OK {
		fmt.Println("Init should fail but it did not", string(res.Message))
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
func Test_init(t *testing.T) {
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

func Test_createData(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("20")})
	// create test data
	args := [][]byte{[]byte("createData"),
		[]byte("1"), []byte("test_data"), []byte("10"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name")}
	checkInvokeResponse(t, stub, args, "")

	// Check if it is in the state
	args = [][]byte{[]byte("getDataByIDAndTime"), []byte("1"), []byte("20181212152030")}
	expectedPayload := "{\"RecordType\":\"DATA_ENTRY\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"10\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":20181212152030,\"Publisher\":\"pub_name\"}"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// create test data that has the same ID and creationTime Shoult fail
	args = [][]byte{[]byte("createData"),
		[]byte("1"), []byte("test_data"), []byte("50"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name")}
	expectedMessage := "This data entry already exists: 1~20181212152030"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createData that have one empty arg
	args = [][]byte{[]byte("createData"),
		[]byte(""), []byte("test_data"), []byte("10"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name")}
	// it should not save to the state and it should fail
	expectedMessage = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createData that have one empty arg
	args = [][]byte{[]byte("createData"),
		[]byte("2"), []byte(""), []byte("10"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name")}
	// it should not save to the state and it should fail
	expectedMessage = "Argument at position 2 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createData that have one empty arg
	args = [][]byte{[]byte("createData"),
		[]byte("2"), []byte("test_data"), []byte(""), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name")}
	// it should not save to the state and it should fail
	expectedMessage = "Argument at position 3 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createData that have one empty arg
	args = [][]byte{[]byte("createData"),
		[]byte("2"), []byte("test_data"), []byte("10"), []byte(""),
		[]byte("20181212152030"), []byte("pub_name")}
	// it should not save to the state and it should fail
	expectedMessage = "Argument at position 4 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createData that have one empty arg
	args = [][]byte{[]byte("createData"),
		[]byte("2"), []byte("test_data"), []byte("10"), []byte("Unit"),
		[]byte(""), []byte("pub_name")}
	// it should not save to the state and it should fail
	expectedMessage = "Argument at position 5 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createData that have one empty arg
	args = [][]byte{[]byte("createData"),
		[]byte("2"), []byte("test_data"), []byte("10"), []byte("Unit"),
		[]byte("20181212152030"), []byte("")}
	// it should not save to the state and it should fail
	expectedMessage = "Argument at position 6 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createData with less than 6 args
	args = [][]byte{[]byte("createData"),
		[]byte("2"), []byte("test_data"), []byte("10"), []byte("Unit"),
		[]byte("20181212152030")}
	// it should not save to the state and it should fail
	expectedMessage = "Incorrect number of arguments. Expecting 6"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createData with more than 6 args
	args = [][]byte{[]byte("createData"),
		[]byte("2"), []byte("test_data"), []byte("10"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("lol")}
	// it should not save to the state and it should fail
	expectedMessage = "Incorrect number of arguments. Expecting 6"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to createData for negative creationTime
	args = [][]byte{[]byte("createData"),
		[]byte("2"), []byte("test_data"), []byte("10"), []byte("Unit"),
		[]byte("-20181212152030"), []byte("pub_name")}
	// it should not save to the state and it should fail
	expectedMessage = "Expecting positiv integer or zero as creation time."
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}

func Test_getDataByIDAndTime(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("1")})
	// create test data
	args := [][]byte{[]byte("createData"),
		[]byte("1"), []byte("test_data"), []byte("10"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name")}
	// it should save to the state
	checkInvokeResponse(t, stub, args, "")

	expectedPayload := "{\"RecordType\":\"DATA_ENTRY\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"10\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":20181212152030,\"Publisher\":\"pub_name\"}"

	// It should get the same expected payload
	args = [][]byte{[]byte("getDataByIDAndTime"), []byte("1"), []byte("20181212152030")}
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should fail to get data that have one empty arg
	args = [][]byte{[]byte("getDataByIDAndTime"), []byte(""), []byte("20181212152030")}
	expectedMessage := "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to get data that have one empty arg
	args = [][]byte{[]byte("getDataByIDAndTime"), []byte("1"), []byte("")}
	expectedMessage = "Argument at position 2 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to get data that have less than 2 args
	args = [][]byte{[]byte("getDataByIDAndTime"), []byte("1")}
	expectedMessage = "Incorrect number of arguments. Expecting data entry Id to get"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to get data that have more than 2 args
	args = [][]byte{[]byte("getDataByIDAndTime"), []byte("1")}
	expectedMessage = "Incorrect number of arguments. Expecting data entry Id to get"
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}

func Test_getAllDataByID(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("1")})
	// create test data
	args := [][]byte{[]byte("createData"),
		[]byte("1"), []byte("test_data"), []byte("10"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name")}
	// it should save to the state
	checkInvokeResponse(t, stub, args, "")
	expectedPayload := "{\"RecordType\":\"DATA_ENTRY\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"10\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":20181212152030,\"Publisher\":\"pub_name\"}"

	// It should get the same expected payload
	args = [][]byte{[]byte("getAllDataByID"), []byte("1")}
	checkInvokeResponse(t, stub, args, "["+expectedPayload+"]")

	// create second test data
	args = [][]byte{[]byte("createData"),
		[]byte("1"), []byte("test_data"), []byte("100"), []byte("Unit"),
		[]byte("20181212152031"), []byte("pub_name")}
	// it should save to the state
	checkInvoke(t, stub, args)
	expectedPayload2 := "{\"RecordType\":\"DATA_ENTRY\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"100\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":20181212152031,\"Publisher\":\"pub_name\"}"

	// It should get both entry as JSON array
	args = [][]byte{[]byte("getAllDataByID"), []byte("1")}
	expectedPayload3 := "[" + expectedPayload + "," + expectedPayload2 + "]"
	checkInvokeResponse(t, stub, args, expectedPayload3)

	// It should not find entry that is not in state
	args = [][]byte{[]byte("getAllDataByID"), []byte("2")}
	// expected only empty array
	expectedPayload = "[]"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should fail with empty arg
	args = [][]byte{[]byte("getAllDataByID"), []byte("")}
	expectedPayload = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedPayload)

	// It should fail with empty more than one arg
	args = [][]byte{[]byte("getAllDataByID"), []byte("1"), []byte("1")}
	expectedPayload = "Incorrect number of arguments. Expecting data entry Id to get"
	checkInvokeResponseFail(t, stub, args, expectedPayload)
}

func Test_getLatestDataByID(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("1")})
	// create test data
	args := [][]byte{[]byte("createData"),
		[]byte("1"), []byte("test_data"), []byte("10"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name")}
	// it should save to the state
	checkInvokeResponse(t, stub, args, "")

	// create second test data that are produced later in time
	args = [][]byte{[]byte("createData"),
		[]byte("1"), []byte("test_data"), []byte("100"), []byte("Unit"),
		[]byte("20181212152031"), []byte("pub_name")}
	// it should save to the state
	checkInvoke(t, stub, args)
	expectedPayload := "{\"RecordType\":\"DATA_ENTRY\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"100\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":20181212152031,\"Publisher\":\"pub_name\"}"

	// It should get both entry as JSON array
	args = [][]byte{[]byte("getLatestDataByID"), []byte("1")}
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should fail with empty arg
	args = [][]byte{[]byte("getLatestDataByID"), []byte("")}
	expectedPayload = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedPayload)

	// It should fail with empty more than one arg
	args = [][]byte{[]byte("getLatestDataByID"), []byte("1"), []byte("1")}
	expectedPayload = "Incorrect number of arguments. Expecting data entry Id to get"
	checkInvokeResponseFail(t, stub, args, expectedPayload)
}

func Test_getDataByPub(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("1")})
	// create test data
	args := [][]byte{[]byte("createData"),
		[]byte("1"), []byte("test_data"), []byte("10"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name")}
	// it should save to the state
	checkInvokeResponse(t, stub, args, "")
	expectedPayload := "{\"RecordType\":\"DATA_ENTRY\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"10\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":20181212152030,\"Publisher\":\"pub_name\"}"

	// It should get the same expected payload
	args = [][]byte{[]byte("getDataByPub"), []byte("pub_name")}
	checkInvokeResponse(t, stub, args, "["+expectedPayload+"]")

	// create second test data
	args = [][]byte{[]byte("createData"),
		[]byte("2"), []byte("test_data"), []byte("100"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name")}
	// it should save to the state
	checkInvoke(t, stub, args)
	expectedPayload2 := "{\"RecordType\":\"DATA_ENTRY\",\"DataEntryID\":\"2\"" +
		",\"Description\":\"test_data\",\"Value\":\"100\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":20181212152030,\"Publisher\":\"pub_name\"}"

	// It should get both entry as JSON array
	args = [][]byte{[]byte("getDataByPub"), []byte("pub_name")}
	expectedPayload3 := "[" + expectedPayload + "," + expectedPayload2 + "]"
	checkInvokeResponse(t, stub, args, expectedPayload3)

	// It should not find entry that is not in state
	args = [][]byte{[]byte("getDataByPub"), []byte("pub_name2")}
	// expected only empty array
	expectedPayload = "[]"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should fail with empty arg
	args = [][]byte{[]byte("getDataByPub"), []byte("")}
	expectedPayload = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedPayload)

	// It should fail with empty more than one arg
	args = [][]byte{[]byte("getDataByPub"), []byte("pub_name"), []byte("pub_name")}
	expectedPayload = "Incorrect number of arguments. Expecting publisher to get"
	checkInvokeResponseFail(t, stub, args, expectedPayload)
}
