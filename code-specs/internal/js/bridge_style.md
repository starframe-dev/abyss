# bridge_style

## Назначение

Файл `internal/js/bridge_style.go` содержит три Go-обёртки для вызовов из JavaScript, работающие с типами JavaScriptCore через CGO.

## Функции

### `getComputedStyleJS(ctx *Context, args []C.JSValueRef) C.JSValueRef`

Если `args` пуст, возвращает `ctx.MakeNull()`. Иначе преобразует первый аргумент в ID через `jsToNodeID`. При отрицательном ID возвращает `null`; при неотрицательном вызывает `getComputedStyle(id)` и возвращает результат. Остальные аргументы игнорируются.

### `getBoundingClientRectJS(ctx *Context, args []C.JSValueRef) C.JSValueRef`

Если `args` пуст, возвращает `null`. Иначе преобразует первый аргумент через `jsToNodeID`; при отрицательном ID возвращает `null`. Для действительного ID вызывает `getBoundingClientRect(id)` и возвращает JavaScript-объект с числовыми полями `x`, `y`, `width`, `height`, `top`, `left`, `bottom`, `right`. Значения: `x = rect.X`, `y = rect.Y`, `width = rect.Width`, `height = rect.Height`, `top = rect.Y`, `left = rect.X`, `bottom = rect.Y + rect.Height`, `right = rect.X + rect.Width`. Остальные аргументы игнорируются.

### `logJS(ctx *Context, args []C.JSValueRef) C.JSValueRef`

Преобразует каждый аргумент через `getArgString`. При ошибке преобразования добавляет `<error>`, иначе добавляет полученную строку. Соединяет аргументы пробелами, добавляет префикс `[JS] ` и завершающий перевод строки. Перед записью в stderr заменяет на `�` управляющие Unicode-символы, кроме перевода строки и табуляции; символ DEL (`0x7f`) также заменяет. Записывает строку через `_fprintf_stderr` и возвращает `ctx.MakeUndefined()`. При пустом списке аргументов записывает `[JS] \n`.

## Используемые зависимости

Функции используют `Context`, `jsToNodeID`, `getComputedStyle`, `getBoundingClientRect` и `getArgString`, определённые в других частях пакета. В C-блоке объявлены `makeGoFunction`, `_getType`, `_valueToString`, `_makeArray`, `_fprintf_stderr`, `_protectValue`, `_unprotectValue`, `_callAsFunction`, `_makeString` и `_makeBoolean`; данный Go-файл непосредственно вызывает только `_fprintf_stderr` и использует JavaScriptCore API для создания объекта, строк, чисел и освобождения ресурсов.

Для временных C-строк имён свойств и сообщения вызывается `C.free` через `defer`; для созданных JS-строк имён свойств вызывается `C.JSStringRelease` через `defer`.
