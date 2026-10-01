#!/bin/bash

SERVER_HOST="localhost"
SERVER_PORT=8080
ACCRUAL_HOST="localhost"
ACCRUAL_PORT=8081
SERVER_BINARY=./cmd/gophermart/gophermart.exe
ACCRUAL_BINARY=./cmd/accrual/accrual_windows_amd64.exe
LOG_DIR=logs

if [ ! -e $LOG_DIR ] ; then
    mkdir ${LOG_DIR}
fi

if [ ! -d $LOG_DIR ] ; then
    echo "${LOG_DIR} is not a directory"
    exit 1
fi

if [ -z ${DATABASE_URI} ] ; then
    echo 'Need `export DATABASE_URI=your_db_url`'
    exit 1
fi

echo -n "Running tests	...	"
if ./gophermarttest-windows-amd64.exe \
            -test.v -test.run=^TestGophermart$ \
            -gophermart-binary-path=${SERVER_BINARY} \
            -gophermart-host=${SERVER_HOST} \
            -gophermart-port=${SERVER_PORT} \
            -gophermart-database-uri=${DATABASE_URI} \
            -accrual-binary-path=${ACCRUAL_BINARY} \
            -accrual-host=${ACCRUAL_HOST} \
            -accrual-port=${ACCRUAL_PORT} \
            -accrual-database-uri=${DATABASE_URI} &> ${LOG_DIR}/gophermarttest.log
then
	echo -e "[ \033[32mOK\033[0m ]"
else
	echo -e "[\033[31mFAIL\033[0m]"
	exit 1
fi