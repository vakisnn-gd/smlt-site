# SMLT frontend

Публичный интерфейс [smlt.lol](https://smlt.lol), собранный на Vue 3, Vue Router и Vite. Приложение поддерживает русский и английский языки, маршруты профилей и PWA.

## Разделы сайта

- `/` и `/en` — рейтинг: поиск, фильтр по стране, сортировка по очкам, пройденным уровням и позиции Demonlist, статистика и профили.
- `/events` — видео событий и коллабораций; управление доступно авторизованному администратору.
- `/chats` — переход в отдельное приложение чата.
- `/tools` — сравнение игроков, случайный выбор, тема и стиль интерфейса, установка PWA.
- `/about` — описание сообщества, контакты и вход администратора.
- `/player/:id` — профиль как маршрут-модалка поверх лидерборда; при открытии и закрытии сохраняется позиция прокрутки.

Для английских страниц используется префикс `/en`. Nginx или backend должны отдавать SPA-оболочку для допустимых маршрутов.

## Структура исходников

- `src/App.vue` — загрузка данных и состояние приложения, язык, тема, авторизация и выбор раздела.
- `src/router.js` — маршруты, языковые варианты и сохранение прокрутки профиля.
- `src/api.js` — HTTP-клиент Go API.
- `src/components/LeaderboardView.vue` и `PlayerDetailsView.vue` — рейтинг и профиль.
- `src/components/RecentChanges.vue` — лента изменений и дозагрузка записей.
- `src/components/EventsView.vue` — события, YouTube-видео и формы администрирования.
- `src/components/ChatsView.vue`, `ToolsView.vue`, `AboutView.vue` — переход к чату, инструменты и информация о сообществе.
- `src/components/LoginModal.vue` и `PlayerModal.vue` — вход и профиль-модалка.
- `src/data.js` — локальные метаданные стран; `src/useDialog.js` — общая логика диалогов.
- `src/styles.css` — общие стили и адаптивная верстка.
- `public/` — PWA manifest, service worker, favicon, иконки, флаги и аватар.

## Разработка и сборка

Требуемые версии Node.js и pnpm зафиксированы через `package.json` и `pnpm-lock.yaml`.

```bash
corepack pnpm install --frozen-lockfile
corepack pnpm dev
corepack pnpm run build
```

Сборка записывается в `../frontend-build/`; это локальный промежуточный артефакт, а не production-каталог. Не копируйте его на сервер автоматически. API запросы обслуживает Go backend; при локальной работе используйте только тестовые данные и безопасную конфигурацию.

## Проверки

Перед Pull Request соберите production frontend командой `corepack pnpm run build` и просмотрите diff. Отдельные автоматизированные UI-тесты в этом каталоге не настроены. Из корня репозитория Go проверки запускаются командами `go test ./backend/...` и `go vet ./backend/...`; корневой `go.work` подключает модуль `backend/`.

## Границы проекта

Это не весь runtime сайта: Go API хранится рядом в `../backend/`, а SMLT Max и Chat — в `../messenger/`. Backend, PostgreSQL, production `.env`, сервисные файлы и nginx конфигурация не являются frontend исходниками. Не меняйте production или базу из локального dev-сервера.

См. [`CONTRIBUTING.md`](CONTRIBUTING.md) для порядка изменений.
