# Начало работы

Abyss — TUI-браузер. Он выводит поддерживаемое подмножество HTML и CSS в терминал как ANSI-арт.

## Запуск

```bash
go run .
```

Или сборка и запуск:

```bash
go build -o abyss .
./abyss
```

Неинтерактивный дамп ANSI:

```bash
go run . --headless-dump /tmp/page.ansi --url https://example.com --width 120
```

## Управление

| Клавиша | Действие |
|---------|----------|
| `Ctrl+L` | Фокус на адресной строке |
| `Enter` | Загрузить URL |
| `↑/↓` `PgUp/PgDn` | Прокрутка страницы |
| `q` / `Ctrl+C` | Выход |
