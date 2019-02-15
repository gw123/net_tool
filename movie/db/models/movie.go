package models

type Movie struct {
	ID              int    `gorm:"primary_key"`
	Title           string `gorm:"type:varchar(128);not null;index:title_idx"`
	Note            string `gorm:"type:varchar(128);comment:'影片备注'"`
	Actor           string `gorm:"type:varchar(256);"`
	Direction       string `gorm:"type:varchar(256);"`
	Tppe            string `gorm:"type:varchar(32);"`
	FromUrl         string `gorm:"type:varchar(256);"`
	Area            string `gorm:"type:varchar(56);"`
	LastUpdatedTime string `gorm:"type:varchar(20);"`
	PublishedTime   string `gorm:"type:varchar(20);"`
	Language        string `gorm:"type:varchar(20);"`
	Status          string `gorm:"type:varchar(20);"`
	Url1            string `gorm:"type:varchar(256);"`
	Url2            string `gorm:"type:varchar(256);"`
	Desc            string `gorm:"type:varchar(1024);"`
	Img             string `gorm:"type:varchar(256);"`
}
