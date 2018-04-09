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

func checkState(t *testing.T, stub *shim.MockStub, name string, value string) {
	bytes := stub.State[name]
	if bytes == nil {
		fmt.Println("State", name, "failed to get value")
		t.Fail()
	}
	if string(bytes) != value {
		fmt.Println("State value", name, "was", string(bytes), "instead required", value)
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
	stub := shim.NewMockStub("nit_test", cc)

	// Init should always success
	checkInit(t, stub, [][]byte{[]byte("1"), []byte("20")})
}

func Test_createDataEntryAd(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("1"), []byte("20")})
	// create test data entry ad
	args := [][]byte{[]byte("createDataEntryAd"),
		[]byte("1"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should save to the state
	checkInvoke(t, stub, args)
	expectedPayload := "{\"RecordType\":\"DATA_ENTRY_AD\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"???\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":\"20181212152030\",\"Publisher\":\"pub_name\"," +
		"\"Price\":10,\"AccountNo\":\"2\"}"
	checkState(t, stub, "1", expectedPayload)

	// create test data entry ad that has the same ID, Shoult fail
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("1"), []byte("test_data"), []byte("50"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// It should get updated value for the entry ID
	checkInvokeFail(t, stub, args)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte(""), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should not save to the state and it should fail
	checkInvokeFail(t, stub, args)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte(""), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should not save to the state and it should fail
	checkInvokeFail(t, stub, args)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte(""), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should not save to the state and it should fail
	checkInvokeFail(t, stub, args)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte(""),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should not save to the state and it should fail
	checkInvokeFail(t, stub, args)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte(""), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should not save to the state and it should fail
	checkInvokeFail(t, stub, args)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte(""), []byte("10"), []byte("2")}
	// it should not save to the state and it should fail
	checkInvokeFail(t, stub, args)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte(""), []byte("2")}
	// it should not save to the state and it should fail
	checkInvokeFail(t, stub, args)

	// It should fail to createDataEntryAd that have one empty arg
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("")}
	// it should not save to the state and it should fail
	checkInvokeFail(t, stub, args)

	// It should fail to createDataEntryAd for negative price
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("-10"), []byte("2")}
	// it should not save to the state and it should fail
	checkInvokeFail(t, stub, args)
}

func Test_getDataAdByID(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("1"), []byte("20")})
	// create test data entry ad
	args := [][]byte{[]byte("createDataEntryAd"),
		[]byte("1"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should save to the state
	checkInvoke(t, stub, args)
	expectedPayload := "{\"RecordType\":\"DATA_ENTRY_AD\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"???\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":\"20181212152030\",\"Publisher\":\"pub_name\"," +
		"\"Price\":10,\"AccountNo\":\"2\"}"
	checkState(t, stub, "1", expectedPayload)

	// It should get the same expected payload
	args = [][]byte{[]byte("getDataAdByID"), []byte("1")}
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.Fail()
	}
	if string(res.Payload) != expectedPayload {
		fmt.Println("getDataAdByID \"1\" payload was not as expected")
		fmt.Println("Expected:", expectedPayload, "Got:", string(res.Payload))
		t.Fail()
	}
	/*
		// This cannot be tested because of the MockStub implementation limitations
		// It should fail to get ID that is not in the ledger
		args = [][]byte{[]byte("getDataAdByID"),
			[]byte("2")}
		checkInvoke(t, stub, args)
	*/
}

func Test_queryDataByPub(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("init_test", cc)

	// Init
	checkInit(t, stub, [][]byte{[]byte("1"), []byte("20")})
	// create test data entry ad
	args := [][]byte{[]byte("createDataEntryAd"),
		[]byte("1"), []byte("test_data"), []byte("???"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should save to the state
	checkInvoke(t, stub, args)
	expectedPayload1 := "{\"RecordType\":\"DATA_ENTRY_AD\",\"DataEntryID\":\"1\"" +
		",\"Description\":\"test_data\",\"Value\":\"???\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":\"20181212152030\",\"Publisher\":\"pub_name\"," +
		"\"Price\":10,\"AccountNo\":\"2\"}"
	checkState(t, stub, "1", expectedPayload1)

	// It should get the same expected payload
	args = [][]byte{[]byte("queryDataByPub"), []byte("pub_name")}
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.Fail()
	}
	if string(res.Payload) != "["+expectedPayload1+"]" {
		fmt.Println("queryDataByPub \"pub_name\" payload was not as expected")
		fmt.Println("Expected:", expectedPayload1, "Got:", string(res.Payload))
		t.Fail()
	}

	// create second test data entry ad
	args = [][]byte{[]byte("createDataEntryAd"),
		[]byte("2"), []byte("test_data"), []byte("100"), []byte("Unit"),
		[]byte("20181212152030"), []byte("pub_name"), []byte("10"), []byte("2")}
	// it should save to the state
	checkInvoke(t, stub, args)
	expectedPayload2 := "{\"RecordType\":\"DATA_ENTRY_AD\",\"DataEntryID\":\"2\"" +
		",\"Description\":\"test_data\",\"Value\":\"100\",\"Unit\":\"Unit\"," +
		"\"CreationTime\":\"20181212152030\",\"Publisher\":\"pub_name\"," +
		"\"Price\":10,\"AccountNo\":\"2\"}"
	checkState(t, stub, "2", expectedPayload2)

	// It should get both entry as JSON array
	args = [][]byte{[]byte("queryDataByPub"), []byte("pub_name")}
	res = stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.Fail()
	}
	expectedPayload3 := "[" + expectedPayload1 + "," + expectedPayload2 + "]"
	if string(res.Payload) != expectedPayload3 {
		fmt.Println("queryDataByPub \"pub_name\" payload was not as expected")
		fmt.Println("Expected:", expectedPayload3, "Got:", string(res.Payload))
		t.Fail()
	}

	// It should not find entry that is not in state
	args = [][]byte{[]byte("queryDataByPub"), []byte("pub_name2")}
	res = stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.Fail()
	}
	// expected only empty array
	expectedPayload4 := "[]"
	if string(res.Payload) != expectedPayload4 {
		fmt.Println("queryDataByPub \"pub_name2\" payload was not as expected")
		fmt.Println("Expected:", expectedPayload4, "Got:", string(res.Payload))
		t.Fail()
	}
}
