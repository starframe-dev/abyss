# `internal/js`

Пакет `js` предоставляет выполнение JavaScript и преобразование значений и свойств через фреймворк JavaScriptCore (JSC) от Apple. Это мост Go-to-C, которому требуется фреймворк JavaScriptCore в macOS; исходный код включает cgo и компилирует код моста на Objective-C.

## Жизненный цикл контекста

`Context` владеет глобальным контекстом выполнения JSC. Создайте его с помощью `NewContext`, используйте его методы для вычисления и операций со значениями, затем вызовите `Close`, чтобы освободить нативный контекст. `Close` также сбрасывает систему событий пакета для этого контекста. Вызывать `Close` повторно безопасно: после освобождения сохранённая нативная ссылка становится nil. Остальные методы предполагают, что контекст активен; они не проверяют, был ли он закрыт.

```go
ctx := js.NewContext()
defer ctx.Close()

result, err := ctx.Evaluate("1 + 2")
```

## Публичный API

| API | Поведение |
| --- | --- |
| `NewContext() *Context` | Создаёт новый глобальный контекст JavaScriptCore. |
| `(*Context).Close()` | Сбрасывает состояние системы событий для контекста и освобождает его. |
| `(*Context).GetContextRef() C.JSGlobalContextRef` | Предоставляет базовую ссылку на контекст cgo/JSC. |
| `(*Context).Evaluate(script string) (string, error)` | Выполняет исходный код и преобразует результат в строковое представление JavaScript. В качестве URL источника использует `script.js`. |
| `(*Context).EvaluateRaw(script string) (C.JSValueRef, error)` | Выполняет исходный код и возвращает необработанную ссылку на значение JSC. В качестве URL источника использует `script.js`. |
| `(*Context).GlobalObject() C.JSObjectRef` | Возвращает глобальный объект контекста. |
| `(*Context).SetProperty(name string, value C.JSValueRef) error` | Устанавливает свойство глобального объекта. |
| `(*Context).SetPropertyOn(obj C.JSObjectRef, name string, value C.JSValueRef) error` | Устанавливает свойство указанного объекта. |
| `(*Context).GetProperty(name string) (C.JSValueRef, error)` | Получает свойство глобального объекта. |
| `(*Context).ValueToString(value C.JSValueRef) (string, error)` | Преобразует значение JSC в строку Go. |
| `(*Context).MakeString(value string) C.JSValueRef` | Создаёт строку JavaScript. |
| `(*Context).MakeNumber(value float64) C.JSValueRef` | Создаёт число JavaScript. |
| `(*Context).MakeBoolean(value bool) C.JSValueRef` | Создаёт логическое значение JavaScript. |
| `(*Context).MakeUndefined() C.JSValueRef` | Создаёт JavaScript `undefined`. |
| `(*Context).MakeNull() C.JSValueRef` | Создаёт JavaScript `null`. |
| `(*Context).ToNumber(value C.JSValueRef) (float64, error)` | Выполняет числовое преобразование JavaScriptCore и возвращает значение Go `float64`. |
| `(*Context).ToBoolean(value C.JSValueRef) bool` | Выполняет преобразование по правилам истинности JavaScript. |
| `(*Context).GetType(value C.JSValueRef) string` | Возвращает нормализованное имя типа JavaScript. |

Типы `C.JSValueRef`, `C.JSObjectRef` и `C.JSGlobalContextRef` представляют ссылки JavaScriptCore в cgo. Они не принадлежат Go; вызывающий код должен использовать их с создавшим их контекстом и не сохранять дольше срока его жизни.

## Выполнение и ошибки

`Evaluate` и `EvaluateRaw` вызывают вычислитель моста C и проверяют выходные данные исключений JavaScript. Исключения преобразуются в ошибки Go с префиксом `js error:`; если исключение невозможно преобразовать в строку, методы возвращают `unknown js error`. `Evaluate` преобразует успешный результат в строку, а `EvaluateRaw` сохраняет нативную ссылку на значение. Результат nil от `Evaluate` представляется пустой строкой без ошибки.

Операции со свойствами и числовое преобразование также сообщают об исключениях JSC как об ошибках Go с префиксами, соответствующими операции (`set property error:`, `get property error:` или `to number error:`); если исключение невозможно преобразовать, используется сообщение `unknown ... error`. `ValueToString` сообщает об исключениях преобразования с префиксом `value to string error:`; если строковый результат отсутствует и исключение не сообщалось, метод возвращает пустую строку без ошибки.

## Поведение типов и преобразований

`GetType` сопоставляет типы JSC со значениями `undefined`, `null`, `boolean`, `number`, `string` или `object`. Любой тип JSC, не охваченный этими вариантами, преобразуется в `unknown`.

`ToBoolean` напрямую возвращает результат логического преобразования JSC и не предоставляет результат ошибки. `ToNumber` может завершиться ошибкой, если JavaScriptCore выбросит исключение во время преобразования. Конструкторы и методы работы со свойствами выделяют временные строки C для имён и строковых значений и освобождают эти буферы после вызова моста.

## Примечания по реализации

Обёртка Go делегирует нативные операции мосту Objective-C/C, объявленному в преамбуле cgo, включая выполнение кода, доступ к свойствам, создание примитивов, преобразование и проверку типов. `Evaluate` и `EvaluateRaw` используют фиксированный URL источника `script.js`; API пока не позволяет вызывающим задавать URL. Вызовы JSC и нативные ссылки привязаны к контексту, поэтому используйте один контекст последовательно и закрывайте его по завершении работы.
