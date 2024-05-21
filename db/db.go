package db

import (
	"fmt"

	"schema-diff/conf"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DbConf struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DbName   string `yaml:"db_name"`
}

type SchemaRes struct {
	Table      string `gorm:"column:Table"`
	CreateStmt string `gorm:"column:Create Table"`
}

// type Db struct {
// 	*gorm.DB
// }

type Db struct {
	*gorm.DB
}

func NewDB(conf *conf.DbConf) (*Db, error) {
	goDB, err := gorm.Open(mysql.Open(conf.Dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}

	db, err := goDB.DB()
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(64)
	return &Db{goDB}, nil
}

func (db *Db) GetAllTables() ([]string, error) {
	tableName := []string{}
	tx := db.Raw("show tables").Scan(&tableName)
	if tx.Error != nil {
		return tableName, tx.Error
	}
	return tableName, nil
}

func (db *Db) GetTableSchema(tableName string) (string, error) {
	migrator := db.Migrator()

	if !migrator.HasTable(tableName) {
		return "", ErrNoSuchTable
	}

	Schema := &SchemaRes{}
	tx := db.Raw(fmt.Sprintf("show create table `%s`", tableName)).Scan(Schema)
	if tx.Error != nil {
		return "", tx.Error
	}
	return Schema.CreateStmt, nil
}

func (db *Db) GetMysqlVersion() (string, error) {
	var version string
	tx := db.Raw("SELECT VERSION()").Scan(&version)
	if tx.Error != nil {
		return "", tx.Error
	}
	return version, nil
}
