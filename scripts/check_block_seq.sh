#!/bin/bash
exec 3<"1"
exec 4<"2"
while read line1<&3 && read line2<&4
do
        echo $line1 $line2
echo $((line2))
echo $((line1))
if [ $((line2)) -ne $((line1+1)) ]; then
	exit -1
fi
done
