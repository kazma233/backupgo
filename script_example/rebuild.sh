#!/bin/bash

# 进程名称
PROCESS_NAME="backupgo"

# 日志文件路径
LOG_FILE="$PROCESS_NAME.log"

# 拉取最新代码
echo "Pull code change"
git pull

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

# 编译新的二进制文件
rm -f "$PROCESS_NAME"
echo "Remove old $PROCESS_NAME binary"
go build -o "$PROCESS_NAME"
echo "Compiled new $PROCESS_NAME binary"

# 删除旧的日志文件(如果存在)
rm -f "$LOG_FILE"

# 启动新的进程,将输出重定向到日志文件
nohup "./$PROCESS_NAME" >> "$LOG_FILE" 2>&1 &
echo "Started new $PROCESS_NAME process. Logs will be written to $LOG_FILE"