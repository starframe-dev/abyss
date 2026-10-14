# bridge_style

## Общее описание

Файл `bridge_style.go` предоставляет JavaScript-обёртки для нативных функций Go, которые работают с WebKit Runtime. Этот код является частью мостового слоя, позволяющего JavaScript вызовы Go-функции.

## Публичный API

### Тип Context

`Context` — тип контекста WebKit Runtime (не экспортируется из пакета).

### Функции JS-моста

#### `getComputedStyleJS(ctx *Context, args []C.JSValueRef) C.JSValueRef`

Получает вычисленные стили для элемента.

**Параметры:**
- `ctx` — контекст Runtime
- `args` — массив JS значений, где первый аргумент — ID элемента

**Поведение:**
1. Если передано меньше 1 аргумента, возвращается `null`
2. Конвертирует первый аргумент в ID узла через `jsToNodeID`
3. Если ID < 0, возвращается `null`
4. Вызывает `getComputedStyle(id)` и возвращает результат

**Возвращаемое значение:**
- `C.JSValueRef` — объект вычисленных стилей или `null`

---

#### `getBoundingClientRectJS(ctx *Context, args []C.JSValueRef) C.JSValueRef`

Получает границы элемента в клиентских координатах.

**Параметры:**
- `ctx` — контекст Runtime
- `args` — массив JS значений, где первый аргумент — ID элемента

**Поведение:**
1. Если передано меньше 1 аргумента, возвращается `null`
2. Конвертирует первый аргумент в ID узла через `jsToNodeID`
3. Если ID < 0, возвращается `null`
4. Вызывает `getBoundingClientRect(id)`
5. Создаёт JS-объект с полями: x, y, width, height, top, left, bottom, right
6. Заполняет объект значениями

**Возвращаемое значение:**
- `C.JSValueRef` — объект прямоугольника или `null`

**Поля возвращаемого объекта:**
| Поле | Тип | Описание |
|------|-----|----------|
| x | float64 | Левый край элемента |
| y | float64 | Верхний край элемента |
| width | float64 | Ширина элемента |
| height | float64 | Высота элемента |
| top | float64 | Верхний край (равен y) |
| left | float64 | Левый край (равен x) |
| bottom | float64 | Нижний край (y + height) |
| right | float64 | Правый край (x + width) |

---

#### `logJS(ctx *Context, args []C.JSValueRef) C.JSValueRef`

Выводит аргументы в stderr через `[JS]` префикс.

**Параметры:**
- `ctx` — контекст Runtime
- `args` — массив JS значений для вывода

**Поведение:**
1. Конвертирует каждый аргумент в строку через `getArgString`
2. При ошибке конвертации использует `<error>`
3. Объединяет все части с префиксом `[JS] ` и суффиксом `\n`
4. Вызывает C-функцию `_fprintf_stderr`
5. Возвращает `undefined`

**Возвращаемое значение:**
- `C.JSValueRef` — `undefined`

---

## Типы и структуры

### C.JSValueRef

Неэкспортированный тип ссылки на JS значение WebKit.

### C.JSObjectRef

Неэкспортированный тип ссылки на JS объект WebKit.

### C.JSContextRef

Неэкспортированный тип ссылки на JS контекст WebKit.

---

## Внутренние функции

### `jsToNodeID(ctx *Context, arg C.JSValueRef) int`

Конвертирует JS значение в ID узла.

**Поведение:**
- Принимает JS значение
- Возвращает int ID узла
- Возвращает отрицательное число при ошибке

### `getComputedStyle(id int) C.JSValueRef`

Получает вычисленные стили для узла.

### `getBoundingClientRect(id int) struct{}`

Получает границы узла. Возвращает структуру с полями:
- `X` float64
- `Y` float64
- `Width` float64
- `Height` float64

### `getArgString(ctx *Context, arg C.JSValueRef) (string, error)`

Конвертирует JS значение в строку.

---

## Связанные C-функции

### `makeGoFunction(ctx, funcName)`

Создаёт функцию Go для использования из JavaScript.

### `_getType(ctx, value)`

Получает тип JS значения.

### `_valueToString(ctx, value, exception)`

Конвертирует JS значение в строку, устанавливая ошибку при провале.

### `_makeArray(ctx, values, count, exception)`

Создаёт JS массив из значений.

### `_fprintf_stderr(msg)`

Выводит сообщение в stderr.

### `_protectValue(ctx, value)`

Защищает JS значение от Garbage Collector.

### `_unprotectValue(ctx, value)`

Убирает защиту JS значения от Garbage Collector.

### `_callAsFunction(ctx, func, thisObj, argCount, args, exception)`

Вызывает функцию как метод.

### `_makeString(ctx, value)`

Создаёт строку из C-строки.

### `_makeBoolean(ctx, value)`

Создаёт логическое значение.

---

## Правила работы с ошибками

1. Функции возвращают `null` или `undefined` вместо паники
2. Ошибки конвертации обрабатываются внутренне
3. C-строки освобождается через `C.free` с `defer`
4. JS-строки освобождаются через `C.JSStringRelease`

---

## Важные примечания

1. Код использует CGO для вызова WebKit C-API
2. Все JS-значения должны быть правильно освобождены
3. ID узла должен быть неотрицательным
4. Меньшее количество аргументов, чем ожидается, возвращает `null`
5. Log-функция не обрабатывает ошибки конвертации, заменяя их `<error>`
