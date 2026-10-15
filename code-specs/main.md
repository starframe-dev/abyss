# Abyss — спецификация точки входа

## `main.go`

Программа разбирает флаги командной строки:

- `--headless-dump` — путь к файлу для сохранения ANSI-рендера. Если задан, включается headless-режим.
- `--url` — URL для загрузки в headless-режиме.
- `--width` — ширина viewport в headless-режиме, по умолчанию 200.

В headless-режиме при отсутствии `--url` программа пишет `--headless-dump requires --url` в stderr и завершает процесс с кодом 2. Если URL задан, вызывает `browser.HeadlessDump(url, width, path)` и завершается.

Без `--headless-dump` создаётся модель через `browser.New()`, запускается Bubble Tea с альтернативным экраном (`tea.WithAltScreen()`). Если `Run()` возвращает ошибку, она выводится в stderr с префиксом `Error: `, после чего процесс завершается с кодом 1.
