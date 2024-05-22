package diffsql

import (
	"fmt"
	"io"
	"log"
	"os"

	"schema-diff/conf"
	"schema-diff/db"

	"vitess.io/vitess/go/mysql/collations"
	"vitess.io/vitess/go/vt/schemadiff"
	"vitess.io/vitess/go/vt/sqlparser"
	"vitess.io/vitess/go/vt/vtenv"
)

func GetSchemas(schemaConf *conf.SchemaConf) (schemas map[string]string, err error) {
	if schemaConf.Source == conf.SOURCEISSERVER {
		mydb, err := db.NewDB(schemaConf.Dsn)
		if err != nil {
			return nil, err
		}
		return mydb.GetAllTableSchema()
	}
	if schemaConf.Source == conf.SOURCEISFILE {
		return GetTableSchemasFromFile(schemaConf.SqlFile, schemaConf.MysqlVersion)
	}
	return nil, conf.ErrUnknownSchemaSource
}

func DiffSchemas(MysqlVersion string, IgnoreCharset bool, srcSchemas map[string]string, dstSchemas map[string]string) (diffs []string, err error) {
	if len(srcSchemas) == 0 && len(dstSchemas) == 0 {
		log.Println("源库和目标库Schema都为空，无法进行对比")
		return diffs, db.ErrSrcDstSchemaIsNull
	}

	srcOpt := vtenv.Options{
		MySQLServerVersion: MysqlVersion,
	}

	srcEnv, err := vtenv.New(srcOpt)
	if err != nil {
		return nil, err
	}

	diffHints := &schemadiff.DiffHints{
		StrictIndexOrdering:         false,
		AutoIncrementStrategy:       schemadiff.AutoIncrementIgnore,
		TableCharsetCollateStrategy: schemadiff.TableCharsetCollateIgnoreAlways,
	}
	if IgnoreCharset {
		diffHints.TableCharsetCollateStrategy = schemadiff.TableCharsetCollateIgnoreAlways
	}

	collationID := collations.NewEnvironment(MysqlVersion).DefaultConnectionCharset()
	schemadiffEnv := schemadiff.NewEnv(srcEnv, collationID)

	for tableName, schema := range srcSchemas {
		dstSchema, ok := dstSchemas[tableName]
		if !ok {
			log.Printf("目标库中不存在[%s]表\n", tableName)
		}
		diff, err := schemadiff.DiffCreateTablesQueries(schemadiffEnv, dstSchema, schema, diffHints)
		if err != nil {
			log.Printf("对比表结构失败[%s]\n", tableName)
			return diffs, err
		}
		if diff.IsEmpty() {
			log.Printf("源库和目标库表结构一致[%s]\n", tableName)
		} else {
			log.Printf("源库和目标库表结构不一致[%s]\n", tableName)
			diffs = append(diffs, diff.StatementString())
		}
	}
	return diffs, nil
}

func GetTableSchemasFromFile(fileName string, mysqlVersion string) (map[string]string, error) {
	allSchemas := make(map[string]string)
	sqlData, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	convVersion, err := sqlparser.ConvertMySQLVersionToCommentVersion(mysqlVersion)
	if err != nil {
		return nil, err
	}

	parser, err := sqlparser.New(sqlparser.Options{
		MySQLServerVersion: convVersion,
		TruncateUILen:      512,
		TruncateErrLen:     0,
	})
	if err != nil {
		return nil, err
	}
	tokens := parser.NewStringTokenizer(string(sqlData))
	for {
		stmt, err := sqlparser.ParseNext(tokens)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("解析sql出错: %+v\n", err)
			continue
		}
		switch stmt := stmt.(type) {
		case *sqlparser.CreateTable:
			tableName := stmt.Table.Name.String()
			allSchemas[tableName] = sqlparser.String(stmt)
		default:
			log.Printf("%s中存在非建表sql: %s\n", fileName, sqlparser.String(stmt))
		}
	}
	return allSchemas, nil
}

func GenerateDiffSQL(alterTableSql []string) string {
	diffSqlStr := ""
	for _, sql := range alterTableSql {
		if diffSqlStr == "" {
			diffSqlStr = fmt.Sprintf("%s;\n", sql)
		} else {
			diffSqlStr = fmt.Sprintf("%s\n%s;\n", diffSqlStr, sql)
		}
	}
	return diffSqlStr
}
