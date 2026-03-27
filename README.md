# backupgo

定时将你的目录/文件压缩，然后上传到OSS上。支持前置/后置自定义命令，也支持内置 Postgres 和 MongoDB 备份源。

# 使用

## 配置文件

位置: `config.yml`，示例配置参考如下：

```yaml
notice:
  mail:
    smtp: 'smtp'
    port: 456
    user: 'user'
    password: 'password'
    to:
      - 'notice@example.com'
  telegram:
    bot_token: '123456:ABCDEF'
    # 私聊/群组建议填写数字 chat_id；公开频道可填写 @channel_username
    chat_id: '@your_channel'

oss:
  bucket_name: 'bucket'
  endpoint: 'endpoint'
  fast_endpoint: 'fast_endpoint'
  access_key: 'access_key'
  access_key_secret: 'access_key_secret'

backup:
  - id: 'app1'
    before_command: 'docker cp xxx:/app/data/ ./export'
    type: 'path'
    backup_path: './export'
    after_command: 'rm -rf ./export'
    backup_task: '0 25 0 * * ?'

  - id: 'postgres_prod'
    type: 'postgres'
    backup_task: '0 40 0 * * ?'
    postgres:
      mode: 'docker'
      container: 'postgres'
      user: 'postgres'
      password: 'password'
      databases:
        - 'app'
        - 'analytics'
      extra_args:
        - '--no-owner'

  - id: 'mongodb_prod'
    type: 'mongodb'
    backup_task: '0 55 0 * * ?'
    mongodb:
      mode: 'docker'
      container: 'mongo'
      username: 'root'
      password: 'password'
      auth_database: 'admin'
      gzip: true
      databases:
        - 'app'
        - 'logs'
```

## 启动脚本

``` bash
# 构建
go build -o backupgo .

# 前台启动调度器
./backupgo start

# 后台启动调度器
./backupgo start -d

# 停止后台进程
./backupgo stop

# 查看调度器状态和备份任务列表
./backupgo status

# 手动执行指定备份
./backupgo backup <backup-id>

# 查看帮助
./backupgo --help
```

# 配置说明

**notice**

- 顶层 `notice` 可选，统一放通知相关配置。
- `notice.mail` 可选；启用邮件通知时，需要填写 `smtp`、`port`、`user`、`password`、`to`。
- `notice.telegram` 可选；启用 Telegram 通知时，需要填写 `bot_token` 和 `chat_id`。
- 如果是 bot 私聊发给你自己，`chat_id` 通常填写你自己的数字 ID，并且你需要先给 bot 发送一次 `/start`。
- 如果是 bot 往群组或超级群发消息，建议优先使用群组数字 `chat_id`，常见格式如 `-1001234567890`。
- 如果是 bot 往公开频道发消息，可以直接使用频道用户名，例如 `@your_channel`。

**oss**

- 顶层 `oss` 必填。
- `bucket_name` 必填。
- `endpoint` 必填。
- `access_key` 必填。
- `access_key_secret` 必填。
- `fast_endpoint` 可选；不填时只使用主上传地址。

**backup**

- 顶层 `backup` 必填，至少需要定义一个任务。
- `backup` 每一项表示一个备份任务。
- 每个备份任务都必须填写唯一的 `id`。
- 通用字段 `type` 可选，支持 `path`、`postgres`、`mongodb`，默认是 `path`。
- 通用字段 `backup_task` 可选，默认是 `0 25 0 * * ?`。
- 通用字段 `before_command` 可选，在备份开始前执行。
- 通用字段 `after_command` 可选，在压缩完成后执行。
- 一个任务只能配置一种备份源，不能同时配置 `backup_path`、`postgres`、`mongodb`。

**backup.path**

- 适用于 `type: path`。
- `backup_path` 必填。
- 这个模式保持兼容，仍然是压缩指定目录后上传。
- 适合已有脚本、已有导出目录、已有 `docker cp` 流程的场景。

**backup.postgres**

- 适用于 `type: postgres`。
- `postgres` 节点必填。
- `postgres.databases` 必填。
- `postgres.container` 仅在 `postgres.mode: docker` 时必填。
- `postgres.mode` 可选，默认 `local`，可选值为 `local` 或 `docker`。
- `postgres.host` 可选。
- `postgres.port` 可选。
- `postgres.user` 可选。
- `postgres.password` 可选。
- `postgres.extra_args` 可选。
- 内置模式会先执行 `pg_dump` 导出到临时目录，再复用现有压缩和上传逻辑。
- 依赖 `pg_dump`；如果 `postgres.mode: docker`，则依赖宿主机可执行 `docker`，并要求容器内可执行 `pg_dump`。
- 不填 `postgres.password` 时仍可备份，但仅适用于非交互免密场景，因为程序固定使用 `pg_dump --no-password`。
- Postgres 免密常见场景包括 `peer` / `trust` 认证、同机或同容器 Unix Socket、已配置 `.pgpass` / `PGPASSFILE`。
- 如果是远程 Postgres，且服务端要求 `md5` / `scram` 密码认证，则需要显式配置 `postgres.password`。

**backup.mongodb**

- 适用于 `type: mongodb`。
- `mongodb` 节点必填。
- `mongodb.databases` 必填。
- `mongodb.container` 仅在 `mongodb.mode: docker` 时必填。
- `mongodb.mode` 可选，默认 `local`，可选值为 `local` 或 `docker`。
- `mongodb.uri` 可选。
- `mongodb.host` 可选。
- `mongodb.port` 可选。
- `mongodb.username` 可选。
- `mongodb.password` 可选。
- `mongodb.auth_database` 可选。
- `mongodb.gzip` 可选，默认 `false`。
- `mongodb.extra_args` 可选。
- 如果未设置 `mongodb.uri`，且填写了 `mongodb.username`，则必须同时填写 `mongodb.password`。
- 内置模式会执行 `mongodump --archive` 导出到临时目录，再复用现有压缩和上传逻辑。
- 依赖 `mongodump`；如果 `mongodb.mode: docker`，则依赖宿主机可执行 `docker`，并要求容器内可执行 `mongodump`。
- `mongodb.password` 是可选的；如果服务端允许无认证访问，或者你通过 `mongodb.uri` 使用了不依赖密码的认证方式，可以不填。
- 如果 MongoDB 开启了用户名密码认证，并且未通过 `mongodb.uri` 提供其他认证方式，则需要填写 `mongodb.username` 和 `mongodb.password`。

# 恢复示例

先从 OSS 下载对应的备份 zip 文件并解压。内置数据库备份解压后通常会得到这样的文件：

- Postgres: `<backup-id>/<database>.dump`
- MongoDB: `<backup-id>/<database>.archive` 或 `<backup-id>/<database>.archive.gz`

例如任务 ID 为 `postgres_prod` / `mongodb_prod`，解压后可能得到：

```text
postgres_prod/app.dump
mongodb_prod/app.archive.gz
```

**恢复 Postgres**

内置 Postgres 备份使用的是 `pg_dump --format=custom`，所以恢复时应使用 `pg_restore`。

```bash
export PGPASSWORD='<密码>'

pg_restore \
  -h <主机> \
  -p <端口> \
  -U <用户> \
  -d <目标数据库> \
  --clean \
  --if-exists \
  --no-owner \
  ./postgres_prod/app.dump
```

如果目标数据库还不存在，可以先创建：

```bash
createdb -h <主机> -p <端口> -U <用户> <目标数据库>
```

**恢复 MongoDB**

内置 MongoDB 备份使用的是 `mongodump --archive`，恢复时应使用 `mongorestore --archive`。

未启用 `mongodb.gzip` 时：

```bash
mongorestore \
  --host <主机> \
  --port <端口> \
  --username <用户> \
  --password '<密码>' \
  --authenticationDatabase <认证库> \
  --drop \
  --archive=./mongodb_prod/app.archive
```

启用 `mongodb.gzip: true` 时：

```bash
mongorestore \
  --host <主机> \
  --port <端口> \
  --username <用户> \
  --password '<密码>' \
  --authenticationDatabase <认证库> \
  --drop \
  --gzip \
  --archive=./mongodb_prod/app.archive.gz
```

- `--drop` 表示恢复前先删除已存在的集合；如果你不希望覆盖现有数据，请去掉它。
- 如果你使用的是 `mongodb.uri` 方式连接，也可以直接写成 `mongorestore --uri '<连接串>' --drop --archive=...`；`gzip` 文件额外加上 `--gzip`。
