# Исправление компиляции DOM API Abyss

## Контекст

Graphloom не может построить UCM для Abyss. Пакет `internal/js` не компилируется из-за файла `dom_api.go`: он содержит неполную копию DOM API без импортов и с дубликатами `closest` и `InitDOM`. Канонические реализации уже находятся в `dom.go`, `dom_query.go`, `dom_insert.go` и `dom_events.go`.

## Цель

Восстановить компиляцию `github.com/starframe-dev/abyss/internal/js`, сохранив экспортируемый API, используемый integration tests, и не дублируя каноническую DOM-логику.

## Что изменится

1. `internal/js/dom_api.go` — заменить сломанные реализации тонкими экспортируемыми обёртками над каноническими внутренними функциями; удалить дублирующие `closest` и `InitDOM`.
2. `internal/js/dom_api_test.go` или существующие integration tests — добавить/уточнить проверки только если текущего покрытия недостаточно.
3. `code-specs/graphloom-dom-api-compile-fix.md` — зафиксировать результаты.

## Детали реализации

1. Добавить необходимые импорты `JavaScriptCore`, `strings` и `internal/dom` только там, где они действительно нужны.
2. Экспортируемые методы `CreateElement`, `CreateTextNode`, `AppendChild`, `QuerySelector`, `Closest` и остальные API-методы должны делегировать каноническим функциям.
3. Для событий использовать `addEventListener`/`removeEventListener` из `dom_events.go`, а не отдельную карту listeners.
4. Не изменять поведение канонических DOM, query, insert, event и bridge-файлов.
5. Проверить `gofmt`, `go build ./...`, `go test ./...`, `go vet ./...` и повторный запуск Graphloom Go analyzer на Abyss.

## Критерии приёмки

- [ ] `internal/js` компилируется без duplicate redeclared и undefined ошибок.
- [ ] Экспортируемые DOM API integration tests проходят.
- [ ] `go build ./...`, `go test ./...` и `go vet ./...` проходят.
- [ ] Graphloom analyzer строит UCM для Abyss без ошибки пакета `internal/js`.
- [ ] Коммиты не создаются без отдельного разрешения.
