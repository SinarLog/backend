package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"sinarlog.com/config"
	"sinarlog.com/internal/composer"
	v2 "sinarlog.com/internal/delivery/v2"
	"sinarlog.com/pkg/bucket"
	"sinarlog.com/pkg/doorkeeper"
	httpserver "sinarlog.com/pkg/http"
	"sinarlog.com/pkg/logger"
	"sinarlog.com/pkg/mailer"
	"sinarlog.com/pkg/mongo"
	"sinarlog.com/pkg/postgres"
	"sinarlog.com/pkg/pubsub"
	"sinarlog.com/pkg/rater"
	"sinarlog.com/pkg/redis"

	_ "sinarlog.com/internal/utils"
)

func Run(cfg *config.Config) {
	// Global context
	app_context, cancel := context.WithCancel(context.Background())

	// Postgres
	pg := postgres.GetPostgres(
		cfg.Db.URL,
		postgres.MaxPoolSize(cfg.Db.MaxPoolSize),
		postgres.MaxOpenCoon(cfg.Db.MaxOpenConn),
		postgres.MaxConnLifetime(cfg.Db.MaxConnLifetime),
	)

	// Redis
	rdis := redis.NewRedisClient(
		redis.RegisterAddress(cfg.Redis.Address),
		redis.RegisterPassword(cfg.Redis.Password),
		redis.RegisterDB(cfg.Redis.Db),
		redis.RegisterReadTimeout(cfg.Redis.ReadTimeout),
		redis.RegisterWriteTimeout(cfg.Redis.WriteTimeout),
		redis.RegisterMinIdleConn(cfg.Redis.MinIdleConn),
		redis.RegisterMaxIdleConn(cfg.Redis.MaxIdleConn),
		redis.RegisterMaxIdleTime(cfg.Redis.MaxIdleTime),
	)

	// Mongo
	mg := mongo.GetMongoClient(
		app_context,
		mongo.RegisterURI(cfg.Mongo.URI),
		mongo.RegisterDbName(cfg.Mongo.DbName),
		mongo.MaxConn(cfg.Mongo.MaxOpenConn),
		mongo.MaxPoolSize(cfg.Mongo.MaxPoolSize),
		mongo.MaxConnLifetime(cfg.Mongo.MaxConnLifetime),
	)

	// Logger
	logger := logger.NewAppLogger(cfg.App.LogPath)

	// Doorkeeper
	dk := doorkeeper.GetDoorkeeper(
		doorkeeper.RegisterHasherFunc(cfg.Doorkeeper.HashMethod),
		doorkeeper.RegisterSignMethod(cfg.Doorkeeper.SigningMethod, cfg.Doorkeeper.SignSize),
		doorkeeper.RegisterIssuer(cfg.Doorkeeper.Issuer),
		doorkeeper.RegisterAccessDuration(cfg.Doorkeeper.AccessDuration),
		doorkeeper.RegisterRefreshDuration(cfg.Doorkeeper.RefreshDuration),
		doorkeeper.RegisterPrivatePath(cfg.Doorkeeper.PrivPath),
		doorkeeper.RegisterPublicPath(cfg.Doorkeeper.PubPath),
		doorkeeper.RegisterOTPSecretLength(cfg.Doorkeeper.OTPSecretLength),
		doorkeeper.RegisterOTPExpDuration(cfg.Doorkeeper.OTPExp),
	)

	// Rate Limitter
	rt := rater.GetRater(app_context,
		rater.RegisterRateLimitForEachClient(cfg.App.RaterLimit),
		rater.RegisterBurstLimitForEachClient(cfg.App.BurstLimit),
		rater.RegisterEvaluationInterval(cfg.App.RaterEvaluationInterval),
		rater.RegisterDeletionTime(cfg.App.RaterDeletionTime),
	)

	// Mailer
	ml := mailer.GetMailer(
		mailer.RegisterSenderAddress(cfg.App.MailerEmailAddress),
		mailer.RegisterSenderPassword(cfg.App.MailerEmailPassword),
		mailer.RegisterTemplatePath(cfg.App.MailerTemplatePath),
	)

	// PubSub
	ps := pubsub.GetPubSubClient(
		app_context,
		pubsub.RegisterEnv(os.Getenv("GO_ENV")),
		pubsub.RegisterProjectId(cfg.App.GoogleProjectId),
		pubsub.RegisterKey(cfg.App.GoogleServiceAccountPath),
	)

	// Firebase bucket
	bkt := bucket.GetFirebaseBucket(app_context, cfg.Bucket.BucketName, cfg.Bucket.ServiceAccountPath)

	// Composers .-.
	serviceComposer := composer.NewServiceComposer(dk, rt, ml, bkt, rdis, ps)
	repoComposer := composer.NewRepoComposer(pg, rdis, mg, cfg.App.Environment)
	usecaseComposer := composer.NewUseCaseComposer(repoComposer, serviceComposer)

	// Http
	var deliveree *gin.Engine
	if os.Getenv("GO_ENV") == "PRODUCTION" {
		deliveree = gin.New()
		deliveree.Use(gin.Recovery())
	} else {
		deliveree = gin.Default()
	}
	v2.NewRouter(deliveree, logger, usecaseComposer)

	httpserver.NewServer(deliveree,
		httpserver.RegisterHostAndPort(cfg.Server.Host, cfg.Server.Port),
	)

	// Waiting Signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	httpserver.Shutdown()
	defer cancel()
}
