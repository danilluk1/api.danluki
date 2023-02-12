package entities

import "time"

type Statistics struct {
	ID          string    `gorm:"primary_key;column:id;type:uuid;default:gen_random_uuid();" json:"id"`
	NumRange    int       `gorm:"column:numRange;type:INT4;" json:"numRange"`
	Path        string    `gorm:"column:path;type:string;" json:"path"`
	RequestedAt time.Time `gorm:"column:requestedAt;type:timestamp;default:CURRENT_TIMESTAMP();" json:"requestedAt"`
}
