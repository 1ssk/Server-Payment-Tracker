# Развёртывание «Биллинг серверов» на сервере с Nginx

Краткая инструкция по развёртыванию на Linux с доменом и Nginx.

## 1. Подготовка

- Установите Docker и docker compose.
- Nginx слушает 80/443.
- DNS: A/AAAA‑запись домена на IP сервера.

## 2. Конфигурация

Клонируйте репозиторий и настройте переменные окружения в `docker-compose.yml`:

```yaml
services:
  app:
    build: .
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      - DB_PATH=/data/app.db
      - ADMIN_USERNAME=ваш_логин
      - ADMIN_PASSWORD=ваш_пароль
      - AUTH_TOKEN=длинный_секретный_токен
    volumes:
      - db-data:/data

volumes:
  db-data:
```

- **ADMIN_USERNAME** и **ADMIN_PASSWORD** — обязательны, без них вход в систему недоступен (сервис вернёт «auth not configured»).
- **AUTH_TOKEN** — секрет для заголовка `Authorization: Bearer …`; задайте свой для production.

## 3. Сборка и запуск

```bash
docker compose build
docker compose up -d
docker compose logs -f
```

Приложение доступно на `http://127.0.0.1:8080` (проксируйте через Nginx).

## 4. Nginx с доменом

Пример `/etc/nginx/sites-available/billing`:

```nginx
server {
    listen 80;
    server_name billing.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Включите сайт и перезагрузите Nginx:

```bash
ln -s /etc/nginx/sites-available/billing /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx
```

## 5. HTTPS

```bash
apt install certbot python3-certbot-nginx
certbot --nginx -d billing.example.com
```

## 6. Авторизация

- Логин и пароль задаются **только** в переменных окружения (`ADMIN_USERNAME`, `ADMIN_PASSWORD`). В коде приложения дефолтных значений нет.
- Пользователь вводит учётные данные на форме входа → фронт отправляет `POST /api/login` → бэкенд сравнивает с переменными окружения и при совпадении возвращает токен.
- Фронт сохраняет токен и подставляет его в заголовок `Authorization: Bearer <токен>` при каждом запросе к API. Без токена или с неверным токеном бэкенд возвращает 401.

## 7. Напоминания и отчётность

- В настройках (иконка «Настройки») задаётся SMTP и **за сколько дней до срока** присылать напоминание (по умолчанию 10). Раз в час бэкенд проверяет сервера с предстоящим платежом и отправляет письмо на указанный адрес.
- Подтверждение оплаты: на карточке сервера кнопка «Подтвердить оплату» — указываете дату и сумму; дата следующего платежа пересчитывается автоматически.
- Вкладка «Отчётность»: выбор периода и сервера, таблица расходов, выгрузка в CSV.

## 8. Типичные проблемы

- **«auth not configured»** — не заданы `ADMIN_USERNAME` или `ADMIN_PASSWORD`. Задайте их в `docker-compose.yml` и перезапустите контейнер.
- **401 при запросах** — выйдите и войдите снова (токен мог измениться или истёк).
- **Письма не приходят** — проверьте SMTP в настройках, кнопка «Отправить тестовое письмо»; убедитесь, что уведомления включены и указано «Кому (To)».
