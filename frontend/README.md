# SMLT frontend

Исходники публичного интерфейса `smlt.lol`. Это обычный проект Vue 3 + Vite.

## Структура

- `src/App.vue` — состояние приложения и переключение разделов.
- `src/components/LeaderboardView.vue` — рейтинг, поиск, фильтры и статистика.
- `src/components/RecentChanges.vue` — история перестановок игроков.
- `src/components/EventsView.vue` — видео и проекты SMLT.
- `src/components/LoginModal.vue` и `PlayerModal.vue` — администрирование.
- `src/api.js` — все обращения к Go API.
- `src/styles.css` — общие стили и мобильная адаптация.

## Разработка

```bash
corepack pnpm install
corepack pnpm dev
```

## Production-сборка

```bash
corepack pnpm run build
```

Vite записывает результат в `/opt/smlt/frontend-build`. Рабочую сборку Go-сервис
читает из `/opt/smlt/frontend-dist`.

## Доступ и выкладка

Разработчики работают только с этим репозиторием через отдельные ветки и Pull
Request. Исходники Go-бекенда, PostgreSQL, `.env`, SSH-ключи и VPS в репозиторий
не входят и через GitHub не выдаются. Правила для участников — в
[`CONTRIBUTING.md`](CONTRIBUTING.md).
