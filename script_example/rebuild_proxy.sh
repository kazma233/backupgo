#!/bin/bash

go env -w GOPROXY=https://goproxy.cn,direct
./rebuild.sh