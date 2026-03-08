# backupgo

定时将你的目录/文件压缩，然后上传到OSS上。支持前置/后置自定义命令。我自己用来备份数据库dump等数据

# 使用

1. 配置文件位置: `config.yml`，示例配置参考:

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
      app1:
        before_command: 'docker cp xxx:/app/data/ ./export'
        back_path: './export'
        after_command: 'rm -rf ./export'
        backup_task: '0 25 0 * * ?'
    ```

    - 通知配置统一放在 `notice` 下，邮件使用 `notice.mail`，Telegram 使用 `notice.telegram`
    - Telegram 需要填写 `notice.telegram.bot_token` 和 `notice.telegram.chat_id`
       - 如果是 bot 私聊发给你自己，通常填写你自己的数字 `chat_id`，并且你需要先给 bot 发送一次 `/start`
       - 如果是 bot 往群组或超级群发消息，建议优先使用群组的数字 `chat_id`，常见格式如 `-1001234567890`
       - 如果是 bot 往公开频道发消息，可以直接使用频道用户名，例如 `@your_channel`
2. 启动脚本

    ``` bash
    cp script_example/rebuild.sh rebuild.sh
    chmod +x rebuild.sh
    ./rebuild.sh
    ```
