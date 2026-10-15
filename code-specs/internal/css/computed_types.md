# Спецификация: `internal/css/computed_types.go`

## Назначение

Файл определяет типы и конструкторы вычисленных CSS-значений. Парсеры строк находятся в `computed_values.go`.

## Перечисления

Определены целочисленные типы и значения по умолчанию через `iota`:

- `DisplayType`: `DisplayBlock`, `DisplayInline`, `DisplayInlineBlock`, `DisplayFlex`, `DisplayGrid`, `DisplayNone`.
- `PositionType`: `PositionStatic`, `PositionRelative`, `PositionAbsolute`.
- `TextAlignType`: `TextAlignLeft`, `TextAlignCenter`, `TextAlignRight`; наследование не моделируется.
- `OverflowType`: `OverflowVisible`, `OverflowHidden`.
- `FlexDirection`: `FlexDirectionRow`, `FlexDirectionColumn`.
- `JustifyContent`: `JustifyFlexStart`, `JustifyCenter`, `JustifyFlexEnd`, `JustifySpaceBetween`, `JustifySpaceAround`.
- `AlignItems`: `AlignFlexStart`, `AlignCenter`, `AlignFlexEnd`, `AlignStretch`.
- `BorderStyle`: `BorderNone`, `BorderSolid`.

## Значения и конструкторы

`Length` хранит признаки auto, процента и дробного grid-размера и соответствующие значения (`Px`, `Percent`, `Fraction`). `Auto()`, `Px(v)`, `Percent(v)` и `Fr(v)` создают длины соответствующих видов. `Resolve(ref)` возвращает 0 для auto и fraction, процент от `ref` для percent и `Px` в остальных случаях. `String()` сериализует auto как `auto`, а остальные варианты с двумя знаками после запятой и суффиксами `%`, `fr` или `px`.

`Color` содержит компоненты RGB типа `uint8` и признак `IsSet`. `Transparent()` возвращает незаданный цвет, `RGB(r,g,b)` — заданный. `String()` возвращает `transparent` для незаданного цвета либо шестнадцатеричную строку `#rrggbb` в нижнем регистре.

`Border` содержит ширину, стиль и цвет. `BoxModel` содержит длины верхней, правой, нижней и левой сторон. `IsZero()` возвращает true, только если каждая сторона — нулевая длина в пикселях (не auto, процент или fraction).

## Вычисленный стиль

`ComputedStyle` содержит display, position, overflow, размеры и ограничения ширины, координаты, margin/padding, четыре границы, фон и цвет текста, z-index, выравнивание текста, flex-настройки, grid-шаблоны, gap и четыре радиуса углов границы.

`Default()` возвращает стиль со следующими заданными значениями: `DisplayBlock`, `PositionStatic`, `OverflowVisible`, ширина и высота `Auto()`, прозрачный фон, цвет текста `RGB(30,30,30)`, направление flex `FlexDirectionRow`, `JustifyFlexStart`, `AlignStretch`, `FlexGrow` 0, `FlexShrink` 1 и `Gap` `Px(0)`. Остальные поля остаются нулевыми значениями Go.
