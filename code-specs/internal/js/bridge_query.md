# Спецификация bridge_query.go

## Обзор

Файл `bridge_query.go` предоставляет реализацию функций-мостов между Go-бэкендом и JavaScriptCore (через JavaScript Core Bridge). Файл содержит C-биндинги для взаимодействия с JavaScriptCore и Go-обёртки для работы с DOM-элементами.

## Пакет

```go
package js
```

## Зависимости

### C-заголовки
- `#include <JavaScriptCore/JavaScriptCore.h>` — JavaScriptCore библиотека
- `#include <stdlib.h>` — Стандартные C-функции

### Экспортируемые C-функции (extern)
- `JSObjectRef makeGoFunction(JSContextRef ctx, const char* funcName)` — Создание функции Go в JS контексте
- `JSType _getType(JSContextRef ctx, JSValueRef value)` — Получение типа значения
- `char* _valueToString(JSContextRef ctx, JSValueRef value, JSValueRef* exception)` — Преобразование значения в строку
- `JSValueRef _makeArray(JSContextRef ctx, JSValueRef* values, size_t count, JSValueRef* exception)` — Создание массива
- `void _fprintf_stderr(const char* msg)` — Вывод в stderr
- `void _protectValue(JSContextRef ctx, JSValueRef value)` — Защита значения от Garbage Collector
- `void _unprotectValue(JSContextRef ctx, JSValueRef value)` — Удаление защиты значения
- `JSValueRef _callAsFunction(JSContextRef ctx, JSObjectRef func, JSObjectRef thisObj, size_t argCount, const JSValueRef args[], JSValueRef* exception)` — Вызов функции
- `JSValueRef _makeString(JSContextRef ctx, const char* value)` — Создание строкового значения
- `JSValueRef _makeBoolean(JSContextRef ctx, bool value)` — Создание булевого значения

## Типы

### Context
Тип `*Context` (выводится из `C.JSValueRef`) представляет собой JavaScriptCore контекст. Используется для создания JS-значений и вызова C-функций.

### C.JSValueRef
Указатель на `JSValueRef` — базовый тип значения JavaScriptCore.

### C.size_t
Тип размера в C, используется для количества элементов в массивах.

## Публичный API

### Функции

#### getBodyJS
```go
func getBodyJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Получение тела документа (эквивалент `document.body`).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS

**Возврат**: `C.JSValueRef` — JS-значение тела документа или `null` если не найдено.

**Поведение**:
- Если `getBody()` возвращает `nil`, возвращается `ctx.MakeNull()`
- В противном случае вызывается `nodeToJS(ctx, body)`

---

#### getElementByIdJS
```go
func getElementByIdJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Получение элемента по ID (эквивалент `document.getElementById(id)`).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS (первый — ID элемента)

**Возврат**: `C.JSValueRef` — JS-значение найденного элемента или `null` если не найдено/недостаточно аргументов.

**Поведение**:
1. Проверяет, что `len(args) >= 1`
2. Вызывает `getArgString(ctx, args[0])` для получения ID
3. Вызывает `getElementByID(id)` для поиска элемента
4. Если элемент `nil`, возвращает `ctx.MakeNull()`
5. Иначе возвращает результат `nodeToJS(ctx, node)`

---

#### querySelectorJS
```go
func querySelectorJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Поиск элемента по селектору (эквивалент `element.querySelector(sel)`).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS (элемент, селектор)

**Возврат**: `C.JSValueRef` — JS-значение найденного элемента или `null` если не найден/недостаточно аргументов/неверный ID.

**Поведение**:
1. Проверяет `len(args) >= 2`
2. Преобразует первый аргумент в NodeID через `jsToNodeID(ctx, args[0])`
3. Если ID < 0, возвращает `ctx.MakeNull()`
4. Преобразует второй аргумент в строку через `getArgString(ctx, args[1])`
5. Вызывает `querySelector(id, sel)`
6. Если элемент `nil`, возвращает `ctx.MakeNull()`
7. Иначе возвращает результат `nodeToJS(ctx, node)`

---

#### querySelectorAllJS
```go
func querySelectorAllJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Поиск всех элементов по селектору (эквивалент `element.querySelectorAll(sel)`).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS (элемент, селектор)

**Возврат**: `C.JSValueRef` — JS-массив найденных элементов или `undefined`/`null` при ошибке.

**Поведение**:
1. Проверяет `len(args) >= 2`
2. Преобразует первый аргумент в NodeID через `jsToNodeID(ctx, args[0])`
3. Если ID < 0, возвращает `ctx.MakeUndefined()`
4. Преобразует второй аргумент в строку через `getArgString(ctx, args[1])`
5. Вызывает `querySelectorAll(id, sel)`
6. Преобразует каждый элемент массива через `nodeToJS(ctx, n)`
7. Если массив пуст, возвращает `C._makeArray(ctx.ctx, nil, 0, &exc)`
8. Иначе возвращает созданный массив через `C._makeArray(...)`

---

#### getAttributeJS
```go
func getAttributeJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Получение атрибута элемента (эквивалент `element.getAttribute(name)`).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS (элемент, имя атрибута)

**Возврат**: `C.JSValueRef` — JS-строка значения атрибута или `null` если не найдено/недостаточно аргументов.

**Поведение**:
1. Проверяет `len(args) >= 2`
2. Преобразует первый аргумент в NodeID через `jsToNodeID(ctx, args[0])`
3. Если ID < 0, возвращает `ctx.MakeNull()`
4. Преобразует второй аргумент в строку через `getArgString(ctx, args[1])`
5. Вызывает `getAttribute(id, name)`
6. Возвращает `ctx.MakeString(val)`

---

#### getParentNodeJS
```go
func getParentNodeJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Получение родительского узла (эквивалент `element.parentNode` или `element.parentElement`).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS (элемент)

**Возврат**: `C.JSValueRef` — JS-значение родительского узла или `null` если не найден/недостаточно аргументов.

**Поведение**:
1. Проверяет `len(args) >= 1`
2. Преобразует первый аргумент в NodeID через `jsToNodeID(ctx, args[0])`
3. Если ID < 0, возвращает `ctx.MakeNull()`
4. Вызывает `getParentNode(id)`
5. Если родитель `nil`, возвращает `ctx.MakeNull()`
6. Иначе возвращает результат `nodeToJS(ctx, parent)`

---

#### getChildrenJS
```go
func getChildrenJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Получение дочерних элементов (эквивалент `element.children` или `element.childNodes` для элементов).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS (элемент)

**Возврат**: `C.JSValueRef` — JS-массив дочерних элементов или `undefined` при ошибке.

**Поведение**:
1. Проверяет `len(args) >= 1`
2. Преобразует первый аргумент в NodeID через `jsToNodeID(ctx, args[0])`
3. Если ID < 0, возвращает `ctx.MakeUndefined()`
4. Вызывает `getChildren(id)`
5. Преобразует каждый элемент массива через `nodeToJS(ctx, child)`
6. Если массив пуст, возвращает `C._makeArray(ctx.ctx, nil, 0, &exc)`
7. Иначе возвращает созданный массив через `C._makeArray(...)`

---

#### getElementsByTagNameJS
```go
func getElementsByTagNameJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Получение элементов по тегу (эквивалент `document.getElementsByTagName(tag)`).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS (тег)

**Возврат**: `C.JSValueRef` — JS-массив найденных элементов или `undefined` при ошибке.

**Поведение**:
1. Проверяет `len(args) >= 1`
2. Преобразует первый аргумент в строку через `getArgString(ctx, args[0])`
3. Вызывает `getElementsByTagName(tagName)`
4. Преобразует каждый элемент массива через `nodeToJS(ctx, n)`
5. Если массив пуст, возвращает `C._makeArray(ctx.ctx, nil, 0, &exc)`
6. Иначе возвращает созданный массив через `C._makeArray(...)`

---

#### getElementsByClassNameJS
```go
func getElementsByClassNameJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Получение элементов по классу (эквивалент `document.getElementsByClassName(className)`).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS (класс)

**Возврат**: `C.JSValueRef` — JS-массив найденных элементов или `undefined` при ошибке.

**Поведение**:
1. Проверяет `len(args) >= 1`
2. Преобразует первый аргумент в строку через `getArgString(ctx, args[0])`
3. Вызывает `getElementsByClassName(className)`
4. Преобразует каждый элемент массива через `nodeToJS(ctx, n)`
5. Если массив пуст, возвращает `C._makeArray(ctx.ctx, nil, 0, &exc)`
6. Иначе возвращает созданный массив через `C._makeArray(...)`

---

#### matchesJS
```go
func matchesJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Проверка соответствия элемента селектору (эквивалент `element.matches(selector)`).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS (элемент, селектор)

**Возврат**: `C.JSValueRef` — Булево значение `false` при ошибке или результате проверки.

**Поведение**:
1. Проверяет `len(args) >= 2`
2. Преобразует первый аргумент в NodeID через `jsToNodeID(ctx, args[0])`
3. Если ID < 0, возвращает `ctx.MakeBoolean(false)`
4. Преобразует второй аргумент в строку через `getArgString(ctx, args[1])`
5. Если ошибка преобразования, возвращает `ctx.MakeBoolean(false)`
6. Вызывает `matches(id, selector)`
7. Возвращает `ctx.MakeBoolean(matches(id, selector))`

---

#### closestJS
```go
func closestJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Поиск ближайшего предка по селектору (эквивалент `element.closest(selector)`).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS (элемент, селектор)

**Возврат**: `C.JSValueRef` — JS-значение найденного предка или `null` если не найден/недостаточно аргументов.

**Поведение**:
1. Проверяет `len(args) >= 2`
2. Преобразует первый аргумент в NodeID через `jsToNodeID(ctx, args[0])`
3. Если ID < 0, возвращает `ctx.MakeNull()`
4. Преобразует второй аргумент в строку через `getArgString(ctx, args[1])`
5. Если ошибка преобразования, возвращает `ctx.MakeNull()`
6. Вызывает `closest(id, selector)`
7. Если узел `nil`, возвращает `ctx.MakeNull()`
8. Иначе возвращает результат `nodeToJS(ctx, node)`

---

#### previousElementSiblingJS
```go
func previousElementSiblingJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Получение предыдущего sibling-элемента (эквивалент `element.previousElementSibling` или `element.previousSibling`).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS (элемент)

**Возврат**: `C.JSValueRef` — JS-значение предыдущего элемента или `null` если не найден/недостаточно аргументов.

**Поведение**:
1. Проверяет `len(args) >= 1`
2. Преобразует первый аргумент в NodeID через `jsToNodeID(ctx, args[0])`
3. Если ID < 0, возвращает `ctx.MakeNull()`
4. Вызывает `previousElementSibling(id)`
5. Если узел `nil`, возвращает `ctx.MakeNull()`
6. Иначе возвращает результат `nodeToJS(ctx, node)`

---

#### nextElementSiblingJS
```go
func nextElementSiblingJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Получение следующего sibling-элемента (эквивалент `element.nextElementSibling` или `element.nextSibling`).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS (элемент)

**Возврат**: `C.JSValueRef` — JS-значение следующего элемента или `null` если не найден/недостаточно аргументов.

**Поведение**:
1. Проверяет `len(args) >= 1`
2. Преобразует первый аргумент в NodeID через `jsToNodeID(ctx, args[0])`
3. Если ID < 0, возвращает `ctx.MakeNull()`
4. Вызывает `nextElementSibling(id)`
5. Если узел `nil`, возвращает `ctx.MakeNull()`
6. Иначе возвращает результат `nodeToJS(ctx, node)`

---

#### firstElementChildJS
```go
func firstElementChildJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Получение первого дочернего элемента (эквивалент `element.firstElementChild` или `element.firstChild`).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS (элемент)

**Возврат**: `C.JSValueRef` — JS-значение первого дочернего элемента или `null` если не найден/недостаточно аргументов.

**Поведение**:
1. Проверяет `len(args) >= 1`
2. Преобразует первый аргумент в NodeID через `jsToNodeID(ctx, args[0])`
3. Если ID < 0, возвращает `ctx.MakeNull()`
4. Вызывает `firstElementChild(id)`
5. Если узел `nil`, возвращает `ctx.MakeNull()`
6. Иначе возвращает результат `nodeToJS(ctx, node)`

---

#### lastElementChildJS
```go
func lastElementChildJS(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

**Назначение**: Получение последнего дочернего элемента (эквивалент `element.lastElementChild` или `element.lastChild`).

**Параметры**:
- `ctx *Context` — JavaScriptCore контекст
- `args []C.JSValueRef` — Массив аргументов JS (элемент)

**Возврат**: `C.JSValueRef` — JS-значение последнего дочернего элемента или `null` если не найден/недостаточно аргументов.

**Поведение**:
1. Проверяет `len(args) >= 1`
2. Преобразует первый аргумент в NodeID через `jsToNodeID(ctx, args[0])`
3. Если ID < 0, возвращает `ctx.MakeNull()`
4. Вызывает `lastElementChild(id)`
5. Если узел `nil`, возвращает `ctx.MakeNull()`
6. Иначе возвращает результат `nodeToJS(ctx, node)`

## Внутренние функции

### nodeToJS
```go
func nodeToJS(ctx *Context, node interface{}) C.JSValueRef
```

**Назначение**: Конвертирует Go-объект узла в JS-значение.

**Поведение**: Создаёт JS-значение для переданного узла.

### getArgString
```go
func getArgString(ctx *Context, arg C.JSValueRef) (string, error)
```

**Назначение**: Преобразует JS-значение в строку.

**Возврат**: Строка и ошибка (если есть).

### jsToNodeID
```go
func jsToNodeID(ctx *Context, arg C.JSValueRef) int
```

**Назначение**: Преобразует JS-значение в ID узла.

**Возврат**: ID узла (отрицательное значение указывает на ошибку).

### getBody
```go
func getBody() interface{}
```

**Назначение**: Получение тела документа.

**Возврат**: Объект тела документа или `nil`.

### getElementByID
```go
func getElementByID(id string) interface{}
```

**Назначение**: Поиск элемента по ID.

**Возврат**: Найденный элемент или `nil`.

### querySelector
```go
func querySelector(id int, selector string) interface{}
```

**Назначение**: Поиск элемента по селектору для конкретного узла.

**Возврат**: Найденный элемент или `nil`.

### querySelectorAll
```go
func querySelectorAll(id int, selector string) []interface{}
```

**Назначение**: Поиск всех элементов по селектору для конкретного узла.

**Возврат**: Срез найденных элементов.

### getAttribute
```go
func getAttribute(id int, name string) string
```

**Назначение**: Получение значения атрибута.

**Возврат**: Значение атрибута.

### getParentNode
```go
func getParentNode(id int) interface{}
```

**Назначение**: Получение родительского узла.

**Возврат**: Родительский узел или `nil`.

### getChildren
```go
func getChildren(id int) []interface{}
```

**Назначение**: Получение дочерних элементов.

**Возврат**: Срез дочерних элементов.

### getElementsByTagName
```go
func getElementsByTagName(tagName string) []interface{}
```

**Назначение**: Получение элементов по тегу.

**Возврат**: Срез найденных элементов.

### getElementsByClassName
```go
func getElementsByClassName(className string) []interface{}
```

**Назначение**: Получение элементов по классу.

**Возврат**: Срез найденных элементов.

### matches
```go
func matches(id int, selector string) bool
```

**Назначение**: Проверка соответствия элемента селектору.

**Возврат**: `true` если соответствует, `false` в противном случае.

### closest
```go
func closest(id int, selector string) interface{}
```

**Назначение**: Поиск ближайшего предка по селектору.

**Возврат**: Найденный предок или `nil`.

### previousElementSibling
```go
func previousElementSibling(id int) interface{}
```

**Назначение**: Получение предыдущего sibling-элемента.

**Возврат**: Предыдущий элемент или `nil`.

### nextElementSibling
```go
func nextElementSibling(id int) interface{}
```

**Назначение**: Получение следующего sibling-элемента.

**Возврат**: Следующий элемент или `nil`.

### firstElementChild
```go
func firstElementChild(id int) interface{}
```

**Назначение**: Получение первого дочернего элемента.

**Возврат**: Первый дочерний элемент или `nil`.

### lastElementChild
```go
func lastElementChild(id int) interface{}
```

**Назначение**: Получение последнего дочернего элемента.

**Возврат**: Последний дочерний элемент или `nil`.

## Особенности реализации

### Обработка ошибок
- Все функции возвращают `null`, `undefined` или `false` при ошибках вместо паники
- Ошибки преобразования аргументов игнорируются, функция возвращает безопасное значение
- Проверка количества аргументов выполняется в начале каждой функции

### Обработка пустых результатов
- Для функций, возвращающих массивы (querySelectorAll, getChildren, getElementsByTagName, getElementsByClassName) используется `C._makeArray` с пустым срезом
- Для функций, возвращающих одиночный узел (getElementById, getBody и др.) используется `ctx.MakeNull()`

### Типы возвращаемых значений
- `null` — когда узел не найден или ошибка преобразования
- `undefined` — для методов, возвращающих коллекцию (querySelectorAll, getChildren и др.)
- `false` — для булевых методов (matches)
- Массив — для методов, возвращающих коллекцию элементов

### Конвертация узлов в JS
Все функции используют внутреннюю функцию `nodeToJS` для конвертации Go-объекта узла в JS-значение.

### Создание массивов
Массивы создаются через C-функцию `C._makeArray` с параметрами:
- `ctx.ctx` — контекст
- `&childValues[0]` — указатель на первый элемент массива
- `C.size_t(len(childValues))` — количество элементов
- `&exc` — указатель на переменную для хранения исключения

## Границы

### Входные данные
- `args` — слайс `C.JSValueRef` указателей на JS-значения
- Аргументы должны быть валидными JS-значениями
- Первый аргумент для методов, принимающих узел, должен быть валидным JS-объектом

### Выходные данные
- Возвращаемые значения — валидные JS-значения для JavaScriptCore
- При ошибке возвращаются безопасные значения (null, undefined, false)

### Runtime assumptions
- JavaScriptCore библиотека должна быть инициализирована
- Контекст `ctx` должен быть валидным и инициализированным
- Экспортные C-функции должны быть доступны через `import "C"`

## Безопасность

### Защита от NULL-указателей
- Все функции проверяют количество аргументов перед обращением к ним
- Проверка `id < 0` защищает от обращения по невалидным ID-у

### Защита от ошибок преобразования
- Ошибки `getArgString` игнорируются, возвращается безопасное значение
- Ошибки преобразования NodeID возвращают null/undefined/false

### GC-защита
- Для значений используется `_protectValue`/`_unprotectValue` (декларировано в C, не используется напрямую в Go)
- Массивы создаются через защищённые C-функции
