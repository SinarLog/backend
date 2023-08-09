package composer

import (
	"context"
	"log"

	impl "sinarlog.com/internal/adapter/repo"
	"sinarlog.com/internal/app/repo"
	"sinarlog.com/pkg/mongo"
	"sinarlog.com/pkg/postgres"
	"sinarlog.com/pkg/redis"
)

type IRepoComposer interface {
	CredentialRepo() repo.ICredentialRepo
	SharedRepo() repo.ISharedRepo
	ConfigRepo() repo.IConfigRepo
	JobRepo() repo.IJobRepo
	RoleRepo() repo.IRoleRepo
	EmployeeRepo() repo.IEmployeeRepo
	AttendanceRepo() repo.IAttendanceRepo
	LeaveRepo() repo.ILeaveRepo
	AnalyticsRepo() repo.IAnalyticsRepo
	ChatRepo() repo.IChatRepo

	Migrate()
}

type repoComposer struct {
	db    *postgres.Postgres
	redis *redis.RedisClient
	mongo *mongo.Mongo
	env   string
}

func NewRepoComposer(db *postgres.Postgres, rdis *redis.RedisClient, mg *mongo.Mongo, env string) IRepoComposer {
	comp := new(repoComposer)
	comp.env = env
	comp.db = db
	comp.redis = rdis
	comp.mongo = mg

	comp.Migrate()
	comp.Seed()
	if comp.env == "DEVELOPMENT" {
		comp.setToDebug()
	}

	return comp
}

// -------------- DI --------------
func (c *repoComposer) CredentialRepo() repo.ICredentialRepo {
	return impl.NewCredentialRepo(c.db.ORM, c.redis.Client)
}

func (c *repoComposer) SharedRepo() repo.ISharedRepo {
	return impl.NewSharedRepo(c.db.ORM)
}

func (c *repoComposer) JobRepo() repo.IJobRepo {
	return impl.NewJobRepo(c.db.ORM)
}

func (c *repoComposer) RoleRepo() repo.IRoleRepo {
	return impl.NewRoleRepo(c.db.ORM)
}

func (c *repoComposer) ConfigRepo() repo.IConfigRepo {
	return impl.NewConfigRepo(c.db.ORM, c.redis.Client)
}

func (c *repoComposer) EmployeeRepo() repo.IEmployeeRepo {
	return impl.NewEmployeeRepo(c.db.ORM)
}

func (c *repoComposer) AttendanceRepo() repo.IAttendanceRepo {
	return impl.NewAttendanceRepo(c.db.ORM, c.redis.Client)
}

func (c *repoComposer) LeaveRepo() repo.ILeaveRepo {
	return impl.NewLeaveRepo(c.db.ORM)
}

func (c *repoComposer) AnalyticsRepo() repo.IAnalyticsRepo {
	return impl.NewAnalyticsRepo(c.db.ORM)
}

func (c *repoComposer) ChatRepo() repo.IChatRepo {
	return impl.NewChatRepo(c.mongo.Conn)
}

// -------------- Setups --------------
func (c *repoComposer) setToDebug() {
	c.db.ORM = c.db.ORM.Debug()
}

func (c *repoComposer) Migrate() {
	c.db.ORM.AutoMigrate(
		impl.GetAllRelationalEntities()...,
	)
}

func (c *repoComposer) Seed() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	seeder := impl.NewSeeder(c.db.ORM)
	err := seeder.Seed(ctx)
	if err != nil {
		log.Fatalf("\n\n\tError during seeding.\n\tDo check your database whether seeding has worked as expected.\n\tThe error(s) is(are):\n%s\n\n", err)
	}
}
