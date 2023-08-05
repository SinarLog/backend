package mongo

import "time"

// Option -.
type Option func(*Mongo)

// URI -.
func RegisterURI(uri string) Option {
	return func(m *Mongo) {
		m.uri = uri
	}
}

// DbNamme -.
func RegisterDbName(db string) Option {
	return func(m *Mongo) {
		m.dbName = db
	}
}

// MaxPoolSize -.
func MaxPoolSize(size int) Option {
	return func(m *Mongo) {
		if size > 0 {
			m.maxPoolSize = uint64(size)
		}
	}
}

// ConnLifetime -.
func MaxConnLifetime(t time.Duration) Option {
	return func(m *Mongo) {
		if t > 0 {
			m.connLifetime = t
		}
	}
}

// ConnTimeout -.
func MaxConn(size int) Option {
	return func(m *Mongo) {
		if size > 0 {
			m.maxOpenConn = uint64(size)
		}
	}
}
