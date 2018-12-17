package db

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"os"
)

var DbInstance *gorm.DB

func Connect() {
	var err error
	//DbInstance, err = gorm.Open("mysql", "root:root@tcp(127.0.0.1:3306)/test?charset=utf8&parseTime=True&loc=Local")
	DbInstance, err = gorm.Open("mysql", "gw:gao123456@tcp(192.168.30.139:3306)/laravelschool?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		fmt.Println("failed to connect database" + err.Error())
		os.Exit(501)
	}
}

func Close() {
	DbInstance.Close()
}

func init() {
	//Connect()
	//caijiModel := &models.Caijie{}
	//DbInstance.AutoMigrate(caijiModel)
	//DbInstance.CreateTable(&models.Caijie{})
}

//func main() {
//	Connect()
//	caijiModel := &models.Caijie{}
//	DbInstance.AutoMigrate(caijiModel)
//
//	//caijiModel.Content = ""
//	//// Create
//	//db.Create(&Product{Code: "L1212", Price: 1000})
//	//
//	//// Read
//	//var product Product
//	//db.First(&product, 1) // find product with id 1
//	//db.First(&product, "code = ?", "L1212") // find product with code l1212
//	//
//	//// Update - update product's price to 2000
//	//db.Model(&product).Update("Price", 2000)
//	//
//	//// Delete - delete product
//	//db.Delete(&product)
//}
