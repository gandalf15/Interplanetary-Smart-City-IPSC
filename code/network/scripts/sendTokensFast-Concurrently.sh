#!/bin/bash

verifyResult () {
	if [ $1 -ne 0 ] ; then
		echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
    echo "////////////// ERROR !!! FAILED to execute sendTokensFast-Concurrently //////////////////////"
		echo
   		exit 1
	fi
}

chaincodeInvoke () {
	peer chaincode invoke --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_tokens -c "${PAYLOAD}" -C channel3
    res=$?
	verifyResult $res "Sending tokens fast concurrently"
}

for (( i = 0; i < 1000; ++i ))
do
    PAYLOAD='{"Args":["sendTokensFast", "1", "2", "1", "false"]}'
	1>/dev/null 2>&1 peer chaincode invoke --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_tokens -c "${PAYLOAD}" -C channel3 &
    # Run the function in subshells
	# chaincodeInvoke &
done
wait

# peer chaincode invoke --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_tokens -c '{"Args":["getAccountTokens", "1"]}' -C channel3