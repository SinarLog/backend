package service

import (
	"bytes"
	"html/template"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
	"sinarlog.com/internal/utils"
	"sinarlog.com/pkg/mailer"
)

const (
	OTP                           string = "OTP"
	CRED                          string = "CRED"
	FORGOT_PASSWORD               string = "FORGOT_PASSWORD"
	OVERTIME_SUBMISSION           string = "OVERTIME_SUBMISSION"
	PROCESSED_OVERTIME_SUBMISSION string = "PROCESSED_OVERTIME_SUBMISSION"
	PROCESSED_LEAVE_BY_MANAGER    string = "PROCESSED_LEAVE_BY_MANAGER"
	PROCESSED_LEAVE_BY_HR         string = "PROCESSED_LEAVE_BY_HR"
	FWD_LEAVE_PROPOSAL            string = "FWD_LEAVE_PROPOSAL"
)

type mailerService struct {
	mailer *mailer.Mailer
}

func NewMailerService(ml *mailer.Mailer) *mailerService {
	return &mailerService{mailer: ml}
}

func (s *mailerService) SendEmail(receiver string, mailType string, data map[string]any) error {
	message := s.mailer.MessagePool.Get().(*gomail.Message)
	body := s.mailer.BufferPool.Get().(*bytes.Buffer)
	defer s.mailer.MessagePool.Put(message)
	defer s.mailer.BufferPool.Put(body)
	message.Reset()
	body.Reset()

	message.SetHeader("From", s.mailer.GetSenderAddress())
	message.SetHeader("To", receiver)

	s.loadBody(message, body, mailType, data)

	message.SetBody("text/html", body.String())
	message.Embed("public/sinarlog.png")
	message.Embed("public/sinarmas.png")

	if err := s.mailer.Dialer.DialAndSend(message); err != nil {
		return err
	}

	return nil
}

func (s *mailerService) loadBody(message *gomail.Message, body *bytes.Buffer, mailType string, data map[string]any) {
	switch mailType {
	case OTP:
		now := time.Now().In(utils.CURRENT_LOC).Local()
		timeString := now.Weekday().String() + ", " + strings.Split(now.String(), " ")[0]

		t := template.Must(template.ParseFiles(s.mailer.TemplatePath + "/otp.gohtml"))
		t.ExecuteTemplate(body, "otp.gohtml", data)

		if data["Action"].(string) == "Clock In" {
			message.SetHeader("Subject", "Clock In OTP "+timeString)
		} else {
			message.SetHeader("Subject", "Update Password OTP "+timeString)
		}
	case CRED:
		t := template.Must(template.ParseFiles(s.mailer.TemplatePath + "/credential.gohtml"))
		t.ExecuteTemplate(body, "credential.gohtml", data)
		message.SetHeader("Subject", "Welcome to SinarLog!")
	case FORGOT_PASSWORD:
		t := template.Must(template.ParseFiles(s.mailer.TemplatePath + "/forgot_password.gohtml"))
		t.ExecuteTemplate(body, "forgot_password.gohtml", data)
		message.SetHeader("Subject", "Forgot Password")
	case OVERTIME_SUBMISSION:
		t := template.Must(template.ParseFiles(s.mailer.TemplatePath + "/overtime_submission.gohtml"))
		t.ExecuteTemplate(body, "overtime_submission.gohtml", data)
		message.SetHeader("Subject", "Staff Submitted Overtime")
	case PROCESSED_OVERTIME_SUBMISSION:
		t := template.Must(template.ParseFiles(s.mailer.TemplatePath + "/processed_overtime_submission.gohtml"))
		t.ExecuteTemplate(body, "processed_overtime_submission.gohtml", data)
		message.SetHeader("Subject", "Your Overtime Submission Has Been Processed")
	case PROCESSED_LEAVE_BY_MANAGER:
		t := template.Must(template.ParseFiles(s.mailer.TemplatePath + "/processed_leave_by_manager.gohtml"))
		t.ExecuteTemplate(body, "processed_leave_by_manager.gohtml", data)
		message.SetHeader("Subject", "Leave Request Status Update")
	case PROCESSED_LEAVE_BY_HR:
		t := template.Must(template.ParseFiles(s.mailer.TemplatePath + "/processed_leave_by_hr.gohtml"))
		t.ExecuteTemplate(body, "processed_leave_by_hr.gohtml", data)
		message.SetHeader("Subject", "Your Leave Request Has Been Processed")
	case FWD_LEAVE_PROPOSAL:
		t := template.Must(template.ParseFiles(s.mailer.TemplatePath + "/forward_leave_proposal.gohtml"))
		t.ExecuteTemplate(body, "forward_leave_proposal.gohtml", data)
		message.SetHeader("Subject", "Your Staff is Requesting a Leave")
	}
}
