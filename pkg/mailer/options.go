package mailer

type Option func(*Mailer)

func RegisterSenderAddress(address string) Option {
	return func(m *Mailer) {
		m.address = address
	}
}

func RegisterSenderPassword(pass string) Option {
	return func(m *Mailer) {
		m.password = pass
	}
}

func RegisterHost(host string) Option {
	return func(m *Mailer) {
		m.host = host
	}
}

func RegisterPort(port int) Option {
	return func(m *Mailer) {
		m.port = port
	}
}
