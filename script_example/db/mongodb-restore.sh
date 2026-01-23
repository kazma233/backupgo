mongorestore --host=<地址> --port=<端口> --username=<用户> --password=<密码> --authenticationDatabase=admin --drop --db=<目标数据库名> <备份目录路径>
# 5.0后新用法
mongorestore --host=<地址> --port=<端口> --username=<用户> --password=<密码> --authenticationDatabase=admin --drop --nsInclude=<目标数据库名>.* <备份目录路径>