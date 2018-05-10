#!/bin/sh
echo "start program"
touch my_logs.log
/go/src/github.com/jetuuuu/youtube2audio/main -addr=$CONSUL_ADDR