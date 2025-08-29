package email

import (
	"fmt"
	"log"
	"net/smtp"
	"new-billing/internal/config"
)

type EmailService struct {
	config *config.SMTPConfig
}

func NewEmailService(cfg *config.SMTPConfig) *EmailService {
	return &EmailService{
		config: cfg,
	}
}

func (es *EmailService) SendNewUserEmail(username, password, email string) error {
	if !es.config.Enabled {
		log.Printf("Email disabled, would send: username=%s, password=%s to %s", username, password, email)
		return nil
	}

	subject := "Ваши данные для входа в систему Ariadna Billing"
	body := fmt.Sprintf(`Здравствуйте!

Для вас был создан аккаунт в системе Ariadna Billing.

Данные для входа:
Логин: %s
Пароль: %s

Вы также можете войти в систему, используя номер любого из ваших договоров с тем же паролем.

С уважением,
Администрация Ariadna Billing`, username, password)

	return es.sendEmail(email, subject, body)
}

func (es *EmailService) SendSupportTicketEmail(email, ticketTitle, message string, isReply bool) error {
	if !es.config.Enabled {
		log.Printf("Email disabled, would send support email to %s: %s", email, ticketTitle)
		return nil
	}

	var subject string
	var body string

	if isReply {
		subject = fmt.Sprintf("Ответ на обращение: %s", ticketTitle)
		body = fmt.Sprintf(`Здравствуйте!

На ваше обращение "%s" получен ответ:

%s

Для ответа или уточнений, пожалуйста, воспользуйтесь системой обращений.

С уважением,
Служба поддержки Ariadna Billing`, ticketTitle, message)
	} else {
		subject = fmt.Sprintf("Новое обращение создано: %s", ticketTitle)
		body = fmt.Sprintf(`Здравствуйте!

Ваше обращение "%s" было получено и зарегистрировано в системе.

Описание обращения:
%s

Мы рассмотрим ваш запрос и свяжемся с вами в ближайшее время.

С уважением,
Служба поддержки Ariadna Billing`, ticketTitle, message)
	}

	return es.sendEmail(email, subject, body)
}

func (es *EmailService) sendEmail(to, subject, body string) error {
	if !es.config.Enabled {
		return nil
	}

	auth := smtp.PlainAuth("", es.config.Username, es.config.Password, es.config.Host)

	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		to, subject, body)

	addr := fmt.Sprintf("%s:%d", es.config.Host, es.config.Port)
	
	err := smtp.SendMail(addr, auth, es.config.From, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("Email sent successfully to %s", to)
	return nil
}