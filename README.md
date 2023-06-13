## 本项目介绍

本项目基于go语言开发

这个项目可以比对两个数据库的schema并生成修改另一个数据库的SQL

本项目默认支持MySQL,MariaDB,不过这里没有做过完整的测试

本项目中的代码不包含diff schema的核心代码, 利用的是`vitess.io/vitess/go/vt/schemadiff`提供的方法

vitess 是YouTube开源的用于MySQL水平扩展的数据库集群系统。其GitHub地址为: [https://github.com/vitessio/vitess/](https://github.com/vitessio/vitess/)

## 构建

```
make
```
注意: 目前在Windows上构建有问题,暂不支持在Windows机器上使用

## 配置

```yaml
src_db: # 基于这里配置的db进行比对
  user: root
  password: 123456
  host: 172.16.0.99
  port: 3306
  db_name: schema_diff_1
dst_db: 
  user: root
  password: 123456
  host: 172.16.0.99
  port: 3306
  db_name: schema_diff_2
diff_conf: # 比对的行为配置
  ignore_character: false # 忽略字符集
  ignore_auto_increment: false # 忽略auto_increment
```

## 使用

1. 执行比对
```
./schema-diff -config ./config.yaml
```

2. 将生成的SQL保存到文件中

```
./schema-diff -config ./config.yaml -out-sql xxx.sql
```

## 其他
有任何问题,欢迎提issue
代码比较水,轻喷