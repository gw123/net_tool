package models

import "time"

type Caijie struct {
	ID        int    `gorm:"primary_key"`
	Type      string `gorm:"type:varchar(20);"`
	Cate      string `gorm:"type:varchar(20);"`
	Title     string `gorm:"type:varchar(128);not null;index:title_idx"`
	Content   string `gorm:"type:text;"`
	FromSite  string `gorm:"type:varchar(256);"`
	FromUrl   string `gorm:"type:varchar(256);"`
	Author    string `gorm:"type:varchar(56);"`
	Info      string `gorm:"type:text;"`
	UsedTimes int
	CreatedAt time.Time
}

