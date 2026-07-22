package lib

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
	"time"
	_ "time/tzdata"
)

// SendProspectMail sends the prospecting email to a single recipient.
// title  – "Mr." or "Ms." derived from the prospect's gender.
// name   – prospect name already uppercased by the caller.
// to     – prospect email address.
func SendProspectMail(to, title, name string) error {

	slog.Info("SendProspectMail: begin", "to", to)

	from := os.Getenv("SENDER")
	password := os.Getenv("ZOHO_PWD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	subject := "Remote Senior Engineer for immediate E2E backend / infra / fullstack sprints"

	body := fmt.Sprintf(`Hi %s %s,

Many engineering teams face bottlenecks where legacy debt slows down feature delivery and inflates Cloud costs.

I am a Senior Software Engineer specializing in E2E system modernization and cloud automation. I help managers safely modernize complex legacy monoliths (especially into highly optimized, asynchronous Cloud native Go or Python/FastAPI architectures), potentially reducing deployment friction and cloud spend by 40-60%%, while delivering also precise, responsive and intuitive UI/UX.

Because I own the complete lifecycle - from needs analysis and solution design to IaC (Terraform) and production monitoring - I operate as an autonomous execution unit without requiring daily management. I also leverage safe AI-augmented workflows to compress delivery timelines significantly.

I operate under an independent B2B contractor framework aligned with your working hours (using standard W-8BEN compliance for seamless US invoicing), meaning zero payroll or HR overhead for your company.

If you need an extra senior engineer to clear legacy bottlenecks or accelerate cloud migrations, let's connect.

Best regards,

Alexandre NGUYEN

Email: modernization@steadypartner.online
Github: github.com/alexGithub202021
LinkedIn: https://www.linkedin.com/in/alexandre-nguyen-senior-swe`,
		title, name)

	// Build raw MIME message.
	header := map[string]string{
		"From":         from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": `text/plain; charset="utf-8"`,
	}
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Establish implicit TLS connection.
	conn, err := tls.Dial("tcp", smtpHost+":"+smtpPort, &tls.Config{ServerName: smtpHost})
	if err != nil {
		return fmt.Errorf("SendProspectMail: TLS dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("SendProspectMail: SMTP client: %w", err)
	}
	defer client.Quit()

	auth := smtp.PlainAuth("", from, password, smtpHost)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SendProspectMail: auth: %w", err)
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("SendProspectMail: MAIL cmd: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("SendProspectMail: RCPT cmd: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SendProspectMail: DATA cmd: %w", err)
	}
	if _, err = w.Write([]byte(message)); err != nil {
		return fmt.Errorf("SendProspectMail: write body: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("SendProspectMail: close writer: %w", err)
	}

	slog.Info("SendProspectMail: sent successfully", "to", to)
	return nil
}

// SendMail is kept for backward compatibility. It is the original
// hard-coded scheduled-mail call used during early testing.
func SendMail() {

	slog.Info("begin SendMail")

	nyLoc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		slog.Error("Failed to load Ho_Chi_Minh location", slog.String("error msg", err.Error()))
	}
	targetTime := time.Date(2026, 7, 15, 21, 45, 0, 0, nyLoc)

	singaporeLoc, err := time.LoadLocation("Asia/Singapore")
	if err != nil {
		slog.Error("Failed to load singapore location", slog.String("error msg", err.Error()))
	}
	convertedTime := targetTime.In(singaporeLoc)
	scheduledString := convertedTime.Format("2006-01-02 15:04:05")

	from := os.Getenv("SENDER")
	password := os.Getenv("ZOHO_PWD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	to := "alex@steadypartner.online"

	header := make(map[string]string)
	header["From"] = from
	header["To"] = to
	header["Subject"] = "test emails scheduling"
	header["X-DELIVER-AT"] = scheduledString
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\nScheduled body..."

	conn, err := tls.Dial("tcp", smtpHost+":"+smtpPort, &tls.Config{ServerName: smtpHost})
	if err != nil {
		slog.Error("TLS dial failed", slog.String("error msg", err.Error()))
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		slog.Error("SMTP client creation failed", slog.String("error msg", err.Error()))
	}
	defer client.Quit()

	auth := smtp.PlainAuth("", from, password, smtpHost)
	if err = client.Auth(auth); err != nil {
		slog.Error("Authentication failed", slog.String("error msg", err.Error()))
	}

	if err = client.Mail(from); err != nil {
		slog.Error("MAIL command failed", slog.String("error msg", err.Error()))
	}
	if err = client.Rcpt(to); err != nil {
		slog.Error("RCPT command failed", slog.String("error msg", err.Error()))
	}

	w, err := client.Data()
	if err != nil {
		slog.Error("DATA command failed", slog.String("error msg", err.Error()))
	}
	_, err = w.Write([]byte(message))
	if err != nil {
		slog.Error("Failed to write body", slog.String("error msg", err.Error()))
	}
	err = w.Close()
	if err != nil {
		slog.Error("Failed to close data writer", slog.String("error msg", err.Error()))
	}

	slog.Info("Scheduled email configured and sent to Zoho successfully.")
}
