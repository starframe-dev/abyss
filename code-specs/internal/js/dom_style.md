# Дом-стиль (DOM Style)

## Обзор

`dom_style.go` предоставляет мост между DOM, стилями CSS и JavaScriptCore: хранит computed styles и layout-боксы, возвращает `Rect` для узла и создаёт JS-объект с выбранными вычисленными CSS-свойствами.

## Функции хранения данных

### `setComputedStyles(styles map[*dom.Node]css.ComputedStyle)` и `SetComputedStyles(styles map[*dom.Node]css.ComputedStyle)`
Обе функции записывают переданную карту в глобальную `computedStyles` под `computedStylesMu`. Экспортированная функция вызывает внутреннюю.

### `setLayoutBoxes(boxes map[*dom.Node]Rect)` и `SetLayoutBoxes(boxes map[*dom.Node]Rect)`
Обе функции записывают переданную карту в глобальную `layoutBoxes` под `layoutBoxesMu`. Экспортированная функция вызывает внутреннюю.

## Получение layout

### `getBoundingClientRect(nodeID NodeID) Rect`
Под `layoutBoxesMu` проверяет, что карта `layoutBoxes` не nil, затем ищет узел в `nodes` под `nodeMu` и соответствующий прямоугольник в карте. Если карта nil, узел отсутствует или равен nil, либо для него нет записи, возвращает `Rect{}`; иначе возвращает сохранённый `Rect`.

## Получение вычисленных стилей

### `getComputedStyle(nodeID NodeID) C.JSValueRef`
Под `computedStylesMu` проверяет карту стилей, ищет узел в `nodes` под `nodeMu` и запись для этого узла. Если карта nil, узел отсутствует или равен nil, либо стиль не найден, возвращает `C.JSValueMakeNull(eventCtx)`. Иначе создаёт JS-объект в `eventCtx`, записывает свойства и возвращает его через `jsObjToVal`.

Свойства объекта:
- Строковые: `display`, `position`, `overflow`, `widthUnit`, `heightUnit`, `color`, `backgroundColor`, `flexDirection`, `justifyContent`, `alignItems`.
- Числовые: `width` и `height` (поле `Px`); `marginTop`, `marginRight`, `marginBottom`, `marginLeft` (соответствующие `Px`); `paddingTop`, `paddingRight`, `paddingBottom`, `paddingLeft` (соответствующие `Px`); `borderTopWidth`, `borderRightWidth`, `borderBottomWidth`, `borderLeftWidth`; `flexGrow`, `flexShrink`, `zIndex` (преобразованный в `float64`), `gap` (поле `Px`).

Для строковых свойств создаются JS-строки, для числовых — JS-числа; свойства устанавливаются с атрибутом `kJSPropertyAttributeNone`.

## Преобразование CSS-типов

- `displayString`: `DisplayInline` → `inline`, `DisplayInlineBlock` → `inline-block`, `DisplayFlex` → `flex`, `DisplayGrid` → `grid`, `DisplayNone` → `none`; прочие значения → `block`.
- `positionString`: `PositionRelative` → `relative`, `PositionAbsolute` → `absolute`; прочие → `static`.
- `overflowString`: `OverflowHidden` → `hidden`; прочие → `visible`.
- `flexDirectionString`: `FlexDirectionColumn` → `column`; прочие → `row`.
- `justifyContentString`: `JustifyCenter` → `center`, `JustifyFlexEnd` → `flex-end`, `JustifySpaceBetween` → `space-between`, `JustifySpaceAround` → `space-around`; прочие → `flex-start`.
- `alignItemsString`: `AlignCenter` → `center`, `AlignFlexEnd` → `flex-end`, `AlignStretch` → `stretch`; прочие → `flex-start`.

## Зависимости и состояние

Файл импортирует `internal/css` и `internal/dom` и использует JavaScriptCore через CGO. В preamble объявлена используемая внешняя функция `_makeString`. Также используются объявленные в других частях пакета переменные `computedStyles`, `layoutBoxes`, `nodes` и соответствующие мьютексы, а также `eventCtx`; `Rect`, `NodeID` и `jsObjToVal` определены вне описываемого файла.

Обе функции получения удерживают мьютекс соответствующей карты на протяжении операции, а доступ к `nodes` отдельно защищён `nodeMu`. Внутренние функции установки освобождают мьютекс через `defer`.
