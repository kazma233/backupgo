DB_CONTAINER="your_container_name"; \
MONGO_USER=""; \
MONGO_PASS=""; \
MONGO_AUTH_DB="admin"; \
HOST_BACKUP_DIR="mongo_dump"; \
CONTAINER_DUMP_DIR="/tmp/db_dump"; \
DB_NAMES=("db1" "db2" "db3"); \

docker exec "${DB_CONTAINER}" mkdir -p "${CONTAINER_DUMP_DIR}" && mkdir -p "${HOST_BACKUP_DIR}"; \

for DB_NAME in "${DB_NAMES[@]}"; do \
    if [ -z "${MONGO_USER}" ]; then \
        docker exec "${DB_CONTAINER}" mongodump --db="${DB_NAME}" --out="${CONTAINER_DUMP_DIR}/${DB_NAME}"; \
    else \
        docker exec "${DB_CONTAINER}" mongodump --authenticationDatabase="${MONGO_AUTH_DB}" --db="${DB_NAME}" --username="${MONGO_USER}" --password="${MONGO_PASS}" --out="${CONTAINER_DUMP_DIR}/${DB_NAME}"; \
    fi; \
done; \

docker cp "${DB_CONTAINER}:${CONTAINER_DUMP_DIR}" "${HOST_BACKUP_DIR}" && \
docker exec "${DB_CONTAINER}" rm -rf "${CONTAINER_DUMP_DIR}" && \
echo "数据库 ${DB_NAMES[*]} 已成功备份到宿主机: ${HOST_BACKUP_DIR}";