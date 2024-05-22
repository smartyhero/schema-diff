package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"schema-diff/conf"
	"schema-diff/diffsql"

	vtconfig "vitess.io/vitess/go/mysql/config"
)

var (
	confFileName    string
	SaveSqlFileName string
	srcDbDsn        string
	dstDbDsn        string
	srcSqlFile      string
	dstSqlFile      string
	MysqlVersion    string
	skipTables      string
	config          *conf.Conf
)

func init() {
	fs := flag.NewFlagSet("global", flag.ExitOnError)
	fs.StringVar(&confFileName, "conf", "./config.yml", "配置文件位置")
	fs.StringVar(&SaveSqlFileName, "save-sql", "", "生成需要在目标库执行的SQL文件,不指定默认打印在标准输出,注意如果文件中有内容,那么该文件会被覆盖")
	fs.StringVar(&srcDbDsn, "src-dsn", "", "源库连接串")
	fs.StringVar(&dstDbDsn, "dst-dsn", "", "目标库连接串")
	fs.StringVar(&srcSqlFile, "src-sql-file", "", "源库SQL文件")
	fs.StringVar(&dstSqlFile, "dst-sql-file", "", "目标库SQL文件")
	fs.StringVar(&MysqlVersion, "mysql-version", "", "mysql版本,仅使用文件比对时需要指定,默认值为:"+vtconfig.DefaultMySQLVersion)
	fs.StringVar(&skipTables, "skip-tables", "", "跳过特定表比对,多个表使用逗号分隔")

	fs.Parse(os.Args[1:])

	var err error

	config, err = conf.NewConfFromFile(confFileName)
	if err != nil {
		log.Println("获取配置文件失败")
		os.Exit(1)
	}
	// 优先使用命令行参数
	if srcDbDsn != "" {
		config.SrcSchemaConf.Dsn = srcDbDsn
	}
	if dstDbDsn != "" {
		config.DstSchemaConf.Dsn = dstDbDsn
	}
	if SaveSqlFileName != "" {
		config.SaveSqlPath = SaveSqlFileName
	}
	if srcSqlFile != "" {
		config.SrcSchemaConf.SqlFile = srcSqlFile
	}
	if dstSqlFile != "" {
		config.DstSchemaConf.SqlFile = dstSqlFile
	}

	if err := config.InitAndCheck(MysqlVersion, skipTables); err != nil {
		log.Printf("初始化配置失败: %+v\n", err)
		os.Exit(1)
	}
}

func main() {
	// 比对数据库版本
	if config.SrcSchemaConf.Source == conf.SOURCEISSERVER && config.DstSchemaConf.Source == conf.SOURCEISSERVER {
		srcVersion := config.SrcSchemaConf.MysqlVersion
		dstVersion := config.DstSchemaConf.MysqlVersion
		if srcVersion != dstVersion {
			log.Printf("注意: 源MySQL库版本为:[%s],目标MySQL库版本为:[%s]\n", srcVersion, dstVersion)
		}
	}
	// 获取源库schema
	srcSchemas, err := diffsql.GetSchemas(config.SrcSchemaConf)
	if err != nil {
		log.Printf("获取源库schema失败: %+v\n", err)
		os.Exit(1)
	}
	// 获取目标库schema
	dstSchemas, err := diffsql.GetSchemas(config.DstSchemaConf)
	if err != nil {
		log.Printf("获取目标库schema失败: %+v\n", err)
		os.Exit(1)
	}

	for _, tableName := range config.SkipTables {
		log.Printf("表[%s]被跳过\n", tableName)
		delete(srcSchemas, tableName)
	}

	diffUseMysqlVersion := config.GetDiffUseVersion()
	alterTableSql, err := diffsql.DiffSchemas(diffUseMysqlVersion, srcSchemas, dstSchemas)
	if err != nil {
		log.Printf("比对失败: %+v\n", err)
		os.Exit(1)
	}

	if len(alterTableSql) == 0 {
		log.Println("比对完成, 两个库的schema完全一致")
		return
	}

	diffSqlStr := diffsql.GenerateDiffSQL(alterTableSql)

	if config.SaveSqlPath == "" {
		fmt.Println("================以下为需要要在目标库执行的SQL================")
		fmt.Println(diffSqlStr)
		return
	}

	f, err := os.OpenFile(config.SaveSqlPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		log.Printf("打开diff结果SQL文件失败[%s]: %+v", config.SaveSqlPath, err)
		return
	}
	defer f.Close()

	_, err = f.WriteString(diffSqlStr)
	if err != nil {
		log.Printf("diff结果SQL文件写入失败失败[%s]: %+v", config.SaveSqlPath, err)
	} else {
		log.Printf("diff结果SQL文件生成成功[%s]", config.SaveSqlPath)
	}
}
