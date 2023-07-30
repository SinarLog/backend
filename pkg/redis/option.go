package redis

import "time"

type Option func(*RedisClient)

func RegisterAddress(addr string) Option {
	return func(r *RedisClient) {
		r.host = addr
	}
}

func RegisterPassword(pass string) Option {
	return func(r *RedisClient) {
		r.password = pass
	}
}

func RegisterDB(db int) Option {
	return func(r *RedisClient) {
		r.db = db
	}
}

func RegisterReadTimeout(t time.Duration) Option {
	return func(r *RedisClient) {
		if t > 0 {
			r.readTimeout = t
		}
	}
}

func RegisterWriteTimeout(t time.Duration) Option {
	return func(r *RedisClient) {
		if t > 0 {
			r.writeTimeout = t
		}
	}
}

func RegisterMinIdleConn(conn int) Option {
	return func(r *RedisClient) {
		if conn > 0 {
			r.minIdleConn = conn
		}
	}
}

func RegisterMaxIdleConn(conn int) Option {
	return func(r *RedisClient) {
		if conn > 0 {
			r.maxIdleConn = conn
		}
	}
}

func RegisterMaxIdleTime(t time.Duration) Option {
	return func(r *RedisClient) {
		if t > 0 {
			r.maxIdleTime = t
		}
	}
}
