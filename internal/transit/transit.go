package transit

import "time"

type Stop struct {
	StopID                 string    `gorm:"column:stop_id;primaryKey"`
	Agency                 string    `gorm:"column:agency"`
	Name                   string    `gorm:"column:name"`
	Latitude               float64   `gorm:"column:latitude"`
	Longitude              float64   `gorm:"column:longitude"`
	CommuteSecondsToMcGill *int      `gorm:"column:commute_seconds_to_mcgill"`
	SnapshotKey            string    `gorm:"column:snapshot_key"`
	ComputedAt             time.Time `gorm:"column:computed_at"`
}

func (Stop) TableName() string { return "transit_stops" }
