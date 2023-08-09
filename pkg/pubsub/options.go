package pubsub

import "log"

// Option -.
type Option func(*PubSub)

// SetEnv
func RegisterEnv(env string) Option {
	return func(ps *PubSub) {
		ps.env = env
	}
}

// SetServiceAccountPath (only for development)
func RegisterKey(path string) Option {
	return func(ps *PubSub) {
		if ps.env == "PRODUCTION" {
			log.Fatalln("reading key as json is not allowed in production")
		}
		ps.keyPath = path
	}
}

func RegisterProjectId(id string) Option {
	return func(ps *PubSub) {
		ps.projectId = id
	}
}
