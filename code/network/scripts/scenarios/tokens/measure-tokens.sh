#!/bin/bash

#echo "How many transactions?"
#read num_tx
#num_tx=$((num_tx/10))
num_tx=1
echo "Loop starts at number?"
read begin_num
((begin_num=begin_num+10))

echo "Loop ends at number?"
read end_num
((end_num=end_num+10))

./sendTokensFast-Concurrently-warming.sh

for (( i = begin_num; i < end_num; ++i)); do
	# { time ./sendTokensFast-Concurrently.sh -b $i -e $((i+num_tx)) ; } 2>>$((num_tx*10))tx-measured_times_safe
	time ./sendTokensFast-Concurrently.sh -b $i -e $((i+num_tx))
done
