#!/bin/bash

echo
echo " ____    _____      _      ____    _____ "
echo "/ ___|  |_   _|    / \    |  _ \  |_   _|"
echo "\___ \    | |     / _ \   | |_) |   | |  "
echo " ___) |   | |    / ___ \  |  _ <    | |  "
echo "|____/    |_|   /_/   \_\ |_| \_\   |_|  "
echo
echo "Build 3 channels network woth 4 peers"
echo
CHANNEL_NAME_BASE="$1"
DELAY="$2"
: ${CHANNEL_NAME_BASE:="channel"}
: ${DELAY:=0}
: ${TIMEOUT:="2"}
COUNTER=1
MAX_RETRY=5
ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem

echo "Channel name : "$CHANNEL_NAME_BASE"1"
echo "Channel name : "$CHANNEL_NAME_BASE"2"
echo "Channel name : "$CHANNEL_NAME_BASE"3"

# verify the result of the end-to-end test
verifyResult () {
	if [ $1 -ne 0 ] ; then
		echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
    echo "========= ERROR !!! FAILED to execute End-2-End Scenario ==========="
		echo
   		exit 1
	fi
}

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

createChannel() {
	setGlobals 0
	
  	if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		peer channel create -o orderer.zak.codes:7050 -c $CHANNEL_NAME -f ./channel-artifacts/${CHANNEL_NAME}.tx >&log.txt
	else
		peer channel create -o orderer.zak.codes:7050 -c $CHANNEL_NAME -f ./channel-artifacts/${CHANNEL_NAME}.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Channel creation failed"
	echo "===================== Channel \"$CHANNEL_NAME\" is created successfully ===================== "
	echo
}

updateAnchorPeers() {
  PEER=$1
  setGlobals $PEER

  if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		peer channel update -o orderer.zak.codes:7050 -c $CHANNEL_NAME -f ./channel-artifacts/${CHANNEL_NAME}${CORE_PEER_LOCALMSPID}anchors.tx >&log.txt
	else
		peer channel update -o orderer.zak.codes:7050 -c $CHANNEL_NAME -f ./channel-artifacts/${CHANNEL_NAME}${CORE_PEER_LOCALMSPID}anchors.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Anchor peer update failed"
	echo "===================== Anchor peers for org \"$CORE_PEER_LOCALMSPID\" on \"$CHANNEL_NAME\" is updated successfully ===================== "
	sleep $DELAY
	echo
}

## Sometimes Join takes time hence RETRY atleast for 5 times
joinWithRetry () {
	peer channel join -b $CHANNEL_NAME.block  >&log.txt
	res=$?
	cat log.txt
	if [ $res -ne 0 -a $COUNTER -lt $MAX_RETRY ]; then
		COUNTER=` expr $COUNTER + 1`
		echo "PEER$1 failed to join the channel, Retry after 2 seconds"
		sleep $DELAY
		joinWithRetry $1
	else
		COUNTER=1
	fi
  verifyResult $res "After $MAX_RETRY attempts, PEER$ch has failed to Join the Channel"
}

joinChannel () {
	for peer in 0 1 2 3; do
		setGlobals $peer
		joinWithRetry $peer
		echo "===================== PEER$peer joined on the channel \"$CHANNEL_NAME\" ===================== "
		sleep $DELAY
		echo
	done
}

installChaincode () {
	PEER=$1
	CC_NAME=$2
	PATH_TO_CC=$3
	setGlobals $PEER
	peer chaincode install -n $CC_NAME -v 1.0 -p $PATH_TO_CC >&log.txt
	res=$?
	cat log.txt
        verifyResult $res "Chaincode installation on remote peer PEER$PEER has Failed"
	echo "===================== Chaincode $CC_NAME is installed on remote peer PEER$PEER ===================== "
	echo
}

instantiateChaincode () {
	PEER=$1
	CC_NAME=$2
	# PP=$3	# policy and payload code for channel number
	setGlobals $PEER
	echo $PP
	# while 'peer chaincode' command can get the orderer endpoint from the peer (if join was successful),
	# lets supply it directly as we know it using the "-o" option

	if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		peer chaincode instantiate -o orderer.zak.codes:7050 -C $CHANNEL_NAME -n $CC_NAME -v 1.0 -c "${PAYLOAD}" -P "${POLICY}" >&log.txt
	else
		peer chaincode instantiate -o orderer.zak.codes:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME -v 1.0 -c "${PAYLOAD}" -P "${POLICY}" >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Chaincode instantiation on PEER$PEER on channel '$CHANNEL_NAME' failed"
	echo "===================== Chaincode Instantiation on PEER$PEER on channel '$CHANNEL_NAME' is successful ===================== "
	echo
}

chaincodeQuery () {
  PEER=$1
  CC_NAME=$2
  EXPECTED_VALUE=$4
  echo "===================== Querying on PEER$PEER on channel '$CHANNEL_NAME' and chaincode '$CC_NAME' ... ===================== "
  setGlobals $PEER
  local rc=1
  local starttime=$(date +%s)

  # continue to poll
  # we either get a successful response, or reach TIMEOUT
  while test "$(($(date +%s)-starttime))" -lt "$TIMEOUT" -a $rc -ne 0
  do
     sleep $DELAY
     echo "Attempting to Query PEER$PEER ...$(($(date +%s)-starttime)) secs"
     peer chaincode query -C $CHANNEL_NAME -n $CC_NAME -c "${PAYLOAD}" >&log.txt
     test $? -eq 0 && VALUE=$(cat log.txt | awk '/Query Result/ {print $NF}')
     test "$VALUE" = "${EXPECTED_VALUE}" && let rc=0
  done
  echo
  cat log.txt
  if test $rc -eq 0 ; then
	echo "===================== Query on PEER$PEER on channel '$CHANNEL_NAME' and chaincode '$CC_NAME' is successful ===================== "
  else
	echo "!!!!!!!!!!!!!!! Query result on PEER$PEER is INVALID !!!!!!!!!!!!!!!!"
        echo "================== ERROR !!! FAILED to execute End-2-End Scenario =================="
	echo
	exit 1
  fi
}

chaincodeInvoke () {
	PEER=$1
	CC_NAME=$2
	setGlobals $PEER
	# while 'peer chaincode' command can get the orderer endpoint from the peer (if join was successful),
	# lets supply it directly as we know it using the "-o" option
	if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		peer chaincode invoke -o orderer.zak.codes:7050 -C $CHANNEL_NAME -n $CC_NAME -c "${PAYLOAD}" >&log.txt
	else
		peer chaincode invoke -o orderer.zak.codes:7050  --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME -c "${PAYLOAD}" >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Invoke execution on PEER$PEER failed "
	# Extract the returned payload without quotes
	RETURNED_PAYLOAD=$(cat log.txt | awk -F"payload:" '{print $2}')
    RETURNED_PAYLOAD=$(echo $RETURNED_PAYLOAD | awk -F">" '{print $1}')
	echo "===================== Invoke transaction on PEER$PEER on channel '$CHANNEL_NAME' and chaincode '$CC_NAME' is successful ===================== "
	echo
}

## Create channels
echo "Creating channels..."
for i in 1 2 3; do
	CHANNEL_NAME="${CHANNEL_NAME_BASE}${i}"
	createChannel
	echo "===================== the channel \"$CHANNEL_NAME\" is created ===================== "
	echo
done

## Join all the peers to the channel
echo "Having all peers join the channel..."
CHANNEL_NAME="${CHANNEL_NAME_BASE}1"
for i in 0 1; do
	setGlobals $i
	joinWithRetry $i
	echo "===================== PEER$i joined on the channel \"$CHANNEL_NAME\" ===================== "
	sleep $DELAY
	echo
done

for i in 2 3; do
	CHANNEL_NAME="${CHANNEL_NAME_BASE}${i}"
	joinChannel
done

## Set the anchor peers for each org in the channel
CHANNEL_NAME="${CHANNEL_NAME_BASE}1"
echo "Updating anchor peers for City1 for channel '${CHANNEL_NAME}' ..."
updateAnchorPeers 0
for i in 2 3; do
	CHANNEL_NAME="${CHANNEL_NAME_BASE}${i}"
	echo "Updating anchor peers for City1 for channel '${CHANNEL_NAME}' ..."
	updateAnchorPeers 0
	echo "Updating anchor peers for City2 for channel '${CHANNEL_NAME}' ..."
	updateAnchorPeers 2
done

## Install chaincode_data on Peer0/City1 and Peer2/City2
echo "--> Installing chaincode_data on Peer0/City1..."
echo
installChaincode 0 chaincode_data github.com/hyperledger/fabric/chaincode/chaincode_data

# Install chaincode_ad on Peer0/City1...
echo "--> Installing chaincode_ad on Peer0/City1..."
echo
installChaincode 0 chaincode_ad github.com/hyperledger/fabric/chaincode/chaincode_ad

# Install chaincode_ad on Peer2/City2...
echo "--> Install chaincode_ad on Peer2/City2..."
echo
installChaincode 2 chaincode_ad github.com/hyperledger/fabric/chaincode/chaincode_ad

# Install chaincode_tokens on Peer0/City1...
echo "--> Installing chaincode_tokens on Peer0/City1..."
echo
installChaincode 0 chaincode_tokens github.com/hyperledger/fabric/chaincode/chaincode_tokens

# Install chaincode_tokens on Peer2/City2...
echo "--> Installing chaincode_tokens on Peer2/City2..."
echo
installChaincode 2 chaincode_tokens github.com/hyperledger/fabric/chaincode/chaincode_tokens

#Instantiate chaincode_data on Peer0/City1
echo "--> Instantiating chaincode_data on Peer0/City1..."
echo
CHANNEL_NAME="${CHANNEL_NAME_BASE}1"
PAYLOAD='{"Args":["2", "1000"]}'
POLICY="AND ('City1MSP.member')"
instantiateChaincode 0 chaincode_data

#Instantiate chaincode_ad on Peer2/City2
echo "--> Instantiating chaincode_ad on Peer2/City2..."
echo
CHANNEL_NAME="${CHANNEL_NAME_BASE}2"
PAYLOAD='{"Args":["2", "1000"]}'
POLICY="OR ('City1MSP.member','City2MSP.member')"
instantiateChaincode 2 chaincode_ad

#Instantiate chaincode_tokens on Peer2/City2
echo "--> Instantiating chaincode_tokens on Peer2/City2..."
echo
CHANNEL_NAME="${CHANNEL_NAME_BASE}3"
PAYLOAD='{"Args":["100000"]}'
POLICY="OR ('City1MSP.member','City2MSP.member')"
instantiateChaincode 2 chaincode_tokens

# Create global variable RETURNED_PAYLOAD that is used in chaincodeInvoke function
RETURNED_PAYLOAD=""

# Invoke on chaincode_tokens on Peer0/City1
echo "--> Sending invoke transaction createAccount on Peer0/City1 on chaincode_tokens"
echo
sleep 2
CHANNEL_NAME="${CHANNEL_NAME_BASE}3"
PAYLOAD='{"Args":["createAccount", "2", "Test_Account"]}'
chaincodeInvoke 0 chaincode_tokens

# Invoke on chaincode_data on Peer0/City1
echo "--> Sending invoke first transaction createData on City1/peer0 on chaincode_data ..."
echo
CHANNEL_NAME="${CHANNEL_NAME_BASE}1"
PAYLOAD='{"Args":["createData", "1", "test data", "50", "Celsius", "20180321163750", "marcel"]}'
chaincodeInvoke 0 chaincode_data

# Invoke on chaincode_ad on Peer0/City1
# Free data
echo "--> Sending invoke first transaction createDataEntryAd on Peer0/City1 on chaincode_ad ..."
echo
CHANNEL_NAME="${CHANNEL_NAME_BASE}2"
PAYLOAD='{"Args":["createDataEntryAd", "1", "test data", "50", "Celsius", "20180321163750", "marcel", "0", "2"]}'
chaincodeInvoke 0 chaincode_ad

# Invoke on chaincode_data on Peer0/City1
echo "--> Sending invoke second transaction createDate on City1/peer0 on chaincode_data ..."
echo
CHANNEL_NAME="${CHANNEL_NAME_BASE}1"
PAYLOAD='{"Args":["createData", "2", "test data", "100", "Celsius", "20180321160000", "marcel"]}'
chaincodeInvoke 0 chaincode_data

# Invoke on chaincode_ad on Peer0/City1
# Paid data 10
echo "--> Sending invoke second transaction createDataEntryAd on Peer0/City1 on chaincode_ad ..."
echo
CHANNEL_NAME="${CHANNEL_NAME_BASE}2"
PAYLOAD='{"Args":["createDataEntryAd", "2", "test data", "???", "Celsius", "20180321160000", "marcel", "10", "2"]}'
chaincodeInvoke 0 chaincode_ad

# Invoke on chaincode_tokens on Peer0/City1
echo "--> Sending invoke transaction sendTokensSafe on Peer0/City1 on chaincode_tokens"
echo
sleep 3
CHANNEL_NAME="${CHANNEL_NAME_BASE}3"
PAYLOAD='{"Args":["sendTokensSafe", "1", "2", "10", "true"]}'
chaincodeInvoke 0 chaincode_tokens

# Invoke on chaincode_ad on Peer0/City1
echo "--> Sending invoke transaction revealPaidData on Peer0/City1 on chaincode_ad"
echo "Using TxID from previous invocation:"
echo $RETURNED_PAYLOAD
echo
sleep 3	# required sleep to wait for previous data to commit and being available
CHANNEL_NAME="${CHANNEL_NAME_BASE}2"
PAYLOAD='{"Args":["revealPaidData", "channel1", "chaincode_data", "2", "20180321160000", "channel3", "chaincode_tokens", '
PAYLOAD=$PAYLOAD$RETURNED_PAYLOAD
PAYLOAD="${PAYLOAD}]}"
chaincodeInvoke 0 chaincode_ad

# Invoke on chaincode_tokens on Peer0/City1
echo "--> Sending invoke transaction sendTokensSafe on Peer0/City1 on chaincode_tokens"
echo
sleep 3
CHANNEL_NAME="${CHANNEL_NAME_BASE}3"
PAYLOAD='{"Args":["sendTokensSafe", "2", "1", "10", "false"]}'
chaincodeInvoke 0 chaincode_tokens


# Invoke on chaincode on Peer0/City1
#echo "--> Sending invoke transaction queryAccountByName on Peer0/City1 on chaincode_tokens"
#echo
#CHANNEL_NAME="${CHANNEL_NAME_BASE}3"
#PAYLOAD='{"Args":["queryAccountByName", "Init_Account"]}'
#chaincodeInvoke 0 chaincode_tokens


# peer chaincode invoke --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_ad -c '{"Args":["revealPaidData", "channel1", "chaincode_data", "2", "20180321160000", "channel3", "chaincode_tokens", "txID"]}' -C channel2

#Query on chaincode on Peer0/Org1
# echo "Querying chaincode on org1/peer0..."
# chaincodeQuery 0 100

## Install chaincode on Peer3/Org2
# echo "Installing chaincode on org2/peer3..."
# installChaincode 3

#Query on chaincode on Peer3/Org2, check if the result is 90
# echo "Querying chaincode on org2/peer3..."
# chaincodeQuery 3 90

echo
echo "========= All GOOD, 3 channels network is created =========== "
echo

echo
echo " _____   _   _   ____   "
echo "| ____| | \ | | |  _ \  "
echo "|  _|   |  \| | | | | | "
echo "| |___  | |\  | | |_| | "
echo "|_____| |_| \_| |____/  "
echo

exit 0
