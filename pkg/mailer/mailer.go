package mailer

import (
	"bytes"
	"log"
	"sync"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	host     string
	port     int
	address  string
	password string

	Dialer      *gomail.Dialer
	MessagePool sync.Pool
	BufferPool  sync.Pool

	TemplatePath string
}

var (
	_defaultMailerHost = "smtp.gmail.com"
	_defaultMailerPort = 587
	_defaultMailerPath = "public"
)

var (
	once                 sync.Once
	mailerSingleInstance *Mailer
)

func GetMailer(opts ...Option) *Mailer {
	if mailerSingleInstance == nil {
		once.Do(func() {
			mailerSingleInstance = &Mailer{
				host:         _defaultMailerHost,
				port:         _defaultMailerPort,
				TemplatePath: _defaultMailerPath,
				MessagePool: sync.Pool{
					New: func() any {
						message := gomail.NewMessage()
						return message
					},
				},
				BufferPool: sync.Pool{
					New: func() any {
						buf := new(bytes.Buffer)
						return buf
					},
				},
			}

			for _, opt := range opts {
				opt(mailerSingleInstance)
			}

			dialer := gomail.NewDialer(
				mailerSingleInstance.host,
				mailerSingleInstance.port,
				mailerSingleInstance.address,
				mailerSingleInstance.password,
			)

			_, err := dialer.Dial()
			if err != nil {
				log.Fatalf("Unable to dial to the given mail credentials %s\n", err)
			}

			mailerSingleInstance.Dialer = dialer
		})
	}

	return mailerSingleInstance
}

func (m *Mailer) GetSenderAddress() string {
	return m.address
}
