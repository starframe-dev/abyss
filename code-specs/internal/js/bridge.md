# Спецификация: internal/js/bridge.go

## Обзор

Файл реализует мост между Go и JavaScriptCore через cgo. Он определяет обработчики именованных bridge-функций, диспетчеризует вызовы из JavaScript и регистрирует функции `__go_<имя>` на глобальном JS-объекте.

## Типы и глобальные значения

- `bridgeFunc` — функция, принимающая `*Context` и `[]C.JSValueRef`, возвращающая `C.JSValueRef`.
- `LogWriter` — экспортированная переменная типа `io.Writer`; начальное значение `io.Discard`.
- `bridgeHandlers` — глобальная карта имен функций к обработчикам. В нее входят: `createElement`, `appendChild`, `setInnerHTML`, `setAttribute`, `setTextContent`, `createTextNode`, `createDocumentFragment`, `setTimeout`, `setInterval`, `clearTimeout`, `clearInterval`, `getBody`, `getElementById`, `getAttribute`, `removeChild`, `getParentNode`, `getChildren`, `getInnerHTML`, `getTextContent`, `getElementsByTagName`, `getElementsByClassName`, `querySelector`, `querySelectorAll`, `addEventListener`, `removeEventListener`, `dispatchEvent`, `getComputedStyle`, `matches`, `closest`, `insertBefore`, `replaceChild`, `cloneNode`, `contains`, `getBoundingClientRect`, `previousElementSibling`, `nextElementSibling`, `firstElementChild`, `lastElementChild`, `insertAdjacentHTML`, `insertAdjacentText`, `insertAdjacentElement`, `outerHTML`, `setOuterHTML`, `log`.

## Функции

- `goBridgeDispatch` — cgo-экспортируемая функция. Преобразует имя C-строки в Go-строку, оборачивает переданный `JSContextRef` в `Context`, копирует заданное число аргументов из `unsafe.Pointer` в Go-срез и вызывает соответствующий обработчик. Если имя отсутствует в карте, создает JS-строку `unknown function: <имя>`, записывает ее через `exceptionPtr` и возвращает nil. Неиспользуемые параметры callback передаются, но не обрабатываются.
- `getArgString` — преобразует значение с помощью `_valueToString`; освобождает полученную C-строку через `C.free` и возвращает ее содержимое. При nil-результате возвращает пустую строку и ошибку `arg error`. Значение исключения, полученное от `_valueToString`, отдельно не обрабатывается.
- `jsToNodeID` — если значение не объект, объект не получен, чтение свойства `__go_id` завершается исключением/пустым значением либо `ToNumber` возвращает ошибку, возвращает `-1`; иначе преобразует число в `NodeID`.
- `nodeToJSValue` — делегирует преобразование узла `*dom.Node` функции `nodeToJS`.
- `(*Context).RegisterBridgeFunctions` — для списка имен обработчиков вызывает `C.makeGoFunction`, формирует глобальное имя `__go_<имя>` и устанавливает его через `SetProperty`. При ошибке установки немедленно возвращает ошибку; иначе возвращает nil. C-строки имен, передаваемые в `makeGoFunction`, здесь явно не освобождаются.
- `jsObjToVal` — преобразует `C.JSObjectRef` в `C.JSValueRef` через `unsafe.Pointer`.

## Соответствие имен регистрации

Список в `RegisterBridgeFunctions` совпадает с ключами `bridgeHandlers` и включает те же 45 имен. Регистрация создает JS-функции для каждого имени; диспетчеризация выполняется отдельно по карте обработчиков.

## Зависимости

Используются JavaScriptCore API через cgo и пакет `github.com/starframe-dev/abyss/internal/dom`. Типы `Context` и `NodeID`, а также `nodeToJS` определены вне этого файла.
