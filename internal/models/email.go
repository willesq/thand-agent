package models

type EmailNotificationRequest struct {
	From    string
	To      []string
	Subject string
	Body    EmailNotificationBody
	Headers map[string][]string
}

type EmailNotificationBody struct {
	Text string
	HTML string
}
