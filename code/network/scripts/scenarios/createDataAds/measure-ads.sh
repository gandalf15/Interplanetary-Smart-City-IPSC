#!/bin/bash

echo "How many transactions?"
read num_tx
num_tx=$((num_tx/10))

echo "Loop starts at number?"
read begin_num

echo "Loop ends at number?"
read end_num

./createDataAds-Concurrently-warming.sh

for (( i = begin_num; i < end_num; ++i)); do
	{ time ./createDataAds-Concurrently.sh -b $i -e $((i+num_tx)) ; } 2>>$((num_tx*10))tx-measured_times
	#time ./createDataAds-Concurrently.sh -b $i -e $((i+num_tx))
done
