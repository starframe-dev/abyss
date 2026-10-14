# Specification: internal/css/types

## Обзор пакета

Пакет `css` предоставляет функциональность для парсинга CSS и вычисления стилей. Пакет работает с основными CSS-свойствами и обеспечивает создание и управление вычисленными стилями для узлов.

## Публичный API

### Типы значений

#### DisplayType

Представляет значения свойства CSS `display`.

| Константа | Значение | Описание |
|-----------|----------|----------|
| `DisplayBlock` | 0 | Блочный элемент |
| `DisplayInline` | 1 | Строковый элемент |
| `DisplayFlex` | 2 | Flex-контейнер |
| `DisplayGrid` | 3 | Grid-контейнер |
| `DisplayNone` | 4 | Элемент скрыт |

#### PositionType

Представляет значения свойства CSS `position`.

| Константа | Значение | Описание |
|-----------|----------|----------|
| `PositionStatic` | 0 | Статическое позиционирование |
| `PositionRelative` | 1 | Относительное позиционирование |
| `PositionAbsolute` | 2 | Абсолютное позиционирование |

#### OverflowType

Представляет значения свойства CSS `overflow`.

| Константа | Значение | Описание |
|-----------|----------|----------|
| `OverflowVisible` | 0 | Переполнение видно |
| `OverflowHidden` | 1 | Переполнение скрыто |

#### FlexDirection

Представляет значения свойства `flex-direction`.

| Константа | Значение | Описание |
|-----------|----------|----------|
| `FlexDirectionRow` | 0 | Строки сверху вниз |
| `FlexDirectionColumn` | 1 | Строки слева направо |

#### JustifyContent

Представляет значения свойства `justify-content`.

| Константа | Значение | Описание |
|-----------|----------|----------|
| `JustifyFlexStart` | 0 | Начальное положение |
| `JustifyCenter` | 1 | Центрирование |
| `JustifyFlexEnd` | 2 | Конечное положение |
| `JustifySpaceBetween` | 3 | Распределение пространства |
| `JustifySpaceAround` | 4 | Распределение с промежутками |

#### AlignItems

Представляет значения свойства `align-items`.

| Константа | Значение | Описание |
|-----------|----------|----------|
| `AlignFlexStart` | 0 | Начальное выравнивание |
| `AlignCenter` | 1 | Центрирование |
| `AlignFlexEnd` | 2 | Конечное выравнивание |
| `AlignStretch` | 3 | Растягивание |

#### BorderStyle

Представляет значения свойства CSS `border-style`.

| Константа | Значение | Описание |
|-----------|----------|----------|
| `BorderNone` | 0 | Граница отсутствует |
| `BorderSolid` | 1 | Прямая граница |

### Структуры данных

#### Length

Структура, представляющая CSS-значение длины.

**Поля:**
- `IsAuto bool` — true если значение `auto`
- `Px float64` — значение в пикселях
- `IsPercent bool` — true если значение в процентах
- `Percent float64` — процентное значение

**Методы:**
- `Auto() Length` — возвращает длину `auto`
- `Px(v float64) Length` — возвращает длину в пикселях
- `Percent(v float64) Length` — возвращает процентную длину
- `Resolve(ref float64) float64` — вычисляет реальное значение длины относительно опорного значения (для процентов)
- `String() string` — возвращает строковое представление значения

#### Color

Структура, представляющая RGB-цвет.

**Поля:**
- `R, G, B uint8` — компоненты цвета
- `IsSet bool` — true если цвет установлен

**Методы:**
- `Transparent() Color` — возвращает прозрачный цвет
- `RGB(r, g, b uint8) Color` — возвращает цвет из RGB-компонент
- `Hex(s string) Color` — парсит hex-строку цвета (#RRGGBB или #RGB)
- `String() string` — возвращает hex-представление цвета

#### BorderStyle

Тип для представления стиля границы (только `BorderNone` и `BorderSolid`).

#### Border

Структура, представляющая CSS-границу.

**Поля:**
- `Width float64` — ширина границы
- `Style BorderStyle` — стиль границы
- `Color Color` — цвет границы

#### BoxModel

Структура, представляющая модель盒子-элемента (top, right, bottom, left).

**Поля:**
- `Top, Right, Bottom, Left Length` — значения для каждой стороны

**Методы:**
- `IsZero() bool` — возвращает true если все стороны равны нулю

#### ComputedStyle

Структура, содержащая все вычисленные CSS-свойства для узла.

**Поля:**
- `Display DisplayType` — свойство display
- `Position PositionType` — свойство position
- `Overflow OverflowType` — свойство overflow
- `Width, Height Length` — ширина и высота
- `MinWidth, MaxWidth Length` — min-width и max-width
- `Top, Left, Right, Bottom Length` — отступы
- `Margin BoxModel` — margin
- `Padding BoxModel` — padding
- `BorderTop, BorderRight, BorderBottom, BorderLeft Border` — границы
- `BackgroundColor, Color Color` — цвета
- `ZIndex int` — z-index
- `FlexDirection, JustifyContent, AlignItems, FlexGrow, FlexShrink` — flex-свойства
- `GridTemplateColumns, GridTemplateRows []Length` — grid-шаблоны
- `Gap Length` — gap
- `BorderRadiusTopLeft, BorderRadiusTopRight, BorderRadiusBottomRight, BorderRadiusBottomLeft Length` — border-radius углы

### Вспомогательные функции

#### Default() ComputedStyle

Возвращает дефолтный вычисленный стиль.

**Дефолтные значения:**
- `Display: DisplayBlock`
- `Position: PositionStatic`
- `Overflow: OverflowVisible`
- `Width, Height: Auto()`
- `BackgroundColor: Transparent()`
- `Color: RGB(30, 30, 30)`
- `FlexDirection: FlexDirectionRow`
- `JustifyContent: JustifyFlexStart`
- `AlignItems: AlignStretch`
- `FlexGrow: 0`
- `FlexShrink: 1`
- `Gap: Px(0)`

## Важные детали реализации

### Обработка ошибок при парсинге hex-цвета

Функция `Hex` возвращает прозрачный цвет при некорректном формате hex-строки:
- Строка должна иметь длину 3 или 6 символов после удаления ведущего `#`
- Для 3-символьной строки каждый символ дублируется (#RGB -> #RRGGBB)
- При некорректной длине возвращается прозрачный цвет

### Вычисление длины

Метод `Length.Resolve`:
- Если `IsAuto` — возвращает 0
- Если `IsPercent` — вычисляет как `ref * Percent / 100`
- Иначе возвращает `Px` значение

### Определение нулевого BoxModel

Метод `BoxModel.IsZero` проверяет:
- Все `Px` компоненты равны 0
- Ни один компонент не имеет `IsAuto`
- Ни один компонент не имеет `IsPercent`

### Инкапсуляция цвета

Цвет хранится как структура с флагом `IsSet` для отличия от прозрачного цвета. Прозрачный цвет — пустая структура `Color{}`.
