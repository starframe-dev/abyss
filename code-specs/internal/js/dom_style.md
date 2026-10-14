# Дом-стиль (DOM Style)

## Обзор

Модуль `dom_style.go` предоставляет мост между Go-реализацией DOM и JavaScriptCore для работы со стилями. Он реализует функции для получения вычисленных стилей (`getComputedStyle`) и границ элементов (`getBoundingClientRect`) через CGO-интерфейс.

## Публичный API

### Функции

#### `setComputedStyles(styles map[*dom.Node]css.ComputedStyle)`

Записывает вычисленные стили для всех узлов DOM.

**Параметры:**
- `styles` — карта от `*dom.Node` к `css.ComputedStyle`, содержащая стили для каждого узла

**Поведение:**
- Функция использует глобальную переменную `computedStylesMu` (мьютекс) для защиты от гонок данных
- После записи стили доступны через `getComputedStyle`

#### `setLayoutBoxes(boxes map[*dom.Node]Rect)`

Записывает layout-боксы для всех узлов DOM.

**Параметры:**
- `boxes` — карта от `*dom.Node` к `Rect`, содержащая layout-информацию для каждого узла

**Поведение:**
- Функция использует глобальную переменную `layoutBoxesMu` (мьютекс) для защиты от гонок данных
- После записи боксы доступны через `getBoundingClientRect`

#### `getBoundingClientRect(nodeID NodeID) Rect`

Возвращает прямоугольник границ элемента.

**Параметры:**
- `nodeID` — идентификатор узла DOM

**Возвращаемое значение:**
- `Rect` — структура с полями `X`, `Y`, `Width`, `Height` (float64)
- Возвращает пустой `Rect{}` если узел не найден или layout-информация отсутствует

**Поведение:**
1. Проверяет наличие `layoutBoxes`
2. Ищет узел по `nodeID` через глобальную `nodes` карту
3. Проверяет наличие layout-информации для узла
4. Возвращает соответствующий `Rect` или пустой `Rect`

#### `getComputedStyle(nodeID NodeID) C.JSValueRef`

Возвращает объект JS с вычисленными CSS-свойствами для узла.

**Параметры:**
- `nodeID` — идентификатор узла DOM

**Возвращаемое значение:**
- `C.JSValueRef` — ссылка на JS-объект в JavaScriptCore

**Поведение:**
1. Проверяет наличие `computedStyles`
2. Ищет узел по `nodeID` через глобальную `nodes` карту
3. Проверяет наличие стилей для узла
4. Создаёт JS-объект и заполняет его свойствами:
   - `display`, `position`, `overflow` — строковые представления CSS свойств
   - `width`, `widthUnit`, `height`, `heightUnit` — размеры и единицы
   - `color`, `backgroundColor` — цвета
   - `marginTop`, `marginRight`, `marginBottom`, `marginLeft` — отступы
   - `paddingTop`, `paddingRight`, `paddingBottom`, `paddingLeft` — внутренние отступы
   - `borderTopWidth`, `borderRightWidth`, `borderBottomWidth`, `borderLeftWidth` — ширина границ
   - `flexDirection`, `justifyContent`, `alignItems` — flex-свойства
   - `flexGrow`, `flexShrink` — flex-grow/shrink
   - `zIndex` — индекс наложения
   - `gap` — gap для flex/grid
5. Использует вспомогательные функции `displayString`, `positionString`, `overflowString` и др. для конвертации CSS-типов в строки

### Вспомогательные функции

#### `displayString(d css.DisplayType) string`

Конвертирует тип `Display` в CSS-строку.

| Тип | Возвращаемое значение |
|-----|----------------------|
| `css.DisplayInline` | `"inline"` |
| `css.DisplayFlex` | `"flex"` |
| `css.DisplayGrid` | `"grid"` |
| `css.DisplayNone` | `"none"` |
| другие | `"block"` |

#### `positionString(p css.PositionType) string`

Конвертирует тип `Position` в CSS-строку.

| Тип | Возвращаемое значение |
|-----|----------------------|
| `css.PositionRelative` | `"relative"` |
| `css.PositionAbsolute` | `"absolute"` |
| другие | `"static"` |

#### `overflowString(o css.OverflowType) string`

Конвертирует тип `Overflow` в CSS-строку.

| Тип | Возвращаемое значение |
|-----|----------------------|
| `css.OverflowHidden` | `"hidden"` |
| другие | `"visible"` |

#### `flexDirectionString(fd css.FlexDirection) string`

Конвертирует тип `FlexDirection` в CSS-строку.

| Тип | Возвращаемое значение |
|-----|----------------------|
| `css.FlexDirectionColumn` | `"column"` |
| другие | `"row"` |

#### `justifyContentString(j css.JustifyContent) string`

Конвертирует тип `JustifyContent` в CSS-строку.

| Тип | Возвращаемое значение |
|-----|----------------------|
| `css.JustifyCenter` | `"center"` |
| `css.JustifyFlexEnd` | `"flex-end"` |
| `css.JustifySpaceBetween` | `"space-between"` |
| `css.JustifySpaceAround` | `"space-around"` |
| другие | `"flex-start"` |

#### `alignItemsString(a css.AlignItems) string`

Конвертирует тип `AlignItems` в CSS-строку.

| Тип | Возвращаемое значение |
|-----|----------------------|
| `css.AlignFlexStart` | `"flex-start"` |
| `css.AlignCenter` | `"center"` |
| `css.AlignFlexEnd` | `"flex-end"` |
| другие | `"stretch"` |

## Типы данных

### `Rect`

Структура, представляющая прямоугольник в layout-системе.

```go
type Rect struct {
    X, Y, Width, Height float64
}
```

| Поле | Тип | Описание |
|------|-----|----------|
| `X` | float64 | Координата X левого края |
| `Y` | float64 | Координата Y верхнего края |
| `Width` | float64 | Ширина прямоугольника |
| `Height` | float64 | Высота прямоугольника |

## CGO-интерфейс

Файл использует внешние функции JavaScriptCore через CGO:

```c
extern void _protectValue(JSContextRef ctx, JSValueRef value);
extern void _unprotectValue(JSContextRef ctx, JSValueRef value);
extern JSValueRef _callAsFunction(JSContextRef ctx, JSObjectRef func, JSObjectRef thisObj, size_t argCount, const JSValueRef args[], JSValueRef* exception);
extern JSValueRef _makeString(JSContextRef ctx, const char* value);
extern JSValueRef _makeBoolean(JSContextRef ctx, bool value);
extern JSValueRef _evaluate(JSContextRef ctx, const char* script, const char* sourceURL, JSValueRef* exception);
```

## Внутренние реализации

### `setStr(name, val string)`

Вспомогательная функция для установки строкового свойства в JS-объект.

**Параметры:**
- `name` — имя свойства
- `val` — значение свойства

**Поведение:**
1. Создаёт строку из UTF-8
2. Вызывает `C.JSObjectSetProperty` для установки свойства в JS-объект
3. Освобождает ресурсы через `defer C.JSStringRelease`

### `setNum(name string, val float64)`

Вспомогательная функция для установки числового свойства в JS-объект.

**Параметры:**
- `name` — имя свойства
- `val` — числовое значение

**Поведение:**
1. Создаёт строку из имени свойства
2. Создаёт числовое JS-значение
3. Вызывает `C.JSObjectSetProperty` для установки свойства в JS-объект
4. Освобождает ресурсы через `defer C.JSStringRelease`

### `displayString`, `positionString`, `overflowString` и др.

Все функции конвертации используют switch-выражения для отображения внутренних типов CSS на стандартные CSS-строки.

## Глобальные переменные

- `computedStyles` — карта `map[*dom.Node]css.ComputedStyle` для хранения вычисленных стилей
- `layoutBoxes` — карта `map[*dom.Node]Rect` для хранения layout-информации
- `nodes` — глобальная карта узлов DOM
- `computedStylesMu` — мьютекс для защиты `computedStyles`
- `layoutBoxesMu` — мьютекс для защиты `layoutBoxes`
- `nodeMu` — мьютекс для защиты `nodes`
- `eventCtx` — контекст события JavaScriptCore

## Ограничения и предположения

1. **Пустой результат** — если узел не найден или данные отсутствуют, возвращается пустой `Rect` или `JSValueNull`
2. **Goroutine safety** — все изменения глобальных переменных защищены мьютексами
3. **JSCore dependency** — требует наличия JavaScriptCore библиотеки и корректной настройки CGO
4. **Memory management** — CGO-ресурсы освобождаются через `defer` вызовы
5. **CSS types** — все CSS-типы должны быть предварительно определены в пакете `css`
