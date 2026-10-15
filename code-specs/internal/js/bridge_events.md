# bridge_events

## Обзор

`bridge_events.go` содержит функции-обработчики JavaScript API таймеров и событий, работающие через JavaScriptCore bridge. Все функции имеют непубличный для Go уровень доступа (начинаются со строчной буквы) и возвращают `C.JSValueRef`.

## Обработчики таймеров

- `setTimeoutJS(ctx, args)` требует не менее двух аргументов. Первый преобразуется через `JSValueToObject`; если результат nil, возвращает число `0`. Второй преобразуется через `ctx.ToNumber`; при ошибке задержка остаётся `0`, иначе число приводится к `int`. Вызывает `setTimeout(callback, delay)` и возвращает выданный ID как JS-число.
- `setIntervalJS(ctx, args)` имеет такую же обработку аргументов, вызывает `setInterval` и возвращает ID как JS-число.
- `clearTimeoutJS(ctx, args)` и `clearIntervalJS(ctx, args)` при отсутствии аргумента возвращают `undefined`. Если преобразование первого аргумента в число успешно, приводят его к `int` и вызывают `clearTimer`; при ошибке ничего не очищают. Всегда возвращают `undefined`.

## Обработчики событий

- `addEventListenerJS(ctx, args)` и `removeEventListenerJS(ctx, args)` требуют не менее трёх аргументов. Первый преобразуется в Node ID через `jsToNodeID`; если ID отрицательный, обработка прекращается с `undefined`. Второй передаётся в `getArgString` (ошибка игнорируется), третий преобразуется через `JSValueToObject`; при nil обработка прекращается. Затем вызывается соответствующая внутренняя функция, результат — `undefined`.
- `dispatchEventJS(ctx, args)` требует не менее двух аргументов, преобразует первый в Node ID и при отрицательном ID возвращает `undefined`. Второй передаётся в `getArgString` с игнорированием ошибки, после чего вызывается `dispatchEvent`; результат — `undefined`.

Проверка, что callback действительно является функцией, отдельно не выполняется: используется только преобразование в объект и проверка nil.

## CGo объявления

C-пролог включает `JavaScriptCore/JavaScriptCore.h` и `stdlib.h` и объявляет `makeGoFunction`, `_getType`, `_valueToString`, `_makeArray`, `_makeString`, `_makeBoolean`, `_fprintf_stderr`, `_protectValue`, `_unprotectValue` и `_callAsFunction`. Этот файл сам их не вызывает.

## Зависимости

Используются тип `Context` и методы `MakeNumber`, `MakeUndefined`, `ToNumber`, а также функции `setTimeout`, `setInterval`, `clearTimer`, `jsToNodeID`, `getArgString`, `addEventListener`, `removeEventListener` и `dispatchEvent`, определённые в других частях пакета.
