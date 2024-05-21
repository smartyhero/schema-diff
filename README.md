## 本项目介绍

本项目基于go语言开发

这个项目可以比对两个数据库的schema并生成修改另一个数据库的SQL

本项目算是一个玩具项目, 没有经过大量测试, 不过代码非常简单, 大家可以直接看代码

比对的结果, 一定一定一定要检查没问题后再在目标库执行

本项目默认支持MySQL,MariaDB,不过这里没有做过完整的测试

本项目中的代码不包含diff schema的核心代码, 利用的是`vitess.io/vitess/go/vt/schemadiff`提供的方法

vitess 是YouTube开源的用于MySQL水平扩展的数据库集群系统。其GitHub地址为: [https://github.com/vitessio/vitess/](https://github.com/vitessio/vitess/)

## 构建

```
make
```

## 配置

```yaml
src_db: # 基于这里配置的db进行比对
  dsn: "root:123456789@tcp(127.0.0.1:3306)/tmp?charset=utf8mb4&parseTime=True"
dst_db:
  dsn: "root:123456789@tcp(127.0.0.1:3306)/bill_analysis?charset=utf8mb4&parseTime=True" 
diff_conf: # 比对的行为配置
  ignore_character: false  # 忽略字符集
  ignore_auto_increment: true # 忽略auto_increment
save_sql_path: ./xxx.sql
```

配置中支持环境变量, 以下写法, `${SAVE_SQL_PATH}`会被解析成真实的值,(通过`os.ExpandEnv()`实现)

```yaml
src_db: # 基于这里配置的db进行比对
  dsn: "root:123456789@tcp(127.0.0.1:3306)/tmp?charset=utf8mb4&parseTime=True"
dst_db:
  dsn: "root:123456789@tcp(127.0.0.1:3306)/bill_analysis?charset=utf8mb4&parseTime=True" 
diff_conf: # 比对的行为配置
  ignore_character: false  # 忽略字符集
  ignore_auto_increment: true # 忽略auto_increment
save_sql_path: ${SAVE_SQL_PATH}
```

## 使用

1. 执行比对
```
./schema-diff -conf ./config.yaml
```

2. 将生成的SQL保存到文件中

```
./schema-diff -conf ./config.yaml -save-sql xxx.sql
```

3. 指定连接串, -src-dsn等参数优先级高于配置文件
```
./schema-diff -conf ./config.yaml -src-dsn "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
```

## 其他
有任何问题,欢迎提issue