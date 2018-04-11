#!/bin/bash

# verify the result
verifyResult () {
	if [ $1 -ne 0 ] ; then
		echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
		echo
   		exit 1
	fi
}


# To generate cover profile for the test file
go test -coverprofile=coverage.out --tags nopkcs11
res=$?
verifyResult $res "Test failed"
# In order to see what part of the code was / was not covered use this command
go tool cover -html=coverage.out
