## Развёртывание Server Management System на сервере с Nginx

Ниже — краткий сценарий для production‑развёртывания на Linux‑сервере с доменом и Nginx.

### 1. Подготовка сервера

1. Установите Docker и docker compose.
2. Убедитесь, что Nginx установлен и слушает порт 80/443.
3. Добавьте DNS‑запись A/AAAA для домена (например, `example.com`) на IP сервера.

### 2. Клонирование и конфигурация проекта

```bash
git clone <ваш-репозиторий> /opt/servers
cd /opt/servers
```

Откройте `docker-compose.yml` и при необходимости задайте свои переменные окружения:

```yaml
services:
  vpn-app:
    build: .
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      - DB_PATH=/data/app.db
      - ADMIN_USERNAME=мойлогин
      - ADMIN_PASSWORD=мойпароль
      - AUTH_TOKEN=очень_секретный_токен
    volumes:
      - db-data:/data
```

- `ADMIN_USERNAME` и `ADMIN_PASSWORD` — логин/пароль для входа в админку.
- `AUTH_TOKEN` — токен, которым backend защищает API (фронтенд автоматически подставляет его после успешного логина).

### 3. Сборка и запуск контейнера

```bash
cd /opt/server-servers
docker compose build
docker compose up -d
```

Проверка логов:

```bash
docker compose logs -f
```

Приложение будет работать на `http://127.0.0.1:8080` (доступ только с локального хоста, т.к. мы пробросили порт через `127.0.0.1`).

### 4. Настройка Nginx с доменом

Создайте конфиг Nginx, например `/etc/nginx/sites-available/server.conf`:

```nginx
server {
    listen 80;
    server_name vpn.example.com;

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

Активируйте сайт и перезагрузите Nginx:

```bash
ln -s /etc/nginx/sites-available/server.conf /etc/nginx/sites-enabled/server.conf
nginx -t
systemctl reload nginx
```

Теперь приложение доступно по адресу `http://example.com`.

### 5. Включение HTTPS (рекомендуется)

Используйте certbot (пример для Debian/Ubuntu):

```bash
apt install certbot python3-certbot-nginx
certbot --nginx -d example.com
```

Certbot сам обновит конфиг Nginx и добавит блоки для HTTPS.

### 6. Как работает авторизация

1. В форме логина вы вводите логин/пароль.
2. Фронтенд отправляет запрос `POST /api/login` в контейнер Go.
3. Backend сравнивает логин/пароль с `ADMIN_USERNAME` / `ADMIN_PASSWORD` (из `docker-compose.yml`).
4. При успехе backend возвращает токен (`AUTH_TOKEN` или значение по умолчанию).
5. Фронтенд сохраняет токен в `localStorage` и передаёт его во всех запросах API в заголовке:

   ```http
   Authorization: Bearer <AUTH_TOKEN>
   ```

6. Backend проверяет этот заголовок для всех защищённых эндпоинтов (`/api/servers`, `/api/smtp-settings`). Без токена или с неверным токеном будет `401 Unauthorized`.

Важно:
- На фронте **нет** захардкоженных логина/пароля — всё задаётся в Docker‑окружении.
- Если вы поменяли `ADMIN_USERNAME`, `ADMIN_PASSWORD` или `AUTH_TOKEN`, перезапустите контейнер:

```bash
docker compose up -d
```

### 7. Типичные проблемы и их решение

- **Ошибка авторизации / не пускает в систему**
  - Проверьте, что в `docker-compose.yml` заданы корректные `ADMIN_USERNAME`/`ADMIN_PASSWORD`.
  - После изменения значений перезапустите контейнер.

- **Ошибка `Error loading servers from API` в браузере**
  - Убедитесь, что вы вошли в систему (без токена backend вернёт `401`).
  - Проверьте логи контейнера: `docker compose logs -f`.

- **Ошибка при удалении сервера (`Unexpected end of JSON input`)**
  - В текущей версии клиент корректно обрабатывает ответы `204 No Content` от backend; если видите эту ошибку, убедитесь, что контейнер пересобран из актуального кода.

