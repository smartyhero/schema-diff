package conf

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type DbConf struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DbName   string `yaml:"db_name"`
}

type DiffConf struct {
	IgnoreCharacter     bool `yaml:"ignore_character"`
	IgnoreAutoIncrement bool `yaml:"ignore_auto_increment"`
}

type Conf struct {
	SrcDbConf *DbConf   `yaml:"src_db"`
	DstDbConf *DbConf   `yaml:"dst_db"`
	DiffConf  *DiffConf `yaml:"diff_conf"`
}

func NewConfFromFile(fileName string) (*Conf, error) {
	f, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return NewConf(f)
}

func NewConf(confStr []byte) (*Conf, error) {
	conf := &Conf{}
	err := yaml.Unmarshal(confStr, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func ReadFileToStr(file string) (string, error) {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(f), nil
}
