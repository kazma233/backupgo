# backupgo

定时将你的目录/文件压缩，然后上传到OSS上。支持前置/后置自定义命令。我自己用来备份数据库dump等数据

# 使用

1. 配置文件位置: `config/config.yml`，示例配置参考： `config/config.example.yml`
2. 启动脚本

    ``` bash
    cp script_example/rebuild.sh rebuild.sh
    chmod +x rebuild.sh
    ./rebuild.sh
    ```