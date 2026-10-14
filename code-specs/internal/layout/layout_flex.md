# Specification: layout_flex.go

## Обзор

Файл `layout_flex.go` реализует алгоритм расчёта позиций и размеров элементов Flexbox в CSS Box Model. Функции находятся в пакете `layout` и работают с элементами, у которых `display: flex` или `display: -webkit-flex`.

---

## Публичный API

В пакете `layout` нет публичных функций, которые начинаются с большой буквы. Все функции `layoutFlex`, `flexAvailableSpace`, `flexTotals`, `flexRemaining`, `layoutFlexChild`, `flexContainerSize` являются внутренними.

Функции работают с типом `*LayoutBox` из пакета `layout` и вычисляют размеры флекс-контейнера, возвращая результат как `float64`.

---

## Типы и структура данных

### LayoutBox

Тип `LayoutBox` определяет расчитанную геометрию элемента DOM:

- `Node *dom.Node` — ссылка на исходный узел DOM
- `Parent *LayoutBox` — родительский элемент в дереве layout
- `Style css.ComputedStyle` — вычисленные CSS стили
- `Content Rect` — область контента (относительно области контента родителя)
- `Padding Rect` — область padding
- `Border Rect` — область border
- `Margin Rect` — область margin
- `Children []*LayoutBox` — дочерние элементы
- `TextLines []TextLine` — строки текста (для инлайн контента)

### Rect

Тип `Rect` представляет прямоугольник:

```go
type Rect struct {
    X, Y, Width, Height float64
}
```

### LayoutBox (поля)

- `Content` — область контента
- `Padding` — область padding
- `Border` — область border
- `Margin` — область margin
- `Children` — массив дочерних LayoutBox
- `TextLines` — массив строк текста

### TextLine

Тип `TextLine` представляет строку текста:

```go
type TextLine struct {
    Text string
    X, Y float64
}
```

---

## Функции и поведение

### layoutFlex(box *LayoutBox, contentWidth float64) float64

**Назначение:** Вычисляет размеры и позиции всех элементов внутри флекс-контейнера.

**Входные параметры:**
- `box` — флекс-контейнер (элемент с `display: flex`)
- `contentWidth` — ширина контента флекс-контейнера (в пикселях)

**Возвращаемое значение:**
- Вернёт высоту флекс-контейнера (максимальную высоту среди детей в column-режиме, или 0 в row-режиме)
- Вызывает `layoutBox` для каждого дочернего элемента, устанавливая их позиции и размеры

**Алгоритм:**
1. Получает стиль контейнера и вычисляет `gap` (отступ между элементами)
2. Если детей нет — возвращает 0
3. Вычисляет общее пространство `availableSpace` (зависит от `flex-direction`)
4. Вычисляет суммы `basis`, `grow`, `shrink` всех детей через `flexTotals`
5. Вычисляет `remaining` (свободное пространство для распределения)
6. Проходит по всем детям и вызывает `layoutFlexChild` для каждого
7. Возвращает итоговый размер контейнера через `flexContainerSize`

### flexAvailableSpace(flexDir css.FlexDirection, contentWidth, totalGap float64) float64

**Назначение:** Вычисляет доступное пространство в зависимости от направления флекса.

**Поведение:**
- Если `flexDir == FlexDirectionRow`: возвращает `contentWidth - totalGap`
- Если `flexDir == FlexDirectionColumn`: возвращает большое число `100000` (используется для вычисления высоты в column-режиме)

### flexTotals(box *LayoutBox, contentWidth float64) (basis, grow, shrink float64)

**Назначение:** Вычисляет суммы `width`, `flex-grow`, `flex-shrink` всех детей.

**Поведение:**
- Проходит по всем детям
- Для каждого ребёнка вычисляет `basis = child.Style.Width.Resolve(contentWidth)`
- Если `basis == 0`, устанавливает дефолтное значение `100`
- Суммирует `basis`, `grow`, `shrink`
- Возвращает три значения

### flexRemaining(flexDir css.FlexDirection, availableSpace, totalBasis float64) float64

**Назначение:** Вычисляет оставшееся пространство после вычета базовых размеров.

**Поведение:**
- Если `flexDir == FlexDirectionRow`: возвращает `availableSpace - totalBasis`
- Если `flexDir == FlexDirectionColumn`: возвращает `0`

### layoutFlexChild(box, child *LayoutBox, flexDir css.FlexDirection, gap, contentWidth, pos, remaining, totalGrow float64) float64

**Назначение:** Вычисляет размер и позицию одного дочернего элемента флекса.

**Поведение:**
1. Вычисляет `basis = child.Style.Width.Resolve(contentWidth)`
2. Если `basis == 0`, устанавливает дефолтное значение `100`
3. Вычисляет `childWidth = basis`
4. Если `totalGrow > 0 && remaining > 0`: добавляет растяжение: `childWidth += remaining * child.Style.FlexGrow / totalGrow`
5. Временно устанавливает `child.Style.Width = css.Px(childWidth)`
6. Вызывает `layoutBox(child, childWidth, ...)` для установки позиций
   - Если `flexDir == FlexDirectionRow`:
     - X = `box.Content.X + pos`
     - Y = `box.Content.Y`
     - Обновляет `pos += childWidth + gap`
   - Если `flexDir == FlexDirectionColumn`:
     - X = `box.Content.X`
     - Y = `box.Content.Y + pos`
     - Обновляет `pos += child.Margin.Height + gap`
7. Восстанавливает оригинальный `child.Style.Width`
8. Возвращает обновлённый `pos`

### flexContainerSize(box *LayoutBox, flexDir css.FlexDirection, pos, gap float64) float64

**Назначение:** Вычисляет итоговый размер контейнера.

**Поведение:**
- Если `flexDir == FlexDirectionRow`:
  - Проходит по всем детям
  - Находит максимальную `Margin.Height` среди детей
  - Возвращает `maxHeight`
- Если `flexDir == FlexDirectionColumn`:
  - Возвращает `pos - gap`

---

## Взаимодействие с другими модулями

### Пакет css

Функция использует типы из пакета `css`:
- `css.FlexDirection` — направление флекса (row/column)
- `css.ComputedStyle` — вычисленные CSS стили
- `css.Px(float64)` — длина в пикселях
- `css.Length.Resolve(float64)` — разрешение длины относительно референса

### Пакет layout

Использует тип `*LayoutBox` из пакета `layout`:
- `LayoutBox.Style` — стили
- `LayoutBox.Children` — дети
- `LayoutBox.Content` — прямоугольник контента
- `LayoutBox.Margin` — margin прямоугольник

---

## Важные детали реализации

### Дефолтные значения

- Если `width` у элемента не задан, `basis` берётся из `Resolve(contentWidth)`
- Если `basis == 0`, используется дефолтное значение `100` пикселей
- Для column-режима используется большое число `100000` как placeholder

### Временное изменение стилей

В функции `layoutFlexChild` временно устанавливается `child.Style.Width` перед вызовом `layoutBox`, а затем восстанавливается оригинальное значение. Это важно, чтобы не нарушить состояние стилей после расчёта.

### Распределение пространства (flex-grow/shrink)

Распределение свободного пространства происходит только если:
- `totalGrow > 0` (есть элементы с `flex-grow > 0`)
- `remaining > 0` (есть свободное пространство)

Формула: `childWidth += remaining * child.Style.FlexGrow / totalGrow`

### Обработка gap

- `gap` берётся из `style.Gap.Resolve(contentWidth)`
- В row-режиме: `totalGap = gap * (numChildren - 1)`
- В column-режиме: gap не используется в расчёте доступного пространства

---

## Тесты

Файл содержит тесты в `layout_flex_test.go`:

- `TestLayout_FlexRow` — тест для row-направления
- `TestLayout_FlexColumn` — тест для column-направления
- `TestLayout_Grid` — тест для grid-раскладки
- `TestLayout_Absolute` — тест для positioned элементов

Тесты проверяют:
- Размеры контейнера
- Позиции детей
- Распределение `flex-grow`

---

## Примечания

1. Функция возвращает размер контейнера, но не сохраняет его в поле `box.Content.Height` — это происходит в вызывающей функции `layoutBox`.

2. В column-режиме функция возвращает `pos - gap`.

3. Дефолтное значение `100` для `basis` при `width == 0` может влиять на поведение элементов без заданной ширины.

4. Временное изменение `child.Style.Width` следует учитывать при отладке — изменения должны быть обратимы.
