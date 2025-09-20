#!/bin/bash
#
# Run this script inside of the directory it resides in.
cd $(dirname $(realpath $0))

fuzz() {
	echo "$1"
	go test -fuzz="$1" -fuzztime=30s
}

fuzz "FuzzVectorNth"
fuzz "FuzzVectorAssoc"
fuzz "FuzzVectorConj"
fuzz "FuzzTransientVectorNth"
fuzz "FuzzTransientVectorAssoc"
fuzz "FuzzTransientVectorConj"
