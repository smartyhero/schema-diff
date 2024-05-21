package conf

import (
	"os"

	"gopkg.in/yaml.v2"
)

type DbConf struct {
	Dsn string `yaml:"dsn"`
}

type DiffConf struct {
	IgnoreCharacter     bool `yaml:"ignore_character"`
	IgnoreAutoIncrement bool `yaml:"ignore_auto_increment"`
}

type Conf struct {
	SrcDbConf   *DbConf   `yaml:"src_db"`
	DstDbConf   *DbConf   `yaml:"dst_db"`
	DiffConf    *DiffConf `yaml:"diff_conf"`
	SaveSqlPath string    `yaml:"save_sql_path"`
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
