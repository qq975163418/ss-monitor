package model

import (
	"time"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pkg/errors"
)

type Heartbeat struct {
	ID        uint   `gorm:"AUTO_INCREMENT"`
	Class     string `gorm:"not null;index"`
	IPVer     uint   `gorm:"index"`
	Name      string `gorm:"index"`
	Time      int64  `gorm:"index"`
	CreatedAt time.Time
}

func GetHeartbeats(db *gorm.DB, time int64, class string, ipVer uint) (heartbeats []Heartbeat, err error) {
	switch class {
	case "tester":
		err = db.Where("time > ?", time).Where("class like ?", class).Where("ip_ver like ?", ipVer).
			Where("ip_ver like ?", 10).Order("name asc").Find(&heartbeats).Error
	default:
		err = db.Where("time > ?", time).Where("class like ?", class).
			Order("name asc").Find(&heartbeats).Error
	}
	if err != nil {
		err = errors.Wrap(err, "GetHeartbeats")
		return
	}
	return
}

func SaveHeartbeat(db *gorm.DB, class string, ipVer uint, name string, t int64) (newHeartbeat Heartbeat, err error) {
	var count int
	var heartbeat Heartbeat
	err = db.Model(&Heartbeat{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		err = errors.Wrap(err, "SaveHeartbeat: CountHeartbeat")
		return
	}
	if count == 0 {
		heartbeat.Class = class
		heartbeat.IPVer = ipVer
		heartbeat.Time = t
		heartbeat.Name = name
		err = db.Create(&heartbeat).Error
		if err != nil {
			err = errors.Wrap(err, "SaveHeartbeat: CreateHeartbeat")
			return
		}
		newHeartbeat = heartbeat
		return
	} else {
		err = db.Where("name = ?", name).First(&heartbeat).Error
		if err != nil {
			err = errors.Wrap(err, "SaveHeartbeat: QueryHeartbeat")
			return
		}
		heartbeat.Class = class
		heartbeat.IPVer = ipVer
		heartbeat.Time = t
		err = db.Model(&heartbeat).Updates(heartbeat).Error
		if err != nil {
			err = errors.Wrap(err, "SaveHeartbeat: UpdateHeartbeat")
			return
		}
	}
	return
}
