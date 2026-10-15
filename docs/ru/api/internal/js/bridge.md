# Мост JavaScript

`internal/js/bridge.go` связывает значения JavaScriptCore и вызываемые функции JavaScript с реализациями DOM API на Go. Он регистрирует функции моста в глобальном объекте контекста JavaScript, преобразует входящие аргументы и передаёт вызовы обработчикам.

## Публичный API

### `LogWriter`

```go
var LogWriter io.Writer = io.Discard
```

Писатель, используемый обработчиком моста `console.log`. По умолчанию задан `io.Discard`, поэтому вывод отбрасывается, пока приложение не назначит другой писатель. Фактическое форматирование и запись реализованы в `logJS`.

### `(*Context).RegisterBridgeFunctions`

```go
func (c *Context) RegisterBridgeFunctions() error
```

Регистрирует точки входа моста в глобальном объекте контекста под именами с префиксом `__go_` (например, `__go_createElement`). Возвращает первую ошибку при установке глобального свойства; при её возникновении регистрация прекращается. Регистрируемые имена:

- Создание и изменение DOM: `createElement`, `createTextNode`, `createDocumentFragment`, `appendChild`, `insertBefore`, `replaceChild`, `removeChild`, `cloneNode`, `setInnerHTML`, `setOuterHTML`, `setTextContent`, `setAttribute`, `insertAdjacentHTML`, `insertAdjacentText`, `insertAdjacentElement`
- Доступ к DOM и поиск: `getBody`, `getElementById`, `getAttribute`, `getParentNode`, `getChildren`, `getInnerHTML`, `getTextContent`, `getElementsByTagName`, `getElementsByClassName`, `querySelector`, `querySelectorAll`, `matches`, `closest`, `contains`, `outerHTML`
- События, таймеры и отображение: `addEventListener`, `removeEventListener`, `dispatchEvent`, `setTimeout`, `setInterval`, `clearTimeout`, `clearInterval`, `getComputedStyle`, `getBoundingClientRect`
- Связи между элементами: `previousElementSibling`, `nextElementSibling`, `firstElementChild`, `lastElementChild`
- Журналирование: `log`

Соответствующие обработчики Go хранятся в карте пакета `bridgeHandlers`. При регистрации каждое имя передаётся C-функции `makeGoFunction`, а полученная функция устанавливается с помощью `Context.SetProperty`.

## Диспетчеризация и преобразования

Экспортируемый cgo-обратный вызов `goBridgeDispatch` — общая точка входа, вызываемая C-мостом. Он преобразует имя C-функции в строку Go, оборачивает контекст JavaScriptCore в `Context` пакета, копирует переданный массив аргументов в управляемую Go память и вызывает соответствующий обработчик. Если обработчик не найден, функция создаёт исключение JavaScript в виде строки `unknown function: <name>` и возвращает нулевое значение.

Эту границу обслуживают внутренние вспомогательные функции:

- `getArgString` преобразует значение JavaScript в строку Go через `_valueToString`, освобождает полученное выделение памяти C и возвращает ошибку, если преобразование дало нулевой указатель.
- `jsToNodeID` принимает только объект JavaScript с числовым свойством `__go_id`, которое можно преобразовать с помощью `Context.ToNumber`. При любой неудачной проверке типа, поиске объекта или свойства либо преобразовании числа возвращается маркер `NodeID(-1)`.
- `nodeToJSValue` оборачивает `*dom.Node` с помощью `nodeToJS`.
- `jsObjToVal` интерпретирует `JSObjectRef` как `JSValueRef` для установки свойства.

## Примечания по реализации

В этом файле используются cgo и типы JavaScriptCore из `<JavaScriptCore/JavaScriptCore.h>`, а также вспомогательные функции, объявленные в прологе C и реализованные в C-коде моста. Перед передачей обработчикам Go обратный вызов копирует аргументы, поэтому обработчики не зависят от C-среза аргументов после завершения обратного вызова. Имена функций моста указаны и в карте обработчиков, и в списке регистрации; чтобы добавить обработчик, необходимо также сделать его имя доступным при регистрации.

`bridgeHandlers`, `bridgeFunc` и функции преобразования являются внутренними для пакета, а не частью публичного API. Поведение DOM на стороне моста обеспечивается отдельными реализациями обработчиков и пакетом `internal/dom`; этот файл определяет диспетчеризацию, регистрацию и преобразование значений, но не подробную семантику этих операций.
