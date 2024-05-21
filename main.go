package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"schema-diff/conf"
	"schema-diff/db"

	_ "github.com/go-sql-driver/mysql"
	"vitess.io/vitess/go/mysql/collations"
	"vitess.io/vitess/go/vt/schemadiff"
	"vitess.io/vitess/go/vt/vtenv"
)

type DbConf struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DbName   string `yaml:"db_name"`
}

var (
	confFileName   string
	outSqlFileName string
)

func main() {
	fs := flag.NewFlagSet("global", flag.ExitOnError)
	fs.StringVar(&confFileName, "conf", "./config.yml", "配置文件位置")
	fs.StringVar(&outSqlFileName, "out-sql", "", "生成需要在目标库执行的SQL文件,不指定默认打印在标准输出,注意如果文件中有内容,那么该文件会被覆盖")
	fs.Parse(os.Args[1:])

	config, err := conf.NewConfFromFile(confFileName)
	if err != nil {
		fmt.Println("获取配置文件失败")
		panic(err)
	}

	srcDb, err := db.NewDB(config.SrcDbConf)
	if err != nil {
		fmt.Println("源库连接失败")
		panic(err)
	}

	dstDb, err := db.NewDB(config.DstDbConf)
	if err != nil {
		fmt.Println("目标库连接失败")
		panic(err)
	}

	srcTables, err := srcDb.GetAllTables()
	if err != nil {
		fmt.Println("获取源库所有表失败")
		panic(err)
	}

	diffHints := &schemadiff.DiffHints{
		StrictIndexOrdering: false,
	}
	if config.DiffConf.IgnoreAutoIncrement {
		diffHints.AutoIncrementStrategy = schemadiff.AutoIncrementIgnore
	}
	if config.DiffConf.IgnoreCharacter {
		diffHints.TableCharsetCollateStrategy = schemadiff.TableCharsetCollateIgnoreAlways
	}

	srcMysqlVersion, err := srcDb.GetMysqlVersion()
	if err != nil {
		fmt.Println("获取源库版本失败")
		panic(err)
	}

	srcOpt := vtenv.Options{
		MySQLServerVersion: srcMysqlVersion,
	}

	srcEnv, err := vtenv.New(srcOpt)
	if err != nil {
		panic(err)
	}
	collationID := collations.NewEnvironment(srcMysqlVersion).DefaultConnectionCharset()
	schemadiffEnv := schemadiff.NewEnv(srcEnv, collationID)

	alterTableSql := []string{}

	for _, tableName := range srcTables {
		srcSchema, err := srcDb.GetTableSchema(tableName)
		if err != nil {
			fmt.Printf("获取源表schema失败[%s]错误信息: %v\n", tableName, err)
			continue
		}

		dstSchema, err := dstDb.GetTableSchema(tableName)
		if err != nil {
			if errors.Is(err, db.ErrNoSuchTable) {
				fmt.Printf("目标库中不存在[%s]表\n", tableName)
				alterTableSql = append(alterTableSql, srcSchema+";")
			} else {
				fmt.Printf("获取目标表schema失败[%s]错误信息: %v\n", tableName, err.Error())
			}
			continue
		}

		diff, err := schemadiff.DiffCreateTablesQueries(schemadiffEnv, dstSchema, srcSchema, diffHints)
		if err != nil {
			fmt.Printf("对比表结构失败[%s]\n", tableName)
			continue
		}
		if diff.IsEmpty() {
			fmt.Printf("源库和目标库表结构一致[%s]\n", tableName)
			continue
		} else {
			fmt.Printf("源库和目标库表结构不一致[%s]\n", tableName)
			alterTableSql = append(alterTableSql, diff.StatementString()+";")
		}
	}

	if len(alterTableSql) == 0 {
		fmt.Println("两个库的schema完全一致")
		return
	}

	diffSqlStr := ""

	for _, sql := range alterTableSql {
		diffSqlStr = fmt.Sprintf("%s\n%s\n", diffSqlStr, sql)
	}

	if outSqlFileName == "" {
		fmt.Println("================以下为需要要在目标库执行的SQL================")
		fmt.Println(diffSqlStr)
	} else {
		f, err := os.OpenFile(outSqlFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			fmt.Printf("打开SQL文件失败[%s]: %+v", outSqlFileName, err)
			fmt.Printf("将SQL信息输出到终端\n%s", diffSqlStr)
			return
		}
		defer f.Close()
		fmt.Println(diffSqlStr)
		_, err = f.Write([]byte(diffSqlStr))
		// _, err = f.WriteString(diffSqlStr)
		if err != nil {
			fmt.Printf("写SQL文件失败[%s]: %+v", outSqlFileName, err)
		}
	}
}
