#!/bin/sh
name="proxy"
set GOARCH=adm64
set GOSS=linux
mkdir -p output/log
go build -o output
chmod +x output/${name}
exec output/${name}