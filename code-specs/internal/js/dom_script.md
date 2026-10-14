# Спецификация: internal/js/dom_script.go

## Обзор

Этот пакет предоставляет интерфейс для выполнения JavaScript-скриптов и работы с DOM-деревом через JavaScriptCore (WebKit). Файл содержит функции для создания JS-объектов из Go-нот DOM, выполнения скриптов (встроенных и внешних) и возврата финальной структуры DOM.

## Публичный API

### Тип Context

`Context` — это тип, представляющий контекст выполнения JavaScript. Содержит ссылку на JSContextRef и предоставляет методы для работы с JS-значениями.

#### Методы

- `MakeNull()` → `C.JSValueRef`
  - Создаёт значение `null` в JS.
  
- `MakeNumber(value float64)` → `C.JSValueRef`
  - Создаёт числовое значение в JS.
  
- `MakeString(value string)` → `C.JSValueRef`
  - Создаёт строковое значение в JS.
  
- `MakeBoolean(value bool)` → `C.JSValueRef`
  - Создаёт булево значение в JS.
  
- `Evaluate(code string)` → `(*Result[string, error], error)`
  - Выполняет JavaScript-код. Возвращает результат и возможную ошибку.
  
- `MakeObject()` → `C.JSObjectRef`
  - Создаёт пустой JS-объект.
  
- `SetProperty(obj C.JSObjectRef, name C.CString, value C.JSValueRef, attr uint32)` → `error`
  - Устанавливает свойство на JS-объекте.

### Функция `nodeToJS(ctx *Context, n *dom.Node) C.JSValueRef`

Создаёт JS-объект-обертку для узла DOM.

#### Параметры

- `ctx` — контекст выполнения JS
- `n` — узел DOM для упаковки

#### Возвращаемое значение

- `C.JSValueRef` — JS-объект, представляющий узел DOM

#### Поведение

1. Получает ID узла через `getNodeID(n)`
2. Если ID < 0, возвращает `null`
3. Создаёт пустой JS-объект через `C.JSObjectMake()`
4. Устанавливает свойство `__go_id` с числовым ID узла
5. Устанавливает свойство `nodeType` в зависимости от типа узла:
   - `dom.ElementNode` → 1
   - `dom.TextNode` → 3
   - `dom.CommentNode` → 8
   - `dom.DocumentNode` → 9
   - `dom.DocumentFragmentNode` → 11
6. Для элементов (`ElementNode`) устанавливает свойство `tagName` (верхний регистр)
7. Для текстовых узлов (`TextNode`) устанавливает свойство `textContent`
8. Возвращает завершённый JS-объект через `jsObjToVal()`

### Функция `GetFinalDOM() *dom.Node`

Возвращает финальное DOM-дерево после выполнения JavaScript-скриптов.

#### Возвращаемое значение

- `*dom.Node` — корневой узел документа

### Функция `ExecuteScripts(ctx *Context, doc *dom.Node, baseURL string) error`

Выполняет все скрипты из тегов `<script>` в документе.

#### Параметры

- `ctx` — контекст выполнения JS
- `doc` — корневой узел документа
- `baseURL` — базовая URL для разрешения относительных путей скриптов

#### Поведение

1. Ищет все теги `<script>` в документе через `QueryAll("script")`
2. Для каждого скрипта:
   - Если есть атрибут `src`:
     - Запрашивает скрипт из внешнего URL через `fetchScript`
     - Если запрос успешен, выполняет код через `ctx.Evaluate()`
     - Ошибки выполнения игнорируются (тишина по умолчанию)
   - Если нет `src` (встроенный скрипт):
     - Получает текст из `TextContent()`
     - Если текст пустой, пропускает
     - Иначе выполняет код через `ctx.Evaluate()`
     - Ошибки выполнения игнорируются
3. Возвращает `nil` после обработки всех скриптов

**Важно:** Ошибки выполнения скриптов игнорируются. Это сделано потому, что большинство веб-страниц имеют ошибки JS, но продолжают работать.

### Функция `fetchScript(baseURL, src string) (string, error)`

Запрашивает и возвращает содержимое внешнего JavaScript-скрипта.

#### Параметры

- `baseURL` — базовая URL документа
- `src` — URL скрипта

#### Возвращаемое значение

- `string` — содержимое скрипта
- `error` — ошибка при запросе

#### Поведение

1. Парсит базовую URL через `url.Parse()`
2. Парсит скриптовую URL через `url.Parse()`
3. Разрешает относительный URL против базы через `base.ResolveReference()`
4. Создаёт HTTP-клиент с таймаутом 3 секунды
5. Создаёт GET-запрос
6. Устанавливает пользовательский агент: `"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15"`
7. Выполняет запрос
8. Читает тело ответа
9. Возвращает содержимое и возможную ошибку

#### Обработка ошибок

- Ошибки парсинга URL возвращаются как ошибки функции
- Ошибки HTTP-запроса возвращаются как ошибки функции
- Ошибки чтения тела ответа возвращаются как ошибки функции

## Типы и константы

### dom.Node

Тип узла DOM из пакета `github.com/starframe-dev/abyss/internal/dom`.

#### Атрибуты

- `Type` — тип узла (ElementNode, TextNode, CommentNode, DocumentNode, DocumentFragmentNode)
- `TagName` — имя тега элемента
- `Text` — текст текстового узла

#### Методы

- `GetAttr(name string)` — получает атрибут узла
- `TextContent()` — получает текст узла
- `QueryAll(selector string)` — выполняет запрос CSS-селектора

### dom.Node типы

- `ElementNode` — элемент (`1`)
- `TextNode` — текст (`3`)
- `CommentNode` — комментарий (`8`)
- `DocumentNode` — документ (`9`)
- `DocumentFragmentNode` — фрагмент документа (`11`)

## Внешние зависимости

### CGO импорты

```c
#include <JavaScriptCore/JavaScriptCore.h>

extern void _protectValue(JSContextRef ctx, JSValueRef value);
extern void _unprotectValue(JSContextRef ctx, JSValueRef value);
extern JSValueRef _callAsFunction(JSContextRef ctx, JSObjectRef func, JSObjectRef thisObj, size_t argCount, const JSValueRef args[], JSValueRef* exception);
extern JSValueRef _makeString(JSContextRef ctx, const char* value);
extern JSValueRef _makeBoolean(JSContextRef ctx, bool value);
extern JSValueRef _evaluate(JSContextRef ctx, const char* script, const char* sourceURL, JSValueRef* exception);
```

### Go импорты

- `"fmt"` — форматирование ошибок
- `"io"` — чтение тела HTTP-ответа
- `"net/http"` — HTTP-запросы
- `"net/url"` — парсинг URL
- `"strings"` — работа со строками
- `"time"` — таймауты
- `"github.com/starframe-dev/abyss/internal/dom"` — тип DOM

## Имплементационные детали

### CGO интерфейс

Файл использует CGO для вызова JavaScriptCore API. Экспортируемые функции C объявлены через `extern` и импортируются через `import "C"`.

### Управление памятью

- Все строковые константы JS через `C.CString(...)`
- Деферируются освобождение строк через `defer C.JSStringRelease(propName)`
- HTTP-ответ закрывается через `defer resp.Body.Close()`

### Обработка ошибок

- Ошибки выполнения скриптов игнорируются
- Ошибки парсинга URL возвращаются как ошибки функции
- Ошибки HTTP-запросов возвращаются как ошибки функции

### Таймауты

- Запросы к скриптам имеют таймаут 3 секунды

## Ограничения и примечания

1. Ошибки выполнения JS-скриптов игнорируются — это может привести к пропуску критических ошибок
2. Используется внешняя CGO зависимость от JavaScriptCore
3. HTTP-запросы к скриптам имеют фиксированный таймаут 3 секунды
4. Пользовательский агент установлен для совместимости с Safari/MacOS
5. Функция `GetFinalDOM` возвращает глобальную переменную `rootDoc`

## Примеры использования

```go
// Выполнение скриптов документа
err := ExecuteScripts(ctx, doc, baseURL)

// Получение финального DOM
finalDOM := GetFinalDOM()

// Создание JS-объекта из узла DOM
jsValue := nodeToJS(ctx, node)
```
