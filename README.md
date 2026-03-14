## Server Management System by Annonyx

Полнофункциональное веб‑приложение для учёта серверов, их стоимости и предстоящих платежей. Состоит из frontend‑части на React и backend‑сервера на Go с базой данных SQLite.

### Возможности

- **Управление серверами**: добавление, редактирование, удаление.
- **Финансовый учёт**: месячная и годовая стоимость, разные периоды оплаты.
- **Предстоящие платежи**: список оплат на ближайшие 30 дней.
- **SMTP‑настройки**: сохранение параметров почтового сервера для уведомлений (логика отправки писем может быть добавлена отдельно).
- **Авторизация**: простой вход по логину/паролю (для админ‑панели).

### Стек технологий

- Backend: **Go**, **SQLite**, чистый `net/http`.
- Frontend: **React 18**, **TypeScript**, **Tailwind CSS v4**, **Recharts**, UI‑компоненты на основе shadcn/ui.
- Сборка: **Vite**.
- Контейнеризация: **Docker**, **docker compose**.

### Структура проекта

- `backend/` — исходный код Go‑сервера
  - `cmd/server` — точка входа (`main.go`)
  - `internal/models` — доменные структуры (`VPNServer`, `SMTPSettings`)
  - `internal/storage` — слой работы с SQLite (`Store`, CRUD, тесты)
  - `internal/server` — HTTP‑обработчики и маршрутизация
- `front/` — React‑приложение (Vite)
- `Dockerfile` — multi‑stage сборка фронта и backend‑бинарника
- `docker-compose.yml` — запуск всего приложения одним сервисом

### Быстрый старт (Docker)

```bash
docker compose build
docker compose up
```

После сборки приложение будет доступно по адресу `http://localhost:8080`.

### Локальный запуск без Docker

Backend:

```bash
cd backend
go test ./...
go run ./cmd/server
```

Frontend:

```bash
cd front
npm install
npm run dev
```

### Лицензия

Проект распространяется по лицензии **MIT** с указанием правообладателя **Annonyx**. Полный текст — в файле `LICENSE`.

