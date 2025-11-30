package thand

import (
	"bytes"
	_ "embed"
	"html/template"

	"github.com/sirupsen/logrus"
)

//go:embed email_template.html
var emailTemplateHTML string

//go:embed approval_email_content.html
var approvalEmailContentHTML string

//go:embed authorize_email_content.html
var authorizeEmailContentHTML string

//go:embed revoke_email_content.html
var revokeEmailContentHTML string

//go:embed form_email_content.html
var formEmailContentHTML string

// EmailData is a simple struct for email template data
type EmailData struct {
	Title   string
	Content template.HTML
}

var emailTemplate *template.Template
var approvalContentTemplate *template.Template
var authorizeContentTemplate *template.Template
var revokeContentTemplate *template.Template
var formContentTemplate *template.Template

func init() {
	var err error

	emailTemplate, err = template.New("email").Parse(emailTemplateHTML)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to parse email template")
	}

	approvalContentTemplate, err = template.New("approval_content").Parse(approvalEmailContentHTML)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to parse approval content template")
	}

	authorizeContentTemplate, err = template.New("authorize_content").Parse(authorizeEmailContentHTML)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to parse authorize content template")
	}

	revokeContentTemplate, err = template.New("revoke_content").Parse(revokeEmailContentHTML)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to parse revoke content template")
	}

	formContentTemplate, err = template.New("form_content").Parse(formEmailContentHTML)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to parse form content template")
	}
}

// RenderEmail renders a simple HTML email with title and content
func RenderEmail(title string, content string) (string, error) {
	data := EmailData{
		Title:   title,
		Content: template.HTML(content),
	}

	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, data); err != nil {
		logrus.WithError(err).Error("Failed to execute email template")
		return "", err
	}

	return buf.String(), nil
}

// RenderEmailWithTemplate renders an email using a content template and data map
func RenderEmailWithTemplate(title string, contentTemplate *template.Template, data map[string]any) (string, error) {
	// First render the content template with the data
	var contentBuf bytes.Buffer
	if err := contentTemplate.Execute(&contentBuf, data); err != nil {
		logrus.WithError(err).Error("Failed to execute content template")
		return "", err
	}

	// Then wrap it in the main email template
	return RenderEmail(title, contentBuf.String())
}

// GetApprovalContentTemplate returns the approval content template
func GetApprovalContentTemplate() *template.Template {
	return approvalContentTemplate
}

// GetAuthorizeContentTemplate returns the authorize content template
func GetAuthorizeContentTemplate() *template.Template {
	return authorizeContentTemplate
}

// GetRevokeContentTemplate returns the revoke content template
func GetRevokeContentTemplate() *template.Template {
	return revokeContentTemplate
}

// GetFormContentTemplate returns the form content template
func GetFormContentTemplate() *template.Template {
	return formContentTemplate
}
