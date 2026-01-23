#!/bin/bash

# 进程名称
PROCESS_NAME="backupgo"

# 查找进程 ID
PID=$(lsof -t -c "$PROCESS_NAME")

# 如果进程存在
if [ -n "$PID" ]; then
  # 杀死进程
  kill -9 $PID
  echo "Killed $PROCESS_NAME process with PID $PID"
else
  echo "No $PROCESS_NAME process found"
fi