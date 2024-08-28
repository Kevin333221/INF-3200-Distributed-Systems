#!/bin/bash

python3 testscript.py '["c2-8:50124"]'

for i in $(seq 1 $1)
do
	echo $1
done
