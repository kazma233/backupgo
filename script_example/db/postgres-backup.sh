DB_CONTAINER="your_container_name"; \
DB_USER="user"; \
HOST_BACKUP_DIR="postgres_dump"; \
DB_NAMES=("db1" "db2" "db3"); \

mkdir -p "${HOST_BACKUP_DIR}"; \

# 在 Linux 环境下，当客户端（pg_dump）和服务器（postgres 进程）位于同一台机器或同一容器内时，它们会优先通过 Unix Domain Socket 进行通信，而不是走标准的 TCP/IP 端口
for DB_NAME in "${DB_NAMES[@]}"; do \
    HOST_BACKUP_FILE="${HOST_BACKUP_DIR}/${DB_NAME}_$(date +%Y%m%d_%H%M%S).dump"; \
    docker exec "${DB_CONTAINER}" pg_dump -U "${DB_USER}" -d "${DB_NAME}" -Fc > "${HOST_BACKUP_FILE}"; \
    echo "数据库 ${DB_NAME} 已备份到: ${HOST_BACKUP_FILE}"; \
done; \

echo "所有数据库 ${DB_NAMES[*]} 已成功备份到宿主机: ${HOST_BACKUP_DIR}";