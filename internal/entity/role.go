package entity

// Example of a role is staff, manager, and HR
type Role struct {
	BaseModelID

	Name string `gorm:"uniqueIndex;type:varchar(100)"`
	Code string `gorm:"type:varchar(50)"`

	Employees []Employee

	BaseModelStamps
	BaseModelSoftDelete
}
