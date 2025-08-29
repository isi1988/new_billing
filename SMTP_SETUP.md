# 📧 Настройка SMTP для Email Уведомлений

## Функции Email Уведомлений

Система отправляет email уведомления в следующих случаях:
1. **Создание нового пользователя для клиента** - отправляет данные для входа (логин и пароль)
2. **Комментарии к заявкам** - уведомляет о новых комментариях в системе поддержки

## Настройка SMTP

### 1. Через Docker Compose (Рекомендуется для продакшена)

Отредактируйте `docker-compose.prod.yml` и замените значения переменных:

```yaml
environment:
  SMTP_HOST: smtp.gmail.com
  SMTP_PORT: 587
  SMTP_USERNAME: your-email@gmail.com
  SMTP_PASSWORD: your-app-password
  SMTP_FROM: your-email@gmail.com
  SMTP_ENABLED: true  # Измените на true для включения
```

### 2. Через конфигурационный файл config.yaml

```yaml
smtp:
  host: "smtp.gmail.com"
  port: 587
  username: "your-email@gmail.com"
  password: "your-app-password"
  from: "your-email@gmail.com"
  enabled: true
```

## Настройка для Популярных Провайдеров

### Gmail
```
Host: smtp.gmail.com
Port: 587
Требуется App Password (не обычный пароль)
```

**Как создать App Password для Gmail:**
1. Перейдите в Google Account Settings
2. Security → 2-Step Verification (должна быть включена)
3. App passwords → Generate password for "Mail"
4. Используйте сгенерированный пароль в SMTP_PASSWORD

### Yandex Mail
```
Host: smtp.yandex.ru
Port: 465 (SSL) или 587 (TLS)
Username: your-email@yandex.ru
```

### Outlook/Hotmail
```
Host: smtp-mail.outlook.com
Port: 587
Username: your-email@outlook.com
```

### Mail.ru
```
Host: smtp.mail.ru
Port: 465 (SSL) или 587 (TLS)
Username: your-email@mail.ru
```

## Тестирование

1. Включите SMTP (`SMTP_ENABLED=true`)
2. Создайте нового клиента с указанием email
3. Проверьте, что email с данными для входа отправлен
4. Создайте заявку и добавьте комментарий - должно прийти уведомление

## Безопасность

⚠️ **Важно:**
- Никогда не используйте основной пароль email для SMTP
- Используйте App Passwords или специальные SMTP пароли
- Храните учетные данные в переменных окружения, не в коде
- Регулярно обновляйте пароли

## Отладка

Если email не отправляются:

1. Проверьте логи контейнера:
```bash
docker logs ariadna-backend
```

2. Убедитесь что `SMTP_ENABLED=true`

3. Проверьте правильность настроек SMTP
4. Убедитесь что почтовый провайдер разрешает SMTP подключения
5. Проверьте, что используете правильный тип аутентификации (App Password для Gmail)

## Шаблоны Email

Система использует следующие шаблоны:

### Новый пользователь
```
Тема: Ваши данные для входа в систему Ariadna Billing

Здравствуйте!

Для вас был создан аккаунт в системе Ariadna Billing.

Данные для входа:
Логин: 000123
Пароль: abc12345

Вы также можете войти в систему, используя номер любого из ваших договоров с тем же паролем.

С уважением,
Администрация Ariadna Billing
```

### Комментарий к заявке
```
Тема: Ответ на обращение: [Название заявки]

Здравствуйте!

На ваше обращение "[Название заявки]" получен ответ:

[Текст комментария]

Для ответа или уточнений, пожалуйста, воспользуйтесь системой обращений.

С уважением,
Служба поддержки Ariadna Billing
```