package database

import (
	"fmt"

	"github.com/chilljzz/gohub/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func OpenMySQL(c config.MySQLConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
}
