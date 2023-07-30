package repo

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/utils"
)

type (
	seederFunc func(context.Context) error

	seeder struct {
		db *gorm.DB
	}
)

func NewSeeder(db *gorm.DB) *seeder {
	return &seeder{db}
}

// Seed executes the seeding process.
func (repo *seeder) Seed(ctx context.Context) error {
	var errs error

	// Register your seeder functions here
	seeders := []seederFunc{
		repo.seedConfig,
		repo.seedRoles,
		repo.seedJobs,
		repo.seedHRs,
		repo.seedDummyManager,
	}

	for _, f := range seeders {
		if err := f(ctx); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// Seed config for V2 done.
func (repo *seeder) seedConfig(ctx context.Context) error {
	config := entity.Configuration{
		OfficeStartTime:              time.Date(2000, time.January, 1, 8, 0, 0, 0, utils.CURRENT_LOC),
		OfficeEndTime:                time.Date(2000, time.January, 1, 17, 0, 0, 0, utils.CURRENT_LOC),
		AcceptanceAttendanceInterval: "30m",
		AcceptanceLeaveInterval:      7,
		DefaultYearlyQuota:           12,
		DefaultMarriageQuota:         3,
		MaxOvertimeDailyDur:          3,
		MaxOvertimeWeeklyDur:         14,
	}

	// If no config record.
	if err := repo.db.WithContext(ctx).First(&config).Error; err != nil {
		return repo.db.WithContext(ctx).FirstOrCreate(&config).Error
	}

	return nil
}

// Seed roles for V2 done.
func (repo *seeder) seedRoles(ctx context.Context) error {
	roles := []entity.Role{
		{
			Name: "Staff",
			Code: "staff",
		},
		{
			Name: "Manager",
			Code: "mngr",
		},
		{
			Name: "HR",
			Code: "hr",
		},
	}

	for _, r := range roles {
		err := repo.db.WithContext(ctx).Model(&entity.Role{}).FirstOrCreate(&r, r).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// Seed jobs for V2 done.
func (repo *seeder) seedJobs(ctx context.Context) error {
	jobs := []entity.Job{
		{
			Name: "Software Developer",
		},
		{
			Name: "Product Manager",
		},
		{
			Name: "UI/UX Designer",
		},
		{
			Name: "Business Analyst",
		},
		{
			Name: "Data Analyst",
		},
		{
			Name: "Data Scientist",
		},
		{
			Name: "Human Resource",
		},
	}

	for _, j := range jobs {
		err := repo.db.WithContext(ctx).Model(&entity.Job{}).FirstOrCreate(&j, j).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// Seed hr for V2 done.
func (repo *seeder) seedHRs(ctx context.Context) error {
	var role entity.Role
	var job entity.Job

	// Get the role entity
	if err := repo.db.WithContext(ctx).Model(&role).First(&role, "code = ?", "hr").Error; err != nil {
		return err
	}

	// Get the job entity
	if err := repo.db.WithContext(ctx).Model(&job).First(&job, "name = ?", "Human Resource").Error; err != nil {
		return err
	}

	pass, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	hr := entity.Employee{
		FullName:     "Angela Kumawa",
		Email:        "angelakumawa@sinarlog.com",
		Password:     string(pass),
		ContractType: entity.FULL_TIME,
		JoinDate:     time.Date(2022, 1, 1, 0, 0, 0, 0, utils.CURRENT_LOC),
		Avatar:       "",
		Status:       entity.AVAILABLE,
		EmployeeBiodata: entity.EmployeeBiodata{
			NIK:           "3269020202020202",
			NPWP:          "99.999.999-9.999.999",
			Gender:        entity.M,
			Religion:      entity.BUDDHA,
			PhoneNumber:   "+62-812-3427-4916",
			Address:       "Jln. Tomang Grogol Cibubur Kemang Pondok Indah Blok M",
			BirthDate:     time.Date(2000, 6, 6, 0, 0, 0, 0, utils.CURRENT_LOC),
			MaritalStatus: false,
		},
		EmployeesEmergencyContacts: []entity.EmployeesEmergencyContact{
			{
				FullName:    "Angela Kumawa Senior",
				Relation:    entity.MOTHER,
				PhoneNumber: "+62-812-3456-7899",
			},
		},
		EmployeeLeavesQuota: entity.EmployeeLeavesQuota{
			YearlyCount:   12,
			UnpaidCount:   0,
			MarriageCount: 3,
		},
		RoleID: role.Id,
		Role:   role,
		JobID:  job.Id,
		Job:    job,
	}

	return repo.db.WithContext(ctx).FirstOrCreate(&hr, entity.Employee{Email: hr.Email}).Error
}

// Seed for manager V2 done.
func (repo *seeder) seedDummyManager(ctx context.Context) error {
	var role entity.Role
	var job entity.Job

	// Get the role entity
	if err := repo.db.WithContext(ctx).Model(&role).First(&role, "code = ?", "mngr").Error; err != nil {
		return err
	}

	// Get the job entity
	if err := repo.db.WithContext(ctx).Model(&job).First(&job, "name = ?", "Software Developer").Error; err != nil {
		return err
	}

	pass, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Get angela kumawa hr
	var hr entity.Employee
	if err := repo.db.WithContext(ctx).Model(&hr).First(&hr, "email = ?", "angelakumawa@sinarlog.com").Error; err != nil {
		return err
	}

	manager := entity.Employee{
		FullName:     "Johny Deep",
		Email:        "johnydeep@sinarlog.com",
		Password:     string(pass),
		ContractType: entity.FULL_TIME,
		JoinDate:     time.Date(2022, 2, 2, 0, 0, 0, 0, utils.CURRENT_LOC),
		Avatar:       "",
		Status:       entity.UNAVAILABLE,
		CreatedById:  &hr.Id,
		EmployeeBiodata: entity.EmployeeBiodata{
			NIK:           "3269020202020203",
			NPWP:          "99.999.999-9.999.991",
			Gender:        entity.M,
			Religion:      entity.CONFUCION,
			PhoneNumber:   "+62-812-3427-4917",
			Address:       "Jln. Tomang Grogol Cibubur Kemang Pondok Indah Blok M",
			BirthDate:     time.Date(2000, 6, 6, 0, 0, 0, 0, utils.CURRENT_LOC),
			MaritalStatus: false,
		},
		EmployeesEmergencyContacts: []entity.EmployeesEmergencyContact{
			{
				FullName:    "Johny Deep Mother",
				Relation:    entity.MOTHER,
				PhoneNumber: "+62-812-3456-7897",
			},
		},
		EmployeeLeavesQuota: entity.EmployeeLeavesQuota{
			YearlyCount:   12,
			UnpaidCount:   0,
			MarriageCount: 3,
		},
		RoleID: role.Id,
		Role:   role,
		JobID:  job.Id,
		Job:    job,
	}

	return repo.db.WithContext(ctx).FirstOrCreate(&manager, entity.Employee{Email: manager.Email}).Error
}
