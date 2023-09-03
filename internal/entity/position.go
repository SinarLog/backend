package entity

// Example of a position is Software Developer, UI/UX, Product Manager
type Job struct {
	BaseModelID

	Name string `gorm:"type:varchar(100)"`

	Employees []Employee

	BaseModelStamps
	BaseModelSoftDelete
}
