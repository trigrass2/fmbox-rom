#!/bin/bash

CC=arm-linux-gcc GOARCH=arm CGO_ENABLED=1 go build -o fm-arm .

