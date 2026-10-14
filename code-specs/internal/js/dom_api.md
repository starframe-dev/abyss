# Спецификация: `internal/js/dom_api.go`

## Описание

Публичный Go API поверх неэкспортированных DOM-функций пакета `js`. Нужен интеграционным тестам: создание узлов, вставка, селекторы, атрибуты. Не дублирует `InitDOM` и `closest` — только обёртки.

## Экспорт

| Функция | Делегирует |
|---------|------------|
| `CreateElement` / `CreateTextNode` / `CreateDocumentFragment` | `createElement` / `createTextNode` / `createDocumentFragment` |
| `GetNodeID` / `GetNodeByID` | карта `nodes` |
| `AppendChild` / `InsertBefore` / `ReplaceChild` / `RemoveChild` | `dom_insert.go` / `dom.go` |
| `SetInnerHTML` / `GetInnerHTML` / `SetTextContent` / `GetTextContent` | serialize + insert |
| `QuerySelector` / `QuerySelectorAll` / `Matches` / `Closest` | `dom_query.go` |
| `CloneNode` / `Contains` / `FirstElementChild` / `LastElementChild` | `dom_query.go` |

Слушатели событий экспортируются из `dom_events.go` (`AddEventListener` / `RemoveEventListener`), не из этого файла.

## Правила

- Без cgo в этом файле.
- Без повторного объявления `InitDOM` и `closest`.
- Типы узлов — `*dom.Node`.
