package service

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/mail.v2"
	"log"
	"os"
)

func SendMail(to, url string) error {
	m := mail.NewMessage()

	mailName := os.Getenv("MAIL_NAME")
	mailPassword := os.Getenv("MAIL_PASSWORD")

	m.SetHeader("From", mailName)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Приглашение на собеседование")

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2c3e50; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f8f9fa; padding: 30px; border: 1px solid #dee2e6; }
        .button { 
            display: inline-block; 
            background: #007bff; 
            color: white; 
            padding: 12px 30px; 
            text-decoration: none; 
            border-radius: 5px; 
            margin: 20px 0;
            font-weight: bold;
        }
        .important { background: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .footer { background: #6c757d; color: white; padding: 15px; text-align: center; font-size: 12px; border-radius: 0 0 5px 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🤖 Приглашение на AI-собеседование</h1>
        </div>
        
        <div class="content">
            <h2>Здравствуйте!</h2>
            
            <p>Мы рады сообщить, что ваше резюме прошло первичный отбор! </p>
            
            <p>Приглашаем вас пройти <strong>AI-скрининг интервью</strong> на позицию <strong>%s</strong>.</p>
            
            <h3>Что вас ждет:</h3>
            <ul>
                <li>📝 Интерактивное собеседование с AI-ассистентом</li>
                <li>⏱️ Длительность: 20-30 минут</li>
                <li>💬 Вопросы по вашему опыту и техническим навыкам</li>
                <li>🔄 Возможность уточнить детали в режиме реального времени</li>
            </ul>
            
            <div style="text-align: center;">
                <a href="%s" class="button">НАЧАТЬ СОБЕСЕДОВАНИЕ</a>
            </div>
            
            <div class="important">
                <strong>⚠️ Важно:</strong>
                <br>• Ссылка действительна в течение <strong>7 дней</strong> с момента получения
                <br>• Рекомендуем проходить собеседование в тихой обстановке
                <br>• Подготовьте информацию о своем опыте работы
                <br>• В случае технических проблем свяжитесь с нами
            </div>
            
            <h3>Как это работает:</h3>
            <ol>
                <li>Перейдите по ссылке выше</li>
                <li>AI-ассистент поприветствует вас и объяснит процесс</li>
                <li>Отвечайте на вопросы честно и подробно</li>
                <li>По завершении получите обратную связь</li>
            </ol>
            
            <p>Мы ценим ваше время и интерес к нашей компании. Желаем удачи! </p>
            
            <hr style="margin: 30px 0; border: none; border-top: 1px solid #dee2e6;">
            
            <p style="font-size: 14px; color: #6c757d;">
                Если у вас возникли вопросы, ответьте на это письмо или свяжитесь с нами по телефону: 
                <strong>+7 (xxx) xxx-xx-xx</strong>
            </p>
        </div>
        
        <div class="footer">
            <p>© 2024 Название компании. Все права защищены.</p>
            <p>Это автоматическое сообщение, пожалуйста, не отвечайте на него.</p>
        </div>
    </div>
</body>
</html>
`, "vacancy_title", "interview_url")

	m.SetBody("text/html", body)

	d := mail.NewDialer("smtp.mail.ru", 465, mailName, mailPassword)

	d.TLSConfig = &tls.Config{
		ServerName:         "smtp.mail.ru",
		InsecureSkipVerify: false,
	}

	return d.DialAndSend(m)
}

func main() {
	err := SendMail("qwerty.mart8@gmail.com", "http...")
	if err != nil {
		log.Printf("Ошибка: %v", err)
	} else {
		log.Println("Email успешно отправлен!")
	}
}
