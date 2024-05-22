package conf

import (
	"os"

	"schema-diff/db"

	vtconfig "vitess.io/vitess/go/mysql/config"

	"gopkg.in/yaml.v2"
)

const (
	SOURCEISFILE   = "file"
	SOURCEISSERVER = "server"
)

type DbConf struct {
	Dsn          string `yaml:"dsn"`
	SqlFile      string `yaml:"sql_file"`
	MysqlVersion string `yaml:"mysql_version"`
}

type SchemaConf struct {
	Source       string `yaml:"source"` // file|server
	Dsn          string `yaml:"dsn"`
	SqlFile      string `yaml:"sql_file"`
	MysqlVersion string `yaml:"mysql_version"`
}

type Conf struct {
	SrcSchemaConf *SchemaConf `yaml:"src_schema"`
	DstSchemaConf *SchemaConf `yaml:"dst_schema"`
	SaveSqlPath   string      `yaml:"save_sql_path"`
}

func NewConfFromFile(fileName string) (*Conf, error) {
	f, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return NewConf([]byte(os.ExpandEnv(string(f))))
}

func NewConf(confStr []byte) (*Conf, error) {
	conf := &Conf{}
	err := yaml.Unmarshal(confStr, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func (conf *Conf) InitAndCheck(cmdVarVersion string) error {
	if conf.SrcSchemaConf.Dsn == "" && conf.SrcSchemaConf.SqlFile == "" {
		return ErrConfigSrcMiss
	}
	if conf.DstSchemaConf.Dsn == "" && conf.DstSchemaConf.SqlFile == "" {
		return ErrConfigDstMiss
	}

	if conf.SrcSchemaConf.SqlFile != "" {
		if _, err := os.Stat(conf.SrcSchemaConf.SqlFile); err != nil {
			return err
		}
		conf.SrcSchemaConf.Source = SOURCEISFILE
	} else {
		conf.SrcSchemaConf.Source = SOURCEISSERVER
	}

	if conf.DstSchemaConf.SqlFile != "" {
		if _, err := os.Stat(conf.DstSchemaConf.SqlFile); err != nil {
			return err
		}
		conf.DstSchemaConf.Source = SOURCEISFILE
	} else {
		conf.DstSchemaConf.Source = SOURCEISSERVER
	}

	if err := conf.SrcSchemaConf.initMysqlVersion(cmdVarVersion); err != nil {
		return err
	}
	if err := conf.DstSchemaConf.initMysqlVersion(cmdVarVersion); err != nil {
		return err
	}

	return nil
}

func (sc *SchemaConf) initMysqlVersion(cmdVarVersion string) error {
	if sc.Source == SOURCEISSERVER {
		mydb, err := db.NewDB(sc.Dsn)
		if err != nil {
			return err
		}
		version, err := mydb.GetMysqlVersion()
		if err != nil {
			return err
		}
		sc.MysqlVersion = version
		return nil
	}
	if sc.Source == SOURCEISFILE {
		if cmdVarVersion != "" {
			sc.MysqlVersion = cmdVarVersion
			return nil
		}
		if sc.MysqlVersion == "" {
			sc.MysqlVersion = vtconfig.DefaultMySQLVersion
			return nil
		} else {
			return nil
		}
	}
	return ErrUnknownSchemaSource
}

func (conf *Conf) GetDiffUseVersion() string {
	if conf.SrcSchemaConf.Source == SOURCEISSERVER {
		return conf.SrcSchemaConf.MysqlVersion
	}
	if conf.DstSchemaConf.Source == SOURCEISSERVER {
		return conf.DstSchemaConf.MysqlVersion
	}
	return conf.SrcSchemaConf.MysqlVersion
}
