package demo

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/gw123/net_tool/rouding/db/models"
	"github.com/gw123/net_tool/rouding/db"
)

func main() {
	db.Connect()
	caijiModel := &models.Caijie{}
	db.DbInstance.AutoMigrate(caijiModel)

	//caijiModel.Content = ""
	//// Create
	//db.Create(&Product{Code: "L1212", Price: 1000})
	//
	//// Read
	//var product Product
	//db.First(&product, 1) // find product with id 1
	//db.First(&product, "code = ?", "L1212") // find product with code l1212
	//
	//// Update - update product's price to 2000
	//db.Model(&product).Update("Price", 2000)
	//
	//// Delete - delete product
	//db.Delete(&product)
}
