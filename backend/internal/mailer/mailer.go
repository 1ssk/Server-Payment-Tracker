package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"strings"

	"github.com/anyx/serversanyx-backend/internal/models"
)

// SendReminder отправляет красивое напоминание об оплате за N дней до срока.
func SendReminder(settings models.SMTPSettings, to, serverName, paymentDate string, amount float64, daysLeft int) error {
	body := buildReminderHTML(serverName, paymentDate, amount, daysLeft)
	subject := fmt.Sprintf("Напоминание: оплата «%s» через %d %s", serverName, daysLeft, daysWord(daysLeft))
	return sendMail(settings, to, subject, body)
}

// SendTestEmail отправляет тестовое письмо для проверки SMTP.
func SendTestEmail(settings models.SMTPSettings, to string) error {
	body := `<div style="font-family: sans-serif; max-width: 480px; margin: 0 auto; padding: 24px;">
<div style="background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%); color: white; padding: 20px; border-radius: 12px 12px 0 0; text-align: center;">
<h2 style="margin: 0;">Биллинг серверов</h2>
<p style="margin: 8px 0 0; opacity: 0.9;">Тестовое уведомление</p>
</div>
<div style="padding: 24px; background: #f9fafb; border: 1px solid #e5e7eb; border-top: none; border-radius: 0 0 12px 12px;">
<p style="margin: 0 0 16px; color: #374151;">Если вы получили это письмо, настройки SMTP работают корректно.</p>
<p style="margin: 0; font-size: 12px; color: #6b7280;">Система учёта серверов</p>
</div></div>`
	return sendMail(settings, to, "Тест уведомлений — Биллинг серверов", body)
}

func daysWord(d int) string {
	if d == 1 {
		return "день"
	}
	if d >= 2 && d <= 4 {
		return "дня"
	}
	return "дней"
}

func buildReminderHTML(serverName, paymentDate string, amount float64, daysLeft int) string {
	tpl := `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"></head>
<body style="margin: 0; padding: 20px; background: #f3f4f6;">
<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 520px; margin: 0 auto;">
  <div style="background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%); color: white; padding: 28px 24px; border-radius: 16px 16px 0 0; text-align: center;">
    <h1 style="margin: 0; font-size: 22px;">Напоминание об оплате</h1>
    <p style="margin: 10px 0 0; opacity: 0.95;">Сервер: {{.ServerName}}</p>
  </div>
  <div style="background: white; padding: 28px 24px; border: 1px solid #e5e7eb; border-top: none; border-radius: 0 0 16px 16px; box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.05);">
    <p style="margin: 0 0 20px; color: #374151; font-size: 15px;">До срока оплаты осталось <strong>{{.DaysLeft}} {{.DaysWord}}</strong>.</p>
    <table style="width: 100%; border-collapse: collapse;">
      <tr style="border-bottom: 1px solid #e5e7eb;"><td style="padding: 12px 0; color: #6b7280;">Сумма</td><td style="padding: 12px 0; text-align: right; font-weight: 600;">₽{{.Amount}}</td></tr>
      <tr style="border-bottom: 1px solid #e5e7eb;"><td style="padding: 12px 0; color: #6b7280;">Дата платежа</td><td style="padding: 12px 0; text-align: right;">{{.PaymentDate}}</td></tr>
    </table>
    <p style="margin: 20px 0 0; font-size: 13px; color: #6b7280;">Подтвердите оплату в личном кабинете, чтобы мы перестали напоминать об этом сроке.</p>
    <p style="margin: 24px 0 0; font-size: 12px; color: #9ca3af;">Система учёта серверов</p>
  </div>
</div>
</body>
</html>`
	t, _ := template.New("reminder").Parse(tpl)
	var buf bytes.Buffer
	_ = t.Execute(&buf, map[string]interface{}{
		"ServerName":  serverName,
		"PaymentDate": paymentDate,
		"Amount":      fmt.Sprintf("%.2f", amount),
		"DaysLeft":    daysLeft,
		"DaysWord":    daysWord(daysLeft),
	})
	return buf.String()
}

func sendMail(settings models.SMTPSettings, to, subject, bodyHTML string) error {
	if settings.Host == "" || settings.To == "" {
		return fmt.Errorf("smtp not configured")
	}
	addr := fmt.Sprintf("%s:%d", settings.Host, settings.Port)
	hostForAuth := strings.Split(settings.Host, ":")[0]
	auth := smtp.PlainAuth("", settings.Username, settings.Password, hostForAuth)
	from := settings.From
	if from == "" {
		from = settings.Username
	}
	msg := []byte(
		"From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" + bodyHTML)
	err := smtp.SendMail(addr, auth, from, []string{to}, msg)
	if err != nil {
		log.Printf("mailer: send failed: %v", err)
		return err
	}
	return nil
}
