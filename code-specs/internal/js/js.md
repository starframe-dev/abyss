# Спецификация пакета `js`

## Обзор

Пакет `js` предоставляет выполнение JavaScript через фреймворк JavaScriptCore (JSC) на macOS. Он реализует связку cgo, которая позволяет Go-коду взаимодействовать с нативным API JavaScriptCore.

## Архитектура

### CGO и нативные зависимости

```go
#cgo LDFLAGS: -framework JavaScriptCore
#cgo CFLAGS: -x objective-c
```

Пакет использует фреймворк JavaScriptCore Apple для выполнения JavaScript-кода. Все нативные функции объявлены в блоке `cgo` и вызываются через C-структуры.

### Важные нативные функции

- `makeGoFunction` — создание функции Go внутри JS контекста (реализация в `bridge.c`)
- `_evaluate` — выполнение JavaScript-скрипта
- `_valueToString` — конвертация JSValueRef в строку (C-строку)
- `_setProperty` — установка свойства объекта
- `_getProperty` — получение свойства объекта
- `_makeNumber` — создание числового значения
- `_makeString` — создание строкового значения
- `_makeBoolean` — создание булевого значения
- `_makeUndefined` — создание undefined значения
- `_makeNull` — создание null значения
- `_getType` — получение типа значения
- `_toNumber` — конвертация значения в число
- `_toBoolean` — конвертация значения в булево

## Публичный API

### Тип `Context`

```go
type Context struct {
    ctx C.JSGlobalContextRef
}
```

`Context` представляет собой окружение выполнения JavaScript с глобальным контекстом.

#### Методы `Context`

##### `NewContext() *Context`

Создаёт новый контекст выполнения JavaScript.

**Возвращает:**
- `*Context` — новый контекст с инициализированным JSGlobalContextRef

##### `Close()`

Освобождает ресурсы контекста.

**Действие:**
- Если `ctx` не nil, вызывает `JSGlobalContextRelease`
- Устанавливает `ctx` в nil

##### `GetContextRef() C.JSGlobalContextRef`

Возвращает ссылку на JSGlobalContextRef.

**Возвращает:**
- `C.JSGlobalContextRef` — нативную ссылку

##### `Evaluate(script string) (string, error)`

Выполняет JavaScript-строку и возвращает результат как Go-строку.

**Параметры:**
- `script string` — JavaScript-код для выполнения

**Возвращает:**
- `string` — результат выполнения или пустая строка
- `error` — ошибка при выполнении JS или конвертации результата

**Поведение:**
1. Конвертирует Go-строку в C-строку (cgo)
2. Вызывает `_evaluate` для выполнения скрипта
3. Если есть исключение: конвертирует его в строку через `_valueToString` и возвращает ошибку
4. Если результат nil — возвращает пустую строку
5. Иначе конвертирует результат в строку и возвращает

**Важно:** Результат всегда конвертируется в строку. Для работы с типами используйте другие методы.

##### `EvaluateRaw(script string) (C.JSValueRef, error)`

Выполняет JavaScript и возвращает сырой `JSValueRef`.

**Параметры:**
- `script string` — JavaScript-код

**Возвращает:**
- `C.JSValueRef` — сырой результат выполнения
- `error` — при ошибке

**Важно:** Не используйте этот метод для получения результата, если не знаете тип. Всегда конвертируйте значение через `ValueToString` или другие методы типа.

##### `GlobalObject() C.JSObjectRef`

Возвращает глобальный объект контекста.

**Возвращает:**
- `C.JSObjectRef` — ссылку на глобальный объект

##### `SetProperty(name string, value C.JSValueRef) error`

Устанавливает свойство на глобальном объекте.

**Параметры:**
- `name string` — имя свойства
- `value C.JSValueRef` — значение свойства

**Возвращает:**
- `error` — nil при успехе, ошибку при неудаче

**Поведение:**
1. Конвертирует имя свойства в C-строку
2. Вызывает `_setProperty` с глобальным объектом
3. Обрабатывает исключения

##### `SetPropertyOn(obj C.JSObjectRef, name string, value C.JSValueRef) error`

Устанавливает свойство на конкретном объекте.

**Параметры:**
- `obj C.JSObjectRef` — целевой объект
- `name string` — имя свойства
- `value C.JSValueRef` — значение

**Возвращает:**
- `error` — nil при успехе, ошибку при неудаче

##### `GetProperty(name string) (C.JSValueRef, error)`

Получает свойство с глобального объекта.

**Параметры:**
- `name string` — имя свойства

**Возвращает:**
- `C.JSValueRef` — значение свойства
- `error` — nil при успехе, ошибку при неудаче

##### `ValueToString(value C.JSValueRef) (string, error)`

Конвертирует JSValueRef в Go-строку.

**Параметры:**
- `value C.JSValueRef` — значение для конвертации

**Возвращает:**
- `string` — строковое представление
- `error` — nil при успехе

##### `MakeString(value string) C.JSValueRef`

Создаёт строковое JS-значение.

**Параметры:**
- `value string` — строка для конвертации

**Возвращает:**
- `C.JSValueRef` — созданное значение

##### `MakeNumber(value float64) C.JSValueRef`

Создаёт числовое JS-значение.

**Параметры:**
- `value float64` — число для конвертации

**Возвращает:**
- `C.JSValueRef` — созданное значение

##### `MakeBoolean(value bool) C.JSValueRef`

Создаёт булевое JS-значение.

**Параметры:**
- `value bool` — булево значение

**Возвращает:**
- `C.JSValueRef` — созданное значение

##### `MakeUndefined() C.JSValueRef`

Создаёт undefined JS-значение.

**Возвращает:**
- `C.JSValueRef` — созданное значение

##### `MakeNull() C.JSValueRef`

Создаёт null JS-значение.

**Возвращает:**
- `C.JSValueRef` — созданное значение

##### `ToNumber(value C.JSValueRef) (float64, error)`

Конвертирует JS-значение в Go float64.

**Параметры:**
- `value C.JSValueRef` — значение для конвертации

**Возвращает:**
- `float64` — числовое значение
- `error` — ошибка при конвертации

##### `ToBoolean(value C.JSValueRef) bool`

Конвертирует JS-значение в Go bool.

**Параметры:**
- `value C.JSValueRef` — значение для конвертации

**Возвращает:**
- `bool` — булево значение

##### `GetType(value C.JSValueRef) string`

Возвращает тип JavaScript-значения.

**Параметры:**
- `value C.JSValueRef` — значение для проверки типа

**Возвращает:**
- `string` — один из: "undefined", "null", "boolean", "number", "string", "object", "unknown"

**Логика:**
Использует `kJSTypeUndefined`, `kJSTypeNull`, `kJSTypeBoolean`, `kJSTypeNumber`, `kJSTypeString`, `kJSTypeObject`

## Управление памятью

### CGO и освобождение памяти

Все строки, создаваемые через `C.CString`, освобождаются через `defer C.free`:

- Скрипты для `Evaluate` и `EvaluateRaw`
- Имена свойств для `SetProperty`, `SetPropertyOn`, `GetProperty`
- Значения для `MakeString`, `MakeNumber` (через C-строку для MakeString)

### Освобождение контекста

Метод `Close()` освобождает JSGlobalContextRef через `JSGlobalContextRelease`. Важно вызывать его при завершении работы с контекстом.

### Утечки памяти

Пакет не использует `go:linkname` или другие обходы Cgo. Все операции через стандартный cgo API.

## Обработка ошибок

### Формат ошибок

Все ошибки имеют префикс с указанием операции:
- `js error: ...` — общая ошибка JS
- `set property error: ...` — ошибка установки свойства
- `get property error: ...` — ошибка получения свойства
- `value to string error: ...` — ошибка конвертации значения
- `to number error: ...` — ошибка конвертации в число

### Источники ошибок

1. JS-исключения (синтаксические ошибки, runtime ошибки)
2. Ошибки конвертации значений
3. Ошибки работы с свойствами

## Внутренние детали

### CGO структуры

Все нативные типы из JavaScriptCore:
- `JSGlobalContextRef` — контекст выполнения
- `JSObjectRef` — объект JavaScript
- `JSValueRef` — значение JavaScript
- `C.JSValueRef` — тип для JSValueRef в Go

### CGO функции

Все внешние функции объявлены в cgo блоке. Реализация находится в `bridge.c` (не включена в этот файл).

### Импорт stdlib

- `fmt` — для форматирования ошибок
- `unsafe` — для работы с сырыми указателями памяти

## Типы значений

### JSValueRef

Нативный тип JS-значения. Все значения возвращаются или принимаются как `C.JSValueRef`.

### C-строки

Все строки из C-мира конвертируются через:
- `C.CString` для создания
- `C.free` для освобождения
- `C.GoString` для конвертации в Go

## Рекомендации по использованию

1. **Всегда вызывайте `Close()`** — освобождайте контекст при завершении работы
2. **Используйте `Evaluate` для выполнения кода** — он конвертирует результат в строку
3. **Используйте `MakeString`, `MakeNumber`, `MakeBoolean` для создания значений** — используйте `ValueToString` для чтения
4. **Проверяйте ошибки** — все операции возвращают ошибки при неудаче
5. **Не сохраняйте сырые `JSValueRef`** — всегда конвертируйте через методы типа

## Ограничения

1. **Работает только на macOS** — требует фреймворк JavaScriptCore Apple
2. **Результат `Evaluate` всегда строка** — для работы с типами используйте отдельные методы
3. **Все значения через `C.JSValueRef`** — это сырые нативные ссылки, требует аккуратной работы
4. **Не поддерживает передачу Go-функций в JS** — `makeGoFunction` реализован в `bridge.c` отдельно

## Примеры использования

### Выполнение кода

```go
ctx := js.NewContext()
defer ctx.Close()

result, err := ctx.Evaluate("2 + 2")
if err != nil {
    // обработка ошибки
}
fmt.Println(result) // "4"
```

### Работа с типами

```go
num, err := ctx.Evaluate("42")
str, err := ctx.Evaluate("'hello'")

numVal, _ := ctx.ToNumber(ctx.MakeNumber(3.14))
strVal, _ := ctx.MakeString("test")
boolVal, _ := ctx.MakeBoolean(true)
```

### Установка свойств

```go
ctx.SetProperty("myVar", ctx.MakeString("value"))
ctx.GetProperty("myVar")
```
