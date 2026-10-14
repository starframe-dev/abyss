# computed_values

## Общее описание

Этот пакет отвечает за парсинг CSS-значений и их конвертацию в типизированные computed-стилевые значения. Функции преобразуют строковые значения свойств CSS в соответствующие типы данных, используемые в системе стилей.

## Публичный API

### parseDisplay

**Подпись:** `func parseDisplay(val string) DisplayType`

**Назначение:** Парсит значение свойства `display` в тип `DisplayType`.

**Возражения:**
- `"block"` → `DisplayBlock`
- `"inline"` → `DisplayInline`
- `"flex"` → `DisplayFlex`
- `"grid"` → `DisplayGrid`
- `"none"` → `DisplayNone`

**Поведение по умолчанию:** Возвращает `DisplayBlock` для неизвестных значений.

---

### parsePosition

**Подпись:** `func parsePosition(val string) PositionType`

**Назначение:** Парсит значение свойства `position` в тип `PositionType`.

**Возражения:**
- `"static"` → `PositionStatic`
- `"relative"` → `PositionRelative`
- `"absolute"` → `PositionAbsolute`

**Поведение по умолчанию:** Возвращает `PositionStatic` для неизвестных значений.

---

### parseOverflow

**Подпись:** `func parseOverflow(val string) OverflowType`

**Назначение:** Парсит значение свойства `overflow` в тип `OverflowType`.

**Возражения:**
- `"hidden"` → `OverflowHidden`
- `"visible"` → `OverflowVisible`

**Поведение по умолчанию:** Возвращает `OverflowVisible` для неизвестных значений.

---

### parseLength

**Подпись:** `func parseLength(val string) Length`

**Назначение:** Парсит CSS-длину (px, %, rem/em, fr, простое число) в тип `Length`.

**Поддерживаемые форматы:**
- `"auto"` → `Auto()`
- `<число>px` → `Px(<число>)`
- `<число>%` → `Percent(<число>)`
- `<число>rem` → `Px(<число> * 16)` (примечание: rem трактуется как px)
- `<число>em` → `Px(<число> * 16)` (примечание: em трактуется как px)
- `<число>fr` → `Px(<число>)`
- `<число>` (без единиц) → `Px(<число>)`

**Поведение по умолчанию:** Возвращает `Auto()` для неизвестных значений.

---

### parseBoxModel

**Подпись:** `func parseBoxModel(val string) BoxModel`

**Назначение:** Парсит 1-4 значения в shorthand-формате в тип `BoxModel`.

**Форматы:**
- 1 значение: `t` → `BoxModel{Top: t, Right: t, Bottom: t, Left: t}`
- 2 значения: `t r` → `BoxModel{Top: t, Right: r, Bottom: t, Left: r}`
- 3 значения: `t r b` → `BoxModel{Top: t, Right: r, Bottom: b, Left: r}`
- 4 значения: `t r b l` → `BoxModel{Top: t, Right: r, Bottom: b, Left: l}`

**Поведение по умолчанию:** Возвращает `BoxModel{}` для неизвестных значений.

---

### parseBorder

**Подпись:** `func parseBorder(val string) Border`

**Назначение:** Парсит shorthand-значение свойства `border` в тип `Border`.

**Поведение:**
- Обрабатывает значения ширины, цвета и стиля
- По умолчанию возвращает `Border{Style: BorderNone}`

---

### parseBorderWidth

**Подпись:** `func parseBorderWidth(val string) float64`

**Назначение:** Парсит ширину границы в значение `float64`.

**Поддерживаемые форматы:**
- `<число>px` → `<число>` (px отбрасывается)

---

### parseBorderStyle

**Подпись:** `func parseBorderStyle(val string) BorderStyle`

**Назначение:** Парсит значение стиля границы в тип `BorderStyle`.

**Возражения:**
- `"none"` → `BorderNone`
- `"solid"` → `BorderSolid`

**Поведение по умолчанию:** Возвращает `BorderNone` для неизвестных значений.

---

### parseColor

**Подпись:** `func parseColor(val string) Color`

**Назначение:** Парсит CSS-цвет (именованный, hex, rgb()) в тип `Color`.

**Поддерживаемые форматы:**
- Именующие цвета (например, `"red"`, `"blue"` и т.д.)
- HEX-цвета (например, `"#RRGGBB"`, `"#RGB"`)
- `rgb(r, g, b)`

**Поведение по умолчанию:** Возвращает `Transparent()` для неизвестных значений.

**Именованные цвета:**
- `"transparent"` → `Transparent()`
- `"black"` → `RGB(0, 0, 0)`
- `"white"` → `RGB(255, 255, 255)`
- `"red"` → `RGB(255, 0, 0)`
- `"green"` → `RGB(0, 128, 0)`
- `"blue"` → `RGB(0, 0, 255)`
- `"yellow"` → `RGB(255, 255, 0)`
- `"cyan"` → `RGB(0, 255, 255)`
- `"magenta"` → `RGB(255, 0, 255)`
- `"gray"` → `RGB(128, 128, 128)`
- `"lightgray"` → `RGB(211, 211, 211)`
- `"darkgray"` → `RGB(169, 169, 169)`
- `"orange"` → `RGB(255, 165, 0)`
- `"purple"` → `RGB(128, 0, 128)`
- `"pink"` → `RGB(255, 192, 203)`
- `"brown"` → `RGB(165, 42, 42)`

---

### parseRgb

**Подпись:** `func parseRgb(val string) (Color, bool)`

**Назначение:** Парсит значение `rgb(r, g, b)` в тип `Color`. Возвращает `ok = false` для неизвестных значений.

---

### parseFlexDirection

**Подпись:** `func parseFlexDirection(val string) FlexDirection`

**Назначение:** Парсит значение свойства `flex-direction` в тип `FlexDirection`.

**Возражения:**
- `"row"` → `FlexDirectionRow`
- `"column"` → `FlexDirectionColumn`

**Поведение по умолчанию:** Возвращает `FlexDirectionRow` для неизвестных значений.

---

### parseJustifyContent

**Подпись:** `func parseJustifyContent(val string) JustifyContent`

**Назначение:** Парсит значение свойства `justify-content` в тип `JustifyContent`.

**Возражения:**
- `"flex-start"` → `JustifyFlexStart`
- `"center"` → `JustifyCenter`
- `"flex-end"` → `JustifyFlexEnd`
- `"space-between"` → `JustifySpaceBetween`
- `"space-around"` → `JustifySpaceAround`

**Поведение по умолчанию:** Возвращает `JustifyFlexStart` для неизвестных значений.

---

### parseAlignItems

**Подпись:** `func parseAlignItems(val string) AlignItems`

**Назначение:** Парсит значение свойства `align-items` в тип `AlignItems`.

**Возражения:**
- `"flex-start"` → `AlignFlexStart`
- `"center"` → `AlignCenter`
- `"flex-end"` → `AlignFlexEnd`
- `"stretch"` → `AlignStretch`

**Поведение по умолчанию:** Возвращает `AlignStretch` для неизвестных значений.

---

### parseLengthList

**Подпись:** `func parseLengthList(val string) []Length`

**Назначение:** Парсит список длин, поддерживая синтаксис `repeat()`.

**Поддерживаемые форматы:**
- Простой список: `"1fr 1fr 1fr"` → `[]Length{Px(1), Px(1), Px(1)}`
- С синтаксисом repeat: `"repeat(3, 1fr)"` → `[]Length{Px(1), Px(1), Px(1)}`

**Поведение:** Разделяет значения по пробелам и запятым.

---

### parseInt

**Подпись:** `func parseInt(val string) int`

**Назначение:** Парсит целое число, игнорируя ошибки.

---

### parseFloat

**Подпись:** `func parseFloat(val string) float64`

**Назначение:** Парсит число с плавающей точкой, игнорируя ошибки.

---

### isNumber

**Подпись:** `func isNumber(s string) bool`

**Назначение:** Проверяет, является ли строка валидным числом.

---

## Типы

### DisplayType

Перечисляет возможные значения свойства `display`.

```go
const (
    DisplayBlock DisplayType = iota
    DisplayInline
    DisplayFlex
    DisplayGrid
    DisplayNone
)
```

### PositionType

Перечисляет возможные значения свойства `position`.

```go
const (
    PositionStatic PositionType = iota
    PositionRelative
    PositionAbsolute
)
```

### OverflowType

Перечисляет возможные значения свойства `overflow`.

```go
const (
    OverflowVisible OverflowType = iota
    OverflowHidden
)
```

### FlexDirection

Перечисляет возможные значения `flex-direction`.

```go
const (
    FlexDirectionRow FlexDirection = iota
    FlexDirectionColumn
)
```

### JustifyContent

Перечисляет возможные значения `justify-content`.

```go
const (
    JustifyFlexStart JustifyContent = iota
    JustifyCenter
    JustifyFlexEnd
    JustifySpaceBetween
    JustifySpaceAround
)
```

### AlignItems

Перечисляет возможные значения `align-items`.

```go
const (
    AlignFlexStart AlignItems = iota
    AlignCenter
    AlignFlexEnd
    AlignStretch
)
```

### BorderStyle

Перечисляет возможные значения стиля границы.

```go
const (
    BorderNone BorderStyle = iota
    BorderSolid
)
```

### Length

Структура для представления CSS-длины.

```go
type Length struct {
    IsAuto    bool
    Px        float64
    IsPercent bool
    Percent   float64
}
```

**Методы:**

- `Auto() Length` — возвращает `Length{IsAuto: true}`
- `Px(v float64) Length` — возвращает `Length{Px: v}`
- `Percent(v float64) Length` — возвращает `Length{IsPercent: true, Percent: v}`
- `Resolve(ref float64) float64` — решает длину относительно опорного значения (для процентов)
- `String() string` — возвращает строковое представление

### Color

Структура для представления RGB-цвета.

```go
type Color struct {
    R, G, B uint8
    IsSet   bool
}
```

**Методы:**

- `Transparent() Color` — возвращает прозрачный цвет
- `RGB(r, g, b uint8) Color` — возвращает цвет из RGB значений
- `Hex(s string) Color` — парсит HEX-цвет
- `String() string` — возвращает строковое представление

### Border

Структура для представления CSS-границы.

```go
type Border struct {
    Width float64
    Style BorderStyle
    Color Color
}
```

### BoxModel

Структура для представления box-model свойств (top, right, bottom, left).

```go
type BoxModel struct {
    Top, Right, Bottom, Left Length
}
```

**Методы:**

- `IsZero() bool` — возвращает `true`, если все стороны равны нулю

---

## Важные детали реализации

### Обработка ошибок

Все функции парсинга используют стратегию "проглотить ошибку". При невозможности парсинга возвращается значение по умолчанию:

- `DisplayType` → `DisplayBlock`
- `PositionType` → `PositionStatic`
- `OverflowType` → `OverflowVisible`
- `BorderStyle` → `BorderNone`
- `Length` → `Auto()`
- `Color` → `Transparent()`

### Преобразование rem/em в px

Для простоты реализации `rem` и `em` тракуются как `px` с множителем 16:

```go
v, _ := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSuffix(val, "rem"), "em"), 64)
return Px(v * 16)
```

### Синтаксис repeat()

Функция `parseLengthList` поддерживает CSS-синтаксис `repeat(N, value)`:

```go
if strings.HasPrefix(val, "repeat(") {
    end := strings.Index(val, ")")
    if end < 0 {
        break
    }
    inner := val[7:end]
    comma := strings.Index(inner, ",")
    if comma > 0 {
        count, _ := strconv.Atoi(strings.TrimSpace(inner[:comma]))
        item := strings.TrimSpace(inner[comma+1:])
        for i := 0; i < count; i++ {
            result = append(result, parseLength(item))
        }
    }
    val = val[end+1:]
    continue
}
```

### Именованные цвета

Доступно 16 именованных CSS-цветов, включая стандартные и распространенные цвета.

### HEX-парсинг

Функция `Hex` поддерживает как 3-значный (`#RGB`), так и 6-значный (`#RRGGBB`) формат.
