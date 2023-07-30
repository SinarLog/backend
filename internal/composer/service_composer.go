package composer

import (
	"github.com/go-co-op/gocron"
	impl "sinarlog.com/internal/adapter/service"
	"sinarlog.com/internal/app/service"
	"sinarlog.com/pkg/bucket"
	"sinarlog.com/pkg/doorkeeper"
	"sinarlog.com/pkg/mailer"
	"sinarlog.com/pkg/postgres"
	"sinarlog.com/pkg/rater"
	"sinarlog.com/pkg/redis"
)

type IServiceComposer interface {
	DoorkeeperService() service.IDoorkeeperService
	RaterService() service.IRaterService
	MailerService() service.IMailerService
	BucketService() service.IBucketService
	SchedulerService() service.ISchedulerService
	NotifService() service.INotifService
}

type serviceComposer struct {
	dk   *doorkeeper.Doorkeeper
	rt   *rater.Rater
	ml   *mailer.Mailer
	bkt  *bucket.Bucket
	rdis *redis.RedisClient
	sch  *gocron.Scheduler
}

func NewServiceComposer(dk *doorkeeper.Doorkeeper, rt *rater.Rater, ml *mailer.Mailer, bkt *bucket.Bucket, rdis *redis.RedisClient, sch *gocron.Scheduler) IServiceComposer {
	s := &serviceComposer{dk: dk, rt: rt, ml: ml, bkt: bkt, rdis: rdis, sch: sch}

	// Init
	// s.SchedulerService()

	return s
}

func (s *serviceComposer) DoorkeeperService() service.IDoorkeeperService {
	return impl.NewDoorkeeperService(s.dk)
}

func (s *serviceComposer) RaterService() service.IRaterService {
	return impl.NewRaterService(s.rt)
}

func (s *serviceComposer) MailerService() service.IMailerService {
	return impl.NewMailerService(s.ml)
}

func (s *serviceComposer) BucketService() service.IBucketService {
	return impl.NewBucketService(s.bkt)
}

func (s *serviceComposer) NotifService() service.INotifService {
	return impl.NewNotifService(s.rdis.Client)
}

func (s *serviceComposer) SchedulerService() service.ISchedulerService {
	db := postgres.GetPostgres("").ORM
	rdis := redis.NewRedisClient().Client
	return impl.NewSchedulerService(s.sch, db, rdis)
}
