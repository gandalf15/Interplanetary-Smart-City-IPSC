#!/bin/bash
# This commands are executed inside CLI container immediately when started
# It creates blocks for channels 1,2,3 and then it broadcasts them to orderer and connect the channels
ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem
CHANNEL_NAME="channel2"
setGlobals () {

	if [ $1 -eq 0 -o $1 -eq 1 ] ; then
		CORE_PEER_LOCALMSPID="City1MSP"
		CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/city1.zak.codes/peers/peer0.city1.zak.codes/tls/ca.crt
		CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/city1.zak.codes/users/Admin@city1.zak.codes/msp
		if [ $1 -eq 0 ]; then
			CORE_PEER_ADDRESS=peer0.city1.zak.codes:7051
		else
			CORE_PEER_ADDRESS=peer1.city1.zak.codes:7051
			CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/city1.zak.codes/users/Admin@city1.zak.codes/msp
		fi
	else
		CORE_PEER_LOCALMSPID="City2MSP"
		CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/city2.zak.codes/peers/peer0.city2.zak.codes/tls/ca.crt
		CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/city2.zak.codes/users/Admin@city2.zak.codes/msp
		if [ $1 -eq 2 ]; then
			CORE_PEER_ADDRESS=peer0.city2.zak.codes:7051
		else
			CORE_PEER_ADDRESS=peer1.city2.zak.codes:7051
		fi
	fi

	env |grep CORE
}

joinChannel () {
    CHANNEL_NAME = 
	for ch in 0 1 2 3; do
		setGlobals $ch
		peer channel join -b $CHANNEL_NAME.block
		echo "===================== PEER$ch joined on the channel \"$CHANNEL_NAME\" ===================== "
		sleep 1
		echo
	done
}

installChaincode () {
	PEER=$1
	setGlobals $PEER
	peer chaincode install -n cimple_chaincode -v 1.0 -p github.com/hyperledger/fabric/chaincode/simple_chaincode
}

instantiateChaincode () {
	PEER=$1
	setGlobals $PEER
	# while 'peer chaincode' command can get the orderer endpoint from the peer (if join was successful),
	# lets supply it directly as we know it using the "-o" option
	peer chaincode instantiate -o orderer.zak.codes:7050 --tls true --cafile $ORDERER_CA -C $CHANNEL_NAME -n simple_chaincode -v 1.0 -c '{"Args":["init","a","100","b","200"]}' -P "OR	('City1MSP.member','City2MSP.member')"
}

peer channel create -o orderer.zak.codes:7050 -c channel1 -f ./channel-artifacts/channel1.tx --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem
sleep 1
peer channel create -o orderer.zak.codes:7050 -c channel2 -f ./channel-artifacts/channel2.tx --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem
sleep 1
peer channel create -o orderer.zak.codes:7050 -c channel3 -f ./channel-artifacts/channel3.tx --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem
sleep 1

# CHANNEL_NAME="channel1"
# setGlobals 0
# peer channel join -b channel1.block
# setGlobals 1
# peer channel join -b channel1.block

CHANNEL_NAME="channel2"
joinChannel
# CHANNEL_NAME="channel3"
# joinChannel

installChaincode 0
installChaincode 2

instantiateChaincode 2


#peer channel join -b channel1.block
#sleep 2
#peer channel join -b channel2.block
#sleep 2
#peer channel join -b channel3.block
#sleep2
# peer chaincode install -n simple_chaincode -v 1.0 -p github.com/hyperledger/fabric/chaincode/simple_chaincode
# sleep 1
# peer chaincode instantiate -o orderer.zak.codes:7050 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -C channel1 -n simple_chaincode -v 1.0 -c '{"Args":["init","a", "100"]}' -P "OR ('City1MSP.member','City2MSP.member')"
# sleep 1
# peer chaincode invoke --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n simple_chaincode -c '{"Args":["putDataEntry", "1", "test data", "50", "Celsius", "20180321163750", "marcel"]}' -C channel2
# sleep 1
# peer chaincode query -n simple_chaincode -c '{"Args":["getDataEntryById","1"]}' -C channel2
# while true; do sleep 1000; done


# peer chaincode instantiate -o orderer.zak.codes:7050 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -C channel1 -n mycc -v 1.0 -c '{"Args":["init","a", "100", "b","200"]}' -P "OR ('Org1MSP.member','Org2MSP.member')"
# peer chaincode query -C channel1 -n mycc -c '{"Args":["query","a"]}'
# peer chaincode invoke -o orderer.zak.codes:7050 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -C channel1 -n mycc -c '{"Args":["invoke","a","b","10"]}'