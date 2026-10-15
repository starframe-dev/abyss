# Спецификация пакета `js`

## Обзор

Пакет `js` предоставляет выполнение JavaScript через JavaScriptCore (JSC) на macOS, используя cgo и нативное API JavaScriptCore.

## Нативные зависимости и функции

Директивы cgo связывают JavaScriptCore framework и компилируют блок C как Objective-C:

```go
#cgo LDFLAGS: -framework JavaScriptCore
#cgo CFLAGS: -x objective-c
```

Объявлены функции-мосты `_evaluate`, `_valueToString`, `_setProperty`, `_getProperty`, `_makeNumber`, `_makeString`, `_makeBoolean`, `_makeUndefined`, `_makeNull`, `_getType`, `_toNumber`, `_toBoolean` и `makeGoFunction` (реализация моста находится вне этого файла).

## Публичный API

### `Context`

Хранит `C.JSGlobalContextRef`.

- `NewContext() *Context` создаёт глобальный контекст вызовом `JSGlobalContextCreate(nil)`.
- `Close()` при ненулевом контексте вызывает `resetEventSystemForContext`, освобождает его через `JSGlobalContextRelease` и записывает nil. Повторное закрытие не выполняет действий.
- `GetContextRef() C.JSGlobalContextRef` возвращает сохранённую нативную ссылку.
- `GlobalObject() C.JSObjectRef` возвращает глобальный объект контекста.

### Выполнение JavaScript

`Evaluate(script string) (string, error)` создаёт C-строку скрипта и URL `script.js`, выполняет скрипт, затем освобождает входные строки. При исключении возвращает ошибку `js error: ...`, преобразуя исключение в строку; если преобразование невозможно — `unknown js error`. При nil-результате возвращает пустую строку без ошибки. Иначе преобразует результат в строку; если преобразование результата не удалось, возвращает пустую строку без ошибки.

`EvaluateRaw(script string) (C.JSValueRef, error)` выполняет скрипт с тем же URL и обработкой исключений, но возвращает сырой результат. При исключении возвращает nil и ошибку (`js error: ...` либо `unknown js error`).

### Свойства

- `SetProperty(name string, value C.JSValueRef) error` устанавливает свойство глобального объекта.
- `SetPropertyOn(obj C.JSObjectRef, name string, value C.JSValueRef) error` устанавливает свойство указанного объекта.
- `GetProperty(name string) (C.JSValueRef, error)` получает свойство глобального объекта.

Методы создают и освобождают C-строку имени и передают исключения нативных операций. Ошибки установки и получения имеют префиксы `set property error: ` и `get property error: `; если исключение нельзя преобразовать в строку, возвращаются соответственно `unknown set property error` и `unknown get property error`.

### Преобразование и создание значений

- `ValueToString(value C.JSValueRef) (string, error)` преобразует значение в Go-строку. При ненулевом результате преобразования возвращает строку без ошибки; при исключении, если значение исключения преобразуемо, возвращает `value to string error: ...`; иначе возвращает пустую строку и nil.
- `MakeString(value string) C.JSValueRef` создаёт JS-строку; временная C-строка освобождается.
- `MakeNumber(value float64) C.JSValueRef` создаёт JS-число.
- `MakeBoolean(value bool) C.JSValueRef` создаёт JS-boolean.
- `MakeUndefined() C.JSValueRef` создаёт undefined.
- `MakeNull() C.JSValueRef` создаёт null.
- `ToNumber(value C.JSValueRef) (float64, error)` преобразует значение в число; ошибки имеют префикс `to number error: `, либо текст `unknown to number error`, если исключение не преобразуемо.
- `ToBoolean(value C.JSValueRef) bool` преобразует значение в boolean без возвращаемой ошибки.
- `GetType(value C.JSValueRef) string` сопоставляет типы JavaScript со строками `undefined`, `null`, `boolean`, `number`, `string`, `object`; прочие значения обозначаются `unknown`.

## Память и ошибки

Временные C-строки, созданные для скриптов, URL и имён свойств, освобождаются через `C.free`; C-строки результатов `_valueToString` также освобождаются. `Close` освобождает контекст JavaScriptCore и предварительно вызывает `resetEventSystemForContext`.

Операции, которые могут завершиться исключением JavaScriptCore, возвращают ошибки согласно описанным выше префиксам. `ToBoolean` и создание значений не возвращают ошибок.

## Ограничения

- Код требует JavaScriptCore framework и macOS.
- `Evaluate` всегда возвращает строковое представление результата; для получения сырого значения предназначен `EvaluateRaw`.
- В этом файле объявлен `makeGoFunction`, но публичного метода регистрации Go-функций нет.
