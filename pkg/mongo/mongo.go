package mongo

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	once                sync.Once
	mongoSingleInstance *Mongo
)

var (
	_defaultMaxPoolSize     uint64 = 2
	_defaultMaxOpenConn     uint64 = 2
	_defaultMaxConnAttempts        = 10
	_defaultConnLifeTime           = 5 * time.Minute
)

type Mongo struct {
	maxPoolSize     uint64
	maxOpenConn     uint64
	maxConnAttempts int
	dbName          string
	uri             string
	debug           bool
	connLifetime    time.Duration
	Conn            *mongo.Database
}

func GetMongoClient(ctx context.Context, opts ...Option) *Mongo {
	if mongoSingleInstance == nil {
		once.Do(func() {
			mongoSingleInstance = &Mongo{
				maxPoolSize:     _defaultMaxPoolSize,
				maxOpenConn:     _defaultMaxOpenConn,
				connLifetime:    _defaultConnLifeTime,
				maxConnAttempts: _defaultMaxConnAttempts,
			}

			for _, opt := range opts {
				opt(mongoSingleInstance)
			}

			mongoSingleInstance.connect(ctx)
		})
	}

	return mongoSingleInstance
}

func (mg *Mongo) connect(ctx context.Context) {
	opts := options.Client().
		ApplyURI(mongoSingleInstance.uri).
		SetAppName("SinarLog").
		SetMaxPoolSize(mongoSingleInstance.maxPoolSize).
		SetMaxConnIdleTime(mongoSingleInstance.connLifetime).
		SetMaxConnecting(mongoSingleInstance.maxOpenConn)

	if mg.debug {
		cmdMonitor := &event.CommandMonitor{
			Started: func(_ context.Context, evt *event.CommandStartedEvent) {
				log.Println(evt.Command)
			},
		}
		opts = opts.SetMonitor(cmdMonitor)
	}

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatalf("unable to connect to mongo server: %s", err)
	}

	mg.ping(ctx, client)

	mg.Conn = client.Database(mg.dbName)
}

func (mg *Mongo) ping(ctx context.Context, client *mongo.Client) {
	toCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for mg.maxConnAttempts > 0 {
		mg.maxConnAttempts--

		if err := client.Ping(ctx, nil); err != nil {
			log.Fatalf("unable to ping mongo server: %s", err)
		}

		err := client.Ping(toCtx, nil)
		if err == nil {
			log.Printf("INFO - Successfully connected to mongo database after %d attempt(s)", _defaultMaxConnAttempts-mg.maxConnAttempts)
			return
		}
	}

	log.Fatalf("FATAL - Unable to ping database")
}
