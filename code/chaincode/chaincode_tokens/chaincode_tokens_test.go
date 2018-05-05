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
	stub := shim.NewMockStub("tokens_init_test", cc)

	// It should Init 1 account with 10 000 tokens
	checkInit(t, stub, [][]byte{[]byte("10000")})
	checkState(t, stub, "1",
		"{\"RecordType\":\"ACCOUNT\",\"AccountID\":\"1\",\"Name\":\"Init_Account\",\"OwnerID\":\"\",\"Tokens\":10000}")

	// It should Init 1 accounts with 0 tokens
	stub = shim.NewMockStub("tokens_init_test", cc)
	checkInit(t, stub, [][]byte{[]byte("0")})
	checkState(t, stub, "1",
		"{\"RecordType\":\"ACCOUNT\",\"AccountID\":\"1\",\"Name\":\"Init_Account\",\"OwnerID\":\"\",\"Tokens\":0}")

	// It should not Init an account with negative number of tokens
	stub = shim.NewMockStub("tokens_init_test", cc)
	checkInitFail(t, stub, [][]byte{[]byte("-10")})

	// It should not Init with less args than 1
	stub = shim.NewMockStub("tokens_init_test", cc)
	checkInitFail(t, stub, [][]byte{})

	// It should not Init with more args than 1
	stub = shim.NewMockStub("tokens_init_test", cc)
	checkInitFail(t, stub, [][]byte{[]byte("1"), []byte("1")})

	// It should not Init with empty arg
	stub = shim.NewMockStub("tokens_init_test", cc)
	checkInitFail(t, stub, [][]byte{[]byte("")})
}

func Test_InvokeFail(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("invoke_fail_test", cc)
	args := [][]byte{[]byte("NoFunction"), []byte("test")}
	expectedMessage := "Received unknown function invocation"
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}

func Test_createAccount(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("create_acc_test", cc)

	// Init 1 account with 10 000 tokens
	checkInit(t, stub, [][]byte{[]byte("10000")})

	// It should create account
	args := [][]byte{[]byte("createAccount"), []byte("2"), []byte("acc_name")}
	expectedPayload := "Account created"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should fail to create an account with ID that already exists
	args = [][]byte{[]byte("createAccount"), []byte("2"), []byte("acc_name")}
	expectedMessage := "This account already exists: 2"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with empty string arg
	args = [][]byte{[]byte("createAccount"), []byte(""), []byte("acc_name")}
	expectedMessage = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with empty string arg
	args = [][]byte{[]byte("createAccount"), []byte("3"), []byte("")}
	expectedMessage = "Argument at position 2 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with less than 2 args
	args = [][]byte{[]byte("createAccount"), []byte("3")}
	expectedMessage = "Incorrect number of arguments. Expecting account Id and name"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with more than 2 args
	args = [][]byte{[]byte("createAccount"), []byte("3"), []byte("acc_name"), []byte("acc_name")}
	expectedMessage = "Incorrect number of arguments. Expecting account Id and name"
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}

func Test_deleteAccountByID(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("tokens_init_test", cc)

	// Init 1 account with 10 000 tokens
	checkInit(t, stub, [][]byte{[]byte("10000")})
	// create account with 0 tokens
	args := [][]byte{[]byte("createAccount"), []byte("2"), []byte("acc_name")}
	expectedPayload := "Account created"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// it should delete account that have 0 tokens
	args = [][]byte{[]byte("deleteAccountByID"), []byte("2")}
	expectedPayload = "Account deleted"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// it should not be possible to delete an account that have some tokens
	args = [][]byte{[]byte("deleteAccountByID"), []byte("1")}
	expectedMessage := "Account cannot be deleted. Amount of tokens is not 0."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with empty string arg
	args = [][]byte{[]byte("deleteAccountByID"), []byte("")}
	expectedMessage = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with more than 2 args
	args = [][]byte{[]byte("deleteAccountByID"), []byte("2"), []byte("2")}
	expectedMessage = "Incorrect number of arguments. Expecting AccountID."
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}

func Test_getAccountByID(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("get_account_test", cc)

	// Init 1 account with 10 tokens
	checkInit(t, stub, [][]byte{[]byte("10")})

	// It should get account with ID "1" that was Init
	args := [][]byte{[]byte("getAccountByID"), []byte("1")}
	expectedPayload := "{\"RecordType\":\"ACCOUNT\",\"AccountID\":\"1\",\"Name\":\"Init_Account\",\"OwnerID\":\"\",\"Tokens\":10}"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should fail with empty string arg
	args = [][]byte{[]byte("getAccountByID"), []byte("")}
	checkInvokeFail(t, stub, args)

	// It should fail with more than one arg
	args = [][]byte{[]byte("getAccountByID"), []byte("1"), []byte("a")}
	checkInvokeFail(t, stub, args)

	/*
		// This cannot be tested because of the limitations of MockStub implementation
		// It should fail to get account that is not created
		args = [][]byte{[]byte("getAccountByID"), []byte("2")}
		expectedMessage := ""
		checkInvokeResponseFail(t, stub, args, expectedMessage)
	*/
}

func Test_getAccountHistoryByID(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("get_history_test", cc)

	// Init 1 account with 10 tokens
	checkInit(t, stub, [][]byte{[]byte("10")})
	/*
		// This cannot be tested because of the limitations of MockStub implementation
		// It should return history for account ID "1"
		args := [][]byte{[]byte("getAccountHistoryByID"), []byte("1")}
		expectedPayload := ""
		checkInvokeResponse(t, stub, args, expectedPayload)
	*/

	// It should fail with empty string arg
	args := [][]byte{[]byte("getAccountHistoryByID"), []byte("")}
	expectedMessage := "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with more than one args
	args = [][]byte{[]byte("getAccountHistoryByID"), []byte("1"), []byte("lol")}
	expectedMessage = "Incorrect number of arguments. Expecting AccountID"
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}

func Test_getAccountByName(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("get_acc_test", cc)

	// Init 1 account with 10 tokens
	checkInit(t, stub, [][]byte{[]byte("10")})

	// It should return one account
	args := [][]byte{[]byte("getAccountByName"), []byte("Init_Account")}
	expectedPayload := "[{\"RecordType\":\"ACCOUNT\",\"AccountID\":\"1\",\"Name\":\"Init_Account\",\"OwnerID\":\"\",\"Tokens\":10}]"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// create second account
	args = [][]byte{[]byte("createAccount"), []byte("2"), []byte("Init_Account")}
	expectedPayload = "Account created"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should return JSON array of two accounts
	args = [][]byte{[]byte("getAccountByName"), []byte("Init_Account")}
	expectedPayload = "[{\"RecordType\":\"ACCOUNT\",\"AccountID\":\"1\",\"Name\":\"Init_Account\",\"OwnerID\":\"\",\"Tokens\":10}" +
		"," +
		"{\"RecordType\":\"ACCOUNT\",\"AccountID\":\"2\",\"Name\":\"Init_Account\",\"OwnerID\":\"\",\"Tokens\":0}]"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should fail with empty string arg
	args = [][]byte{[]byte("getAccountByName"), []byte("")}
	expectedMessage := "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with more than one args
	args = [][]byte{[]byte("getAccountByName"), []byte("1"), []byte("lol")}
	expectedMessage = "Incorrect number of arguments. Expecting name of account holder"
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}

func Test_sendTokensFast(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("tokens_init_test", cc)

	// Init 1 account with 10 000 tokens
	checkInit(t, stub, [][]byte{[]byte("10000")})

	// create another acc without tokens
	args := [][]byte{[]byte("createAccount"), []byte("2"), []byte("acc_name")}
	expectedPayload := "Account created"
	checkInvokeResponse(t, stub, args, expectedPayload)
	// It should transfer tokens that are not for data purchase immediatelly
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("2"), []byte("1"), []byte("false")}
	expectedPayload = "1"
	checkInvokeResponse(t, stub, args, expectedPayload)
	// Check the result
	args = [][]byte{[]byte("getAccountTokens"), []byte("1")}
	expectedPayload = "9999"
	checkInvokeResponse(t, stub, args, expectedPayload)
	args = [][]byte{[]byte("getAccountTokens"), []byte("2")}
	expectedPayload = "1"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should transfer tokens that are for data purchase and create pendingTx
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("2"), []byte("1"), []byte("true")}
	expectedPayload = "2"
	res := stub.MockInvoke("2", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.Fail()
	}
	if string(res.Payload) != expectedPayload {
		fmt.Println("Expected payload:", expectedPayload)
		fmt.Println("Instead got this:", string(res.Payload))
		t.Fail()
	}
	// Check the result
	args = [][]byte{[]byte("getAccountTokens"), []byte("1")}
	expectedPayload = "9998"
	checkInvokeResponse(t, stub, args, expectedPayload)
	// It is as pending therefore not available for account 2
	args = [][]byte{[]byte("getAccountTokens"), []byte("2")}
	expectedPayload = "1"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should not transfer tokens if tokens limit for fast transfer is exceeded
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("2"), []byte("11"), []byte("false")}
	expectedMessage := "Exceeded max number of tokens for fast transaction. Use safe token transfer instead."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should not transfer negative amount of tokens
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("2"), []byte("-1"), []byte("false")}
	expectedMessage = "Expecting positive integer as number of tokens to transfer."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should not transfer tokens if sender and recipient acc is the same
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("1"), []byte("1"), []byte("false")}
	expectedMessage = "From account and to account cannot be the same."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should not transfer tokens if amount is not a number
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("2"), []byte("lol"), []byte("false")}
	expectedMessage = "Expecting positive integer as number of tokens to transfer."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should not transfer tokens if dataPurchase param is not bool value
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("2"), []byte("1"), []byte("lol")}
	expectedMessage = "Expecting boolean value. If this transfer is for data purchase or not."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with empty string arg
	args = [][]byte{[]byte("sendTokensFast"), []byte(""), []byte("2"), []byte("1"), []byte("false")}
	expectedMessage = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with empty string arg
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte(""), []byte("1"), []byte("false")}
	expectedMessage = "Argument at position 2 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with empty string arg
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("2"), []byte(""), []byte("false")}
	expectedMessage = "Argument at position 3 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with empty string arg
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("2"), []byte("1"), []byte("")}
	expectedMessage = "Argument at position 4 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with more than 4 args
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("2"), []byte("1"), []byte("false"), []byte("lol")}
	expectedMessage = "Incorrect number of arguments. Expecting FromAccountId, ToAccountId, Amount, dataPurchase"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with less than 4 args
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("2"), []byte("1")}
	expectedMessage = "Incorrect number of arguments. Expecting FromAccountId, ToAccountId, Amount, dataPurchase"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	/*
		// These tests cannot be executed in Mock environment yet.
		// They create panic even though chaincode is ok and should return error.
		// It is because of the implementation of Mock state as map and it dereference 0 pointer
		// It should not transfer tokens to account that does not exist
		args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("3"), []byte("100")}
		checkInvokeFail(t, stub, args)

		// It should not transfer tokens from account that does not exist
		args = [][]byte{[]byte("sendTokensFast"), []byte("3"), []byte("1"), []byte("100")}
		checkInvokeFail(t, stub, args)

	*/
}

func Test_sendTokensSafe(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("tokens_init_test", cc)

	// Init 1 account with 10 000 tokens
	checkInit(t, stub, [][]byte{[]byte("10000")})

	// create another acc without tokens
	args := [][]byte{[]byte("createAccount"), []byte("2"), []byte("acc_name")}
	expectedPayload := "Account created"
	checkInvokeResponse(t, stub, args, expectedPayload)
	// It should transfer tokens
	args = [][]byte{[]byte("sendTokensSafe"), []byte("1"), []byte("2"), []byte("100"), []byte("false")}
	expectedPayload = "1"
	checkInvokeResponse(t, stub, args, expectedPayload)
	// Check the result
	args = [][]byte{[]byte("getAccountTokens"), []byte("1")}
	expectedPayload = "9900"
	checkInvokeResponse(t, stub, args, expectedPayload)
	args = [][]byte{[]byte("getAccountTokens"), []byte("2")}
	expectedPayload = "100"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should transfer tokens that are for data purchase and create pendingTx
	args = [][]byte{[]byte("sendTokensSafe"), []byte("1"), []byte("2"), []byte("100"), []byte("true")}
	expectedPayload = "2"
	res := stub.MockInvoke("2", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.Fail()
	}
	if string(res.Payload) != expectedPayload {
		fmt.Println("Expected payload:", expectedPayload)
		fmt.Println("Instead got this:", string(res.Payload))
		t.Fail()
	}
	// Check the result
	args = [][]byte{[]byte("getAccountTokens"), []byte("1")}
	expectedPayload = "9800"
	checkInvokeResponse(t, stub, args, expectedPayload)
	// It is as pending therefore not available for account 2
	args = [][]byte{[]byte("getAccountTokens"), []byte("2")}
	expectedPayload = "100"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should not transfer tokens that are not available
	args = [][]byte{[]byte("sendTokensSafe"), []byte("2"), []byte("1"), []byte("101"), []byte("false")}
	expectedPayload = "Not enough tokens on the sender's account"
	checkInvokeResponseFail(t, stub, args, expectedPayload)

	// It should not transfer tokens if sender and recipient acc is the same
	args = [][]byte{[]byte("sendTokensSafe"), []byte("1"), []byte("1"), []byte("10"), []byte("false")}
	expectedMessage := "From account and to account cannot be the same."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should not transfer negative amount of tokens
	args = [][]byte{[]byte("sendTokensSafe"), []byte("1"), []byte("2"), []byte("-10"), []byte("false")}
	expectedMessage = "Expecting positive integer as number of tokens to transfer."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should not transfer tokens if amount is not a number
	args = [][]byte{[]byte("sendTokensSafe"), []byte("1"), []byte("2"), []byte("lol"), []byte("false")}
	expectedMessage = "Expecting positive integer as number of tokens to transfer."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should not transfer tokens if dataPurchase param is not bool value
	args = [][]byte{[]byte("sendTokensSafe"), []byte("1"), []byte("2"), []byte("1"), []byte("lol")}
	expectedMessage = "Expecting boolean value. If this transfer is for data purchase or not."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with empty string arg
	args = [][]byte{[]byte("sendTokensSafe"), []byte(""), []byte("2"), []byte("10"), []byte("false")}
	expectedMessage = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with empty string arg
	args = [][]byte{[]byte("sendTokensSafe"), []byte("1"), []byte(""), []byte("10"), []byte("false")}
	expectedMessage = "Argument at position 2 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with empty string arg
	args = [][]byte{[]byte("sendTokensSafe"), []byte("1"), []byte("2"), []byte(""), []byte("false")}
	expectedMessage = "Argument at position 3 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with empty string arg
	args = [][]byte{[]byte("sendTokensSafe"), []byte("1"), []byte("2"), []byte("1"), []byte("")}
	expectedMessage = "Argument at position 4 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with more than 4 args
	args = [][]byte{[]byte("sendTokensSafe"), []byte("1"), []byte("2"), []byte("100"), []byte("false"), []byte("lol")}
	expectedMessage = "Incorrect number of arguments. Expecting FromAccountId, ToAccountId, Amount, dataPurchase"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with less than 4 args
	args = [][]byte{[]byte("sendTokensSafe"), []byte("1"), []byte("2"), []byte("1")}
	expectedMessage = "Incorrect number of arguments. Expecting FromAccountId, ToAccountId, Amount, dataPurchase"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	/*
		// These tests cannot be executed in Mock environment yet.
		// They create panic even though chaincode is ok and should return error.
		// It is because of the implementation of Mock state as map and it dereference 0 pointer
		// It should not transfer tokens to account that does not exist
		args = [][]byte{[]byte("sendTokensSafe"), []byte("1"), []byte("3"), []byte("100")}
		checkInvokeFail(t, stub, args)

		// It should not transfer tokens from account that does not exist
		args = [][]byte{[]byte("sendTokensSafe"), []byte("3"), []byte("1"), []byte("100")}
		checkInvokeFail(t, stub, args)

	*/
}

/*
// This test cannot be used because of the limitations of MockChaincode
func Test_getHistoryForAccount(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("tokens_init_test", cc)

	// Init 1 account with 10 000 tokens
	checkInit(t, stub, [][]byte{[]byte("10000")})

	// create another acc without tokens
	args := [][]byte{[]byte("createAccount"), []byte("2"), []byte("acc_name")}
	expectedPayload := "Account created"
	checkInvokeResponse(t, stub, args, expectedPayload)
	// It should transfer tokens
	args := [][]byte{[]byte("transferTokens"), []byte("1"), []byte("2"), []byte("100")}
	checkInvoke(t, stub, args)

	args := [][]byte{[]byte("getAccHistory"), []byte("1")}
	checkInvoke(t, stub, args)
}
*/

func Test_getTxDetails(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("tokens_init_test", cc)

	// Init 1 account with 10 000 tokens
	checkInit(t, stub, [][]byte{[]byte("10000")})

	// create another acc without tokens
	args := [][]byte{[]byte("createAccount"), []byte("2"), []byte("acc_name")}
	expectedPayload := "Account created"
	checkInvokeResponse(t, stub, args, expectedPayload)
	// It should transfer tokens
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("2"), []byte("1"), []byte("false")}
	res := stub.MockInvoke("2", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.Fail()
	}
	// get the TxID and participants
	args = [][]byte{[]byte("getTxDetails"), res.Payload}
	expectedPayload = "1->2->1->ValidTx"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should fail with random TxID
	args = [][]byte{[]byte("getTxDetails"), []byte("-4863asfaebh")}
	expectedMessage := "Transaction was not found."
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with empty string arg
	args = [][]byte{[]byte("getTxDetails"), []byte("")}
	expectedMessage = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with more than one args
	args = [][]byte{[]byte("getTxDetails"), []byte("1"), []byte("lol")}
	expectedMessage = "Incorrect number of arguments. Expecting TxID"
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}

// For this function we cannot test more because of the MockStub limitations
func Test_changePendingTx(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("tokens_init_test", cc)

	// Init 1 account with 10 000 tokens
	checkInit(t, stub, [][]byte{[]byte("10000")})

	// It should fail to changePendingTx with less than 3 args
	args := [][]byte{[]byte("changePendingTx"), []byte("channel1"), []byte("chaincode_ad")}
	expectedMessage := "Incorrect number of arguments. Expecting 3"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to changePendingTx with more than 3 args
	args = [][]byte{[]byte("changePendingTx"), []byte("channel1"), []byte("chaincode_ad"),
		[]byte("TxID-1"), []byte("extra_arg")}
	expectedMessage = "Incorrect number of arguments. Expecting 3"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to changePendingTx with empty string arg
	args = [][]byte{[]byte("changePendingTx"), []byte(""), []byte("chaincode_ad"),
		[]byte("TxID-1")}
	expectedMessage = "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to changePendingTx with empty string arg
	args = [][]byte{[]byte("changePendingTx"), []byte("channel1"), []byte(""),
		[]byte("TxID-1")}
	expectedMessage = "Argument at position 2 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail to changePendingTx with empty string arg
	args = [][]byte{[]byte("changePendingTx"), []byte("channel1"), []byte("chaincode_ad"),
		[]byte("")}
	expectedMessage = "Argument at position 3 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}

func Test_pruneAccountTx(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("tokens_init_test", cc)

	// Init 1 account with 10 000 tokens
	checkInit(t, stub, [][]byte{[]byte("10000")})

	// create another acc without tokens
	args := [][]byte{[]byte("createAccount"), []byte("2"), []byte("acc_name")}
	expectedPayload := "Account created"
	checkInvokeResponse(t, stub, args, expectedPayload)
	// It should transfer tokens
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("2"), []byte("1"), []byte("false")}
	res := stub.MockInvoke("2", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.Fail()
	}
	args = [][]byte{[]byte("sendTokensFast"), []byte("1"), []byte("2"), []byte("1"), []byte("false")}
	res = stub.MockInvoke("3", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.Fail()
	}
	// prune Tx for acc ID
	args = [][]byte{[]byte("pruneAccountTx"), []byte("2")}
	res = stub.MockInvoke("4", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.Fail()
	}

	// get the TxID and participants
	args = [][]byte{[]byte("getTxDetails"), []byte("4")}
	expectedPayload = "pruneTx->2->2->ValidTx"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should fail with empty string arg
	args = [][]byte{[]byte("pruneAccountTx"), []byte("")}
	expectedMessage := "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with more than one args
	args = [][]byte{[]byte("pruneAccountTx"), []byte("1"), []byte("lol")}
	expectedMessage = "Incorrect number of arguments. Expecting account ID"
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}

func Test_updateAccountTokens(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("tokens_init_test", cc)

	// Init 1 account with 10 000 tokens
	checkInit(t, stub, [][]byte{[]byte("10000")})

	// create another acc without tokens
	args := [][]byte{[]byte("createAccount"), []byte("2"), []byte("acc_name")}
	expectedPayload := "Account created"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// send tokens to account 2
	args = [][]byte{[]byte("sendTokensSafe"), []byte("1"), []byte("2"), []byte("100"), []byte("false")}
	expectedPayload = "1"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// accounts should have the initial value because they were not updated yet
	args = [][]byte{[]byte("getAccountByID"), []byte("1")}
	expectedPayload = "{\"RecordType\":\"ACCOUNT\",\"AccountID\":\"1\",\"Name\":\"Init_Account\",\"OwnerID\":\"\",\"Tokens\":10000}"
	checkInvokeResponse(t, stub, args, expectedPayload)
	args = [][]byte{[]byte("getAccountByID"), []byte("2")}
	expectedPayload = "{\"RecordType\":\"ACCOUNT\",\"AccountID\":\"2\",\"Name\":\"acc_name\",\"OwnerID\":\"\",\"Tokens\":0}"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// Check if Tx was successful
	args = [][]byte{[]byte("getAccountTokens"), []byte("1")}
	expectedPayload = "9900"
	checkInvokeResponse(t, stub, args, expectedPayload)
	// Check if Tx was successful
	args = [][]byte{[]byte("getAccountTokens"), []byte("2")}
	expectedPayload = "100"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// Update the amount of tokens on account 1
	args = [][]byte{[]byte("updateAccountTokens"), []byte("1")}
	expectedPayload = "{\"RecordType\":\"ACCOUNT\",\"AccountID\":\"1\",\"Name\":\"Init_Account\",\"OwnerID\":\"\",\"Tokens\":9900}"
	checkInvokeResponse(t, stub, args, expectedPayload)
	// Update the amount of tokens on account 2
	args = [][]byte{[]byte("updateAccountTokens"), []byte("2")}
	expectedPayload = "{\"RecordType\":\"ACCOUNT\",\"AccountID\":\"2\",\"Name\":\"acc_name\",\"OwnerID\":\"\",\"Tokens\":100}"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// accounts should have the updated value
	args = [][]byte{[]byte("getAccountByID"), []byte("1")}
	expectedPayload = "{\"RecordType\":\"ACCOUNT\",\"AccountID\":\"1\",\"Name\":\"Init_Account\",\"OwnerID\":\"\",\"Tokens\":9900}"
	checkInvokeResponse(t, stub, args, expectedPayload)
	args = [][]byte{[]byte("getAccountByID"), []byte("2")}
	expectedPayload = "{\"RecordType\":\"ACCOUNT\",\"AccountID\":\"2\",\"Name\":\"acc_name\",\"OwnerID\":\"\",\"Tokens\":100}"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should fail with empty string arg
	args = [][]byte{[]byte("updateAccountTokens"), []byte("")}
	expectedMessage := "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with more than one args
	args = [][]byte{[]byte("updateAccountTokens"), []byte("1"), []byte("lol")}
	expectedMessage = "Incorrect number of arguments. Expecting account ID"
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}

func Test_getAccountTokens(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("tokens_init_test", cc)

	// Init 1 account with 10 000 tokens
	checkInit(t, stub, [][]byte{[]byte("10000")})

	// create another acc without tokens
	args := [][]byte{[]byte("createAccount"), []byte("2"), []byte("acc_name")}
	expectedPayload := "Account created"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// send tokens to account 2
	args = [][]byte{[]byte("sendTokensSafe"), []byte("1"), []byte("2"), []byte("100"), []byte("false")}
	expectedPayload = "1"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// accounts should have the initial value because they were not updated yet
	args = [][]byte{[]byte("getAccountByID"), []byte("1")}
	expectedPayload = "{\"RecordType\":\"ACCOUNT\",\"AccountID\":\"1\",\"Name\":\"Init_Account\",\"OwnerID\":\"\",\"Tokens\":10000}"
	checkInvokeResponse(t, stub, args, expectedPayload)
	args = [][]byte{[]byte("getAccountByID"), []byte("2")}
	expectedPayload = "{\"RecordType\":\"ACCOUNT\",\"AccountID\":\"2\",\"Name\":\"acc_name\",\"OwnerID\":\"\",\"Tokens\":0}"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// Check if Tx was successful
	args = [][]byte{[]byte("getAccountTokens"), []byte("1")}
	expectedPayload = "9900"
	checkInvokeResponse(t, stub, args, expectedPayload)
	// Check if Tx was successful
	args = [][]byte{[]byte("getAccountTokens"), []byte("2")}
	expectedPayload = "100"
	checkInvokeResponse(t, stub, args, expectedPayload)

	// It should fail with empty string arg
	args = [][]byte{[]byte("getAccountTokens"), []byte("")}
	expectedMessage := "Argument at position 1 must be a non-empty string"
	checkInvokeResponseFail(t, stub, args, expectedMessage)

	// It should fail with more than one args
	args = [][]byte{[]byte("getAccountTokens"), []byte("1"), []byte("lol")}
	expectedMessage = "Incorrect number of arguments. Expecting account ID"
	checkInvokeResponseFail(t, stub, args, expectedMessage)
}
