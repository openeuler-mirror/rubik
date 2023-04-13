#!/bin/bash
rm -f coverprofile.new
head -1 fuzz-test-newconfig/coverprofile > coverprofile.new
cat fuzz-test-newconfig/coverprofile | grep -v golang | grep -v vendor | grep -v Fuzz | grep -v "mode: set" >> coverprofile.new
go tool cover -html=coverprofile.new -o cover.html
python3 -m http.server 8081
