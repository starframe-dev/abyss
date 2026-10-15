# Типы CSS

Пакет `css` определяет типы значений, используемые для представления подмножества свойств CSS, поддерживаемых приложением. Он предоставляет перечисления дискретных значений свойств, объекты значений для размеров и цветов, модели блоков и границ, а также структуру `ComputedStyle` с конструктором поддерживаемых значений по умолчанию.

## Дискретные типы свойств

Все перечисленные ниже типы свойств — целочисленные перечисления. Их константы используют `iota`, поэтому числовые значения начинаются с нуля и следуют в указанном порядке. Они описывают поддерживаемые значения; преобразование в текст CSS этими типами не предусмотрено.

| Тип | Константы в порядке объявления | Свойство |
|---|---|---|
| `DisplayType` | `DisplayBlock`, `DisplayInline`, `DisplayFlex`, `DisplayGrid`, `DisplayNone` | `display` |
| `PositionType` | `PositionStatic`, `PositionRelative`, `PositionAbsolute` | `position` |
| `OverflowType` | `OverflowVisible`, `OverflowHidden` | `overflow` |
| `FlexDirection` | `FlexDirectionRow`, `FlexDirectionColumn` | `flex-direction` |
| `JustifyContent` | `JustifyFlexStart`, `JustifyCenter`, `JustifyFlexEnd`, `JustifySpaceBetween`, `JustifySpaceAround` | `justify-content` |
| `AlignItems` | `AlignFlexStart`, `AlignCenter`, `AlignFlexEnd`, `AlignStretch` | `align-items` |
| `BorderStyle` | `BorderNone`, `BorderSolid` | стиль границы |

## Размеры

`Length` представляет размер в пикселях, процентный размер или автоматический размер:

```go
type Length struct {
    IsAuto    bool
    Px        float64
    IsPercent bool
    Percent   float64
}
```

Поля экспортируются, и структуру можно создавать напрямую. Вспомогательные конструкторы задают общепринятое представление:

- `Auto() Length` помечает значение как автоматическое.
- `Px(v float64) Length` сохраняет значение в пикселях в поле `Px`.
- `Percent(v float64) Length` помечает значение как процентное и сохраняет его числовую величину в поле `Percent`.

`Resolve(ref float64) float64` возвращает `0` для автоматического значения, `ref * Percent / 100` для процентного или `Px` в остальных случаях. `String() string` возвращает `auto`, процентное значение с суффиксом `%` или значение в пикселях с суффиксом `px`. Для форматирования чисел используется `strconv.FormatFloat` с кратчайшим десятичным представлением, которое можно преобразовать обратно без потери точности.

Поскольку структура открыта, вызывающий код может задавать конфликтующие флаги или значения. В обоих методах `IsAuto` имеет приоритет над `IsPercent`; в противном случае `IsPercent` имеет приоритет над `Px`.

## Цвет

`Color` хранит 8-битные каналы красного, зелёного и синего, а также флаг `IsSet`. `IsSet` позволяет отличить явно заданный цвет (включая чёрный) от прозрачного/не заданного.

- `Transparent() Color` возвращает нулевой, не заданный цвет.
- `RGB(r, g, b uint8) Color` возвращает заданный цвет RGB.
- `Hex(s string) Color` разбирает `#RGB` или `#RRGGBB`. Начальный `#` необязателен. Трёхзначная запись расширяется дублированием каждого символа. При неверной длине или недопустимых шестнадцатеричных цифрах возвращается `Transparent()`; разбор не зависит от регистра благодаря использованию системы счисления с основанием 16.
- `String() string` возвращает пустую строку, если цвет не задан, иначе — строчную запись `#rrggbb` с двумя шестнадцатеричными цифрами для каждого канала.

## Модель границ и блоков

`Border` объединяет число с плавающей точкой `Width`, `BorderStyle` и `Color`.

`BoxModel` содержит значения `Length` для сторон `Top`, `Right`, `Bottom` и `Left`. `IsZero() bool` возвращает `true`, только если у каждой стороны `Px == 0` и не установлен ни `IsAuto`, ни `IsPercent`. Любое ненулевое значение в пикселях или установленный флаг автоматического/процентного значения делает результат `false`.

## ComputedStyle

`ComputedStyle` — агрегат поддерживаемых вычисленных свойств. Его экспортируемые поля:

- Компоновка: `Display`, `Position`, `Overflow`, `Width`, `Height`, `MinWidth`, `MaxWidth`, `Top`, `Left`, `Right` и `Bottom`.
- Отступы блока: `Margin` и `Padding`, каждое типа `BoxModel`.
- Границы: `BorderTop`, `BorderRight`, `BorderBottom` и `BorderLeft`.
- Цвета и порядок наложения: `BackgroundColor`, `Color` и целочисленный `ZIndex`.
- Flex-компоновка: `FlexDirection`, `JustifyContent`, `AlignItems`, `FlexGrow` и `FlexShrink`.
- Grid-компоновка: `GridTemplateColumns` и `GridTemplateRows` в виде срезов `Length`.
- Интервалы и углы: `Gap` и четыре размера скругления углов `BorderRadiusTopLeft`, `BorderRadiusTopRight`, `BorderRadiusBottomRight` и `BorderRadiusBottomLeft`.

`Default() ComputedStyle` возвращает базовые значения по умолчанию, поддерживаемые пакетом:

| Поле | Значение по умолчанию |
|---|---|
| `Display` | `DisplayBlock` |
| `Position` | `PositionStatic` |
| `Overflow` | `OverflowVisible` |
| `Width`, `Height` | `Auto()` |
| `BackgroundColor` | `Transparent()` |
| `Color` | `RGB(30, 30, 30)` |
| `FlexDirection` | `FlexDirectionRow` |
| `JustifyContent` | `JustifyFlexStart` |
| `AlignItems` | `AlignStretch` |
| `FlexGrow`, `FlexShrink` | `0`, `1` |
| `Gap` | `Px(0)` |
| Остальные поля | Нулевые значения Go (включая нулевые размеры, не заданные границы/цвета, nil-срезы grid и `ZIndex == 0`) |

Возвращаемое значение — обычная структура, которую вызывающий код может изменить. `Default` и эти типы данных не выполняют проверку, каскадирование, наследование или разбор CSS.
