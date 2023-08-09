package pubsub

import "sinarlog.com/config"

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
		switch ps.env {
		case config.DEVELOPMENT:
			ps.keyPath = path
		case config.TESTING:
			ps.keyPath = path
		}
	}
}

func RegisterProjectId(id string) Option {
	return func(ps *PubSub) {
		ps.projectId = id
	}
}
