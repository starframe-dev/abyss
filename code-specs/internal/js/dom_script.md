# Спецификация: internal/js/dom_script.go

## Обзор

Файл содержит функции для создания JS-обёртки для DOM-узла, выполнения встроенных и внешних JavaScript-скриптов через JavaScriptCore и получения финального DOM.

## `nodeToJS(ctx *Context, n *dom.Node) C.JSValueRef`

Получает идентификатор узла через `getNodeID`. Если он меньше нуля, возвращает JS `null`. Иначе создаёт объект JavaScriptCore и устанавливает `__go_id` (числовой идентификатор) и `nodeType`: для `ElementNode` — 1, `TextNode` — 3, `CommentNode` — 8, `DocumentNode` — 9, `DocumentFragmentNode` — 11. Для элемента устанавливает `tagName` в верхнем регистре; для текстового узла — `textContent` из поля `Text`. Возвращает объект, преобразованный через `jsObjToVal`.

## `GetFinalDOM() *dom.Node`

Возвращает глобальную переменную `rootDoc`.

## `ExecuteScripts(ctx *Context, doc *dom.Node, baseURL string) error`

Вызывает `ExecuteScriptsContext` с `context.Background()`. Поэтому для этой обёртки доступны общий лимит времени и ограничения на количество и объём скриптов, описанные ниже.

## `ExecuteScriptsContext(execCtx context.Context, ctx *Context, doc *dom.Node, baseURL string) error`

Создаёт общий контекст выполнения с таймаутом 30 секунд. Получает узлы через `doc.QueryAll("script")`; рассматривает не более первых 128. Перед обработкой каждого узла проверяет отмену контекста и возвращает ошибку контекста при отмене. Общий бюджет исходного кода — 8 МиБ. Если бюджет исчерпан, обработка прекращается. Внешний скрипт запрашивается через `fetchScriptContext`; при ошибке запроса он пропускается. Если размер полученного кода превышает остаток бюджета, обработка прекращается; иначе размер вычитается и код передаётся в `ctx.Evaluate`. Ошибки выполнения игнорируются. Для встроенного скрипта берётся `TextContent`; пустой после `strings.TrimSpace` текст пропускается. Превышение остатка бюджета прекращает обработку, иначе размер вычитается, код выполняется, а ошибка выполнения игнорируется. При штатном завершении возвращает `nil`.

## `fetchScript(baseURL, src string) (string, error)`

Создаёт контекст с таймаутом 3 секунды и вызывает `fetchScriptContext`. Самостоятельно URL не разбирает и HTTP-запрос не выполняет.

## `fetchScriptContext(parent context.Context, baseURL, src string) (string, error)`

Разбирает базовый URL и `src` через `url.Parse`, возвращая ошибки с пояснением при неудаче. Разрешает `src` относительно базы через `ResolveReference`. Создаёт дочерний контекст с таймаутом 3 секунды (ограниченным также родительским контекстом) и вызывает `fetcher.FetchJS` с разрешённым URL. Реализация HTTP-запроса, заголовков, чтения ответа и закрытия тела находится вне этого файла, в `fetcher`.

## Зависимости и детали реализации

Файл использует CGO и заголовок JavaScriptCore, функции `C.CString`/`C.free`, а также `C.JSStringCreateWithUTF8CString`, `C.JSStringRelease`, `C.JSObjectMake` и `C.JSObjectSetProperty`. В Go-коде используются `context`, `fmt`, `net/url`, `strings`, `time`, `unsafe`, а также пакеты `internal/dom` и `internal/fetcher`. Тексты комментариев в коде не меняют поведение.
