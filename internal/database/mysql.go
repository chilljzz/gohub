package database

import (
	"fmt"

	"github.com/chilljzz/gohub/internal/model"
	"github.com/chilljzz/gohub/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitMySQL() error {
	c := config.Conf.Mysql

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.Charset,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	DB = db
	fmt.Println("MySQL connected success")
	if err := DB.AutoMigrate(&model.User{}); err != nil {
		return err
	}
	return nil
}
