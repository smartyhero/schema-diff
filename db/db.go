package db

import (
	"database/sql"
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

type Db struct {
	User     string
	Password string
	Host     string
	Port     string
	DbName   string
	gormDb   *gorm.DB
}

func NewDB(conf *conf.DbConf) (*Db, error) {
	Db := &Db{
		User:     conf.User,
		Password: conf.Password,
		Host:     conf.Host,
		Port:     conf.Port,
		DbName:   conf.DbName,
	}
	mysqlDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True", conf.User, conf.Password, conf.Host,
		conf.Port, conf.DbName)

	sqlDB, err := sql.Open("mysql", mysqlDsn)
	sqlDB.SetMaxOpenConns(64)
	if err != nil {
		return nil, err
	}
	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}

	goDB, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}
	Db.gormDb = goDB
	return Db, nil
}

func (db *Db) GetAllTables() ([]string, error) {
	tableName := []string{}
	tx := db.gormDb.Raw("show tables").Scan(&tableName)
	if tx.Error != nil {
		return tableName, tx.Error
	}
	return tableName, nil
}

func (db *Db) GetTableSchema(tableName string) (string, error) {
	migrator := db.gormDb.Migrator()

	if !migrator.HasTable(tableName) {
		return "", ErrNoSuchTable
	}

	Schema := &SchemaRes{}
	tx := db.gormDb.Raw(fmt.Sprintf("show create table `%s`", tableName)).Scan(Schema)
	if tx.Error != nil {
		return "", tx.Error
	}
	return Schema.CreateStmt, nil
}

func (db *Db) GetMysqlVersion() (string, error) {
	var version string
	tx := db.gormDb.Raw("SELECT VERSION()").Scan(&version)
	if tx.Error != nil {
		return "", tx.Error
	}
	return version, nil
}
