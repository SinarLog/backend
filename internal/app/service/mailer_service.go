package service

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

type IMailerService interface {
	SendEmail(receiver string, mailType string, data map[string]any) error
}
