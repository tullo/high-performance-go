#!/usr/bin/env bash
set -e
# echo $@ 'filtering lines with (PCDATA | FUNCDATA)'
go build -gcflags=-S 2>&1 $@ | grep -v PCDATA | grep -v FUNCDATA | less
