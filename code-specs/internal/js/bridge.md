# Specification: internal/js/bridge.go

## Обзор

Файл `bridge.go` реализует мост (bridge) между Go-кодом и JavaScriptCore через C-интерфейс. Он позволяет вызывать Go-функции из JavaScript в контекте WebKit/JavaScriptCore.

## Цель

Предоставить механизм вызова Go-кода из JavaScript через JavaScriptCore API, регистрируя обработчики DOM-подобных функций и обеспечивая безопасный интерфейс между двумя средами выполнения.

## Типы

### bridgeFunc

```go
type bridgeFunc func(ctx *Context, args []C.JSValueRef) C.JSValueRef
```

Тип функции-обработчика моста, принимающего контекст и массив аргументов JSValueRef, возвращающий JSValueRef.

### Context

Структура, обертывающая C.JSGlobalContextRef. Используется для работы с JavaScriptCore контекстом.

### NodeID

Тип для идентификаторов узлов (Node), вероятно, используется для передачи идентификаторов между Go и JS.

### LogWriter

Глобальная переменная типа `io.Writer`, используемая для вывода в `console.log`. По умолчанию установлена в `io.Discard`.

## Публичный API

### goBridgeDispatch

```go
func goBridgeDispatch(
    funcName *C.char,
    ctx C.JSContextRef,
    function C.JSObjectRef,
    thisObject C.JSObjectRef,
    argumentCount C.size_t,
    arguments unsafe.Pointer,
    exceptionPtr unsafe.Pointer
) C.JSValueRef
```

Экспортируемая функция, вызываемая из JavaScript. Принимает имя функции, контекст, объект функции, this-объект, количество аргументов, массив аргументов и указатель на результат ошибки. Возвращает результат функции или ошибку.

### RegisterBridgeFunctions

```go
func (c *Context) RegisterBridgeFunctions() error
```

Регистрирует все мостовые функции на глобальном объекте JS. Каждая функция получает префикс `__go_`. Возвращает ошибку при неудаче.

### getArgString

```go
func getArgString(ctx *Context, val C.JSValueRef) (string, error)
```

Помощник для получения Go-строки из JSValueRef аргумента. Возвращает строку или ошибку при некорректном аргументе.

### jsToNodeID

```go
func jsToNodeID(ctx *Context, val C.JSValueRef) NodeID
```

Помощник для получения NodeID из JSValueRef. Возвращает -1, если значение не является объектом или не имеет свойства `__go_id`.

### nodeToJSValue

```go
func nodeToJSValue(ctx *Context, n *dom.Node) C.JSValueRef
```

Помощник для создания JSValueRef из узла dom.Node. Использует внутреннюю функцию `nodeToJS`.

### jsObjToVal

```go
func jsObjToVal(obj C.JSObjectRef) C.JSValueRef
```

Конвертирует JSObjectRef в JSValueRef через unsafe-преобразование.

## Bridge Handlers (Мостовые обработчики)

Карта `bridgeHandlers` связывает имена DOM-подобных функций с их Go-реализациями:

| Имя функции | Описание |
|-------------|----------|
| createElement | Создание элемента |
| appendChild | Добавление дочернего узла |
| setInnerHTML | Установка HTML-контента |
| setAttribute | Установка атрибута |
| setTextContent | Установка текстового контента |
| createTextNode | Создание текстового узла |
| createDocumentFragment | Создание фрагмента документа |
| setTimeout | Установка таймера |
| setInterval | Установка интервала |
| clearTimeout | Очистка таймера |
| clearInterval | Очистка интервала |
| getBody | Получение тега body |
| getElementById | Получение элемента по ID |
| getAttribute | Получение атрибута |
| removeChild | Удаление дочернего узла |
| getParentNode | Получение родительского узла |
| getChildren | Получение дочерних узлов |
| getInnerHTML | Получение HTML-контента |
| getTextContent | Получение текстового контента |
| getElementsByTagName | Получение элементов по тегу |
| getElementsByClassName | Получение элементов по классу |
| querySelector | Поиск первого элемента по селектору |
| querySelectorAll | Поиск всех элементов по селектору |
| addEventListener | Добавление слушателя событий |
| removeEventListener | Удаление слушателя событий |
| dispatchEvent | Диспатч событий |
| getComputedStyle | Получение вычисленных стилей |
| matches | Проверка соответствия селектору |
| closest | Поиск ближайшего предка |
| insertBefore | Вставка перед узлом |
| replaceChild | Замена дочернего узла |
| cloneNode | Клонирование узла |
| contains | Проверка вложенности |
| getBoundingClientRect | Получение границ элемента |
| previousElementSibling | Получение предыдущего сестринского узла |
| nextElementSibling | Получение следующего сестринского узла |
| firstElementChild | Получение первого дочернего элемента |
| lastElementChild | Получение последнего дочернего элемента |
| insertAdjacentHTML | Вставка HTML по позиции |
| insertAdjacentText | Вставка текста по позиции |
| insertAdjacentElement | Вставка элемента по позиции |
| outerHTML | Получение outerHTML |
| setOuterHTML | Установление outerHTML |
| log | Вывод в консоль |

## Реализация

### Архитектура

Функция `goBridgeDispatch`:
1. Преобразует C-строку в Go-строку
2. Создаёт Context из JSContextRef
3. Создает срез аргументов из unsafe.Pointer
4. Ищет обработчик в `bridgeHandlers`
5. Вызывает обработчик, если найдено
6. Если функция неизвестна, создаёт ошибку и устанавливает в exceptionPtr

### Обработка ошибок

- Ошибки устанавливаются через указатель `exceptionPtr` на JSValueRef
- Неизвестные функции генерируют ошибку "unknown function: {name}"
- Вспомогательные функции возвращают `error` для Go-кода

### Невозвращение ошибок

Некоторые функции (например, `goBridgeDispatch`) не возвращают ошибку напрямую, а устанавливают её в указатель `exceptionPtr`.

### Работа с памятью

- Используется `unsafe.Pointer` для прямого доступа к массиву аргументов
- C-строки освобождается через `defer C.free`
- JSStringRelease используется для освобождения JSStringRef

## Глобальные переменные

### LogWriter

```go
var LogWriter io.Writer = io.Discard
```

Глобальный writer для console.log. По умолчанию игнорирует вывод.

### bridgeHandlers

Глобальная карта, связывающая имена функций с их обработчиками.

## Внешние зависимости

- `github.com/starframe-dev/abyss/internal/dom` — пакет для работы с DOM-узлами
- JavaScriptCore C-библиотека (включена через cgo)

## Ключевые ограничения

1. Все функции должны быть зарегистрированы через `RegisterBridgeFunctions()`
2. Нераспознаные функции вызывают ошибку
3. Aргументы передаются как `JSValueRef`
4. Результат возвращается как `JSValueRef`
5. Ошибки устанавливаются в указатель `exceptionPtr`
