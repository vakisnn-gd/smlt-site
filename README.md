# SMLT site

Исходный код сайта [smlt.lol](https://smlt.lol): публичный интерфейс Vue и
Go-бекенд. Репозиторий приватный.

Это единственный канонический Git-репозиторий проекта и источник production-кода.
Рабочая копия на VPS: `/opt/smlt/site-source`; production запускается из
`/opt/smlt/backend/server`, а production-статические файлы находятся в
`/opt/smlt/frontend-dist/`. Каталоги runtime не являются Git-источником.
См. `/opt/smlt/AGENTS.md` на VPS для карты сервисов и безопасного процесса сборки.

```text
.
├── frontend/  Vue 3 + Vite: страницы, стили и компоненты
└── backend/   Go API: рейтинг, ивенты, авторизация и PostgreSQL
```

## Скачать проект

На странице репозитория: **Code → Download ZIP**. Для первого знакомства Git
не нужен. После распаковки фронтенд запускается так:

```bash
cd frontend
corepack pnpm install
corepack pnpm dev
```

## Правила безопасности

Файл `backend/.env` существует только на VPS и никогда не попадает в GitHub:
там пароль админки, секрет сессий и адрес базы данных. В репозитории оставлен
только безопасный шаблон [`backend/.env.example`](backend/.env.example).

Не добавляй в коммиты ключи, токены, пароли, IP-адреса сервера или готовые
сборки. Фронтенд и бекенд выкатываются на VPS только после просмотра изменений.

Подробности о работе с фронтендом — в
[`frontend/CONTRIBUTING.md`](frontend/CONTRIBUTING.md).
