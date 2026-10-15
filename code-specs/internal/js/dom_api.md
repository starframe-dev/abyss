# Спецификация: `internal/js/dom_api.go`

## Описание

Публичный Go API поверх неэкспортированных DOM-функций пакета `js`. Предназначен для интеграционных тестов и предоставляет тонкие обёртки для создания узлов, доступа к реестру узлов, изменения дерева и содержимого, атрибутов и поиска элементов.

## Экспорт

| Функции | Поведение |
|---------|-----------|
| `CreateElement` / `CreateTextNode` / `CreateDocumentFragment` | Делегируют вызовы `createElement` / `createTextNode` / `createDocumentFragment`. |
| `GetNodeID` | Делегирует `getNodeID`. |
| `GetNodeByID` | Под блокировкой `nodeMu` ищет узел в карте `nodes` и возвращает узел и признак наличия. |
| `AppendChild` / `InsertBefore` / `ReplaceChild` / `RemoveChild` | Делегируют одноимённым операциям изменения дерева в нижнем регистре. |
| `SetInnerHTML` / `GetInnerHTML` | Делегируют `setInnerHTML` / `getInnerHTML`. |
| `SetAttribute` / `GetAttribute` | Делегируют `setAttribute` / `getAttribute`. |
| `SetTextContent` / `GetTextContent` | Делегируют `setTextContent` / `getTextContent`. |
| `GetBody` | Делегирует `getBody` и возвращает элемент body. |
| `GetElementByID` | Делегирует `getElementByID`, выполняя поиск по атрибуту id. |
| `QuerySelector` / `QuerySelectorAll` / `Matches` / `Closest` | Делегируют `querySelector` / `querySelectorAll` / `matches` / `closest`. |
| `CloneNode` / `Contains` / `FirstElementChild` / `LastElementChild` | Делегируют одноимённым функциям в нижнем регистре. |

Слушатели событий экспортируются из `dom_events.go` (`AddEventListener` / `RemoveEventListener`), а не из этого файла.

## Правила

- В этом файле нет cgo.
- В этом файле не объявляются `InitDOM` и функция `closest` в нижнем регистре; `Closest` — публичная обёртка над `closest`.
- Используемый тип узлов — `*dom.Node`.
