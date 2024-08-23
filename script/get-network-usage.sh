#!/bin/bash

process_index=$1
printf -v class_id "1:%x" "$process_index"

usage=$(tc -s -d -g  class show dev eno1 | grep -w  -A 2 "${class_id}" | grep "Sent" |  sed 's@^[^0-9]*\([0-9]\+\).*@\1@')

echo "${usage}"