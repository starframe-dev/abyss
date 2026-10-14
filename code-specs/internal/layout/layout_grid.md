# layout_grid

## Обзор

Функция `layoutGrid` вычисляет позиции и размеры узлов DOM, используя CSS box model для сеточной верстки (CSS Grid Layout). Она принимает `LayoutBox` (контейнер) и `contentWidth` (ширина контента) и возвращает общую высоту сетки.

## Публичный API

### layoutGrid

```go
func layoutGrid(box *LayoutBox, contentWidth float64) float64
```

Вычисляет общую высоту сетки для заданного контейнера.

**Параметры:**

- `box` — указатель на `LayoutBox`, представляющий контейнер сетки со стилями, дочерними элементами и координатами.
- `contentWidth` — ширина контента сетки (в пикселях).

**Возвращаемое значение:**

- `float64` — общая высота сетки с учётом отступов (gap) между строками.

**Поведение:**

1. Получает стиль контейнера и вычисляет отступ `gap` через `style.Gap.Resolve(contentWidth)`.
2. Инициирует `GridTemplateColumns`: если массив пуст, создаёт один столбец на всю ширину контента.
3. Инициирует `GridTemplateRows` со значением 100px по умолчанию, если массив пуст.
4. Вызывает `resolveGridColumnWidths` для вычисления ширин столбцов.
5. Вызывает `placeGridChildren` для размещения дочерних элементов и получения высот строк.
6. Вызывает `sumRowHeights` для вычисления итоговой высоты.

## Типы

### LayoutBox

Указатель на структуру `LayoutBox` из пакета `internal/css`, содержащая:

- `Style` — стиль контейнера (свойства grid-template-columns, grid-template-rows, gap и т.д.).
- `Children` — массив дочерних элементов сетки.
- `Content` — координаты контента (`X`, `Y`).

### css.Length

Тип из пакета `internal/css`, представляющий длину с различными единицами измерения:

- `Px` — значение в пикселях.
- `IsPercent` — флаг, что длина задана в процентах.
- `IsAuto` — флаг, что длина равна auto.
- `Resolve(width)` — возвращает значение в пикселях, основываясь на ширине контента.

## Внутренние функции

### resolveGridColumnWidths

```go
func resolveGridColumnWidths(columns []css.Length, contentWidth, gap float64) []float64
```

Вычисляет ширину каждого столбца сетки на основе `GridTemplateColumns`.

**Алгоритм:**

1. Инициализирует массив `colWidths` и переменные `totalFixedWidth` (сумма фиксированных ширин) и `totalFr` (сумма значений fr).
2. Проходит по каждому столбцу:
   - Если `IsPercent` — вычисляет ширину в пикселях через `col.Resolve(contentWidth)` и добавляет к `colWidths` и `totalFixedWidth`.
   - Если не `IsAuto` — использует `col.Px` как ширину, добавляет к `colWidths` и `totalFixedWidth`.
   - Если `IsAuto` (fr unit) — добавляет 0 к `colWidths` и накапливает `col.Px` в `totalFr`.
3. Вычисляет `remainingWidth` как `contentWidth - totalFixedWidth - gap * (len(columns) - 1)`.
4. Если `remainingWidth < 0`, устанавливает в 0.
5. Для всех столбцов с `IsAuto` и `totalFr > 0` вычисляет ширину как `remainingWidth * col.Px / totalFr`.
6. Возвращает `colWidths`.

**Замечания:**

- Fr-сегменты получают оставшуюся ширину пропорционально их значениям.
- Отступы между столбцами не учитываются при вычислении `remainingWidth` для fr-элементов (возможно, баг, но не исправлен).

### placeGridChildren

```go
func placeGridChildren(box *LayoutBox, columns, rows []css.Length, colWidths []float64, gap, contentWidth float64) []float64
```

Размещает дочерние элементы в ячейках сетки и возвращает массив высот строк.

**Алгоритм:**

1. Инициализирует `rowHeights` как пустой массив.
2. Проходит по каждому дочернему элементу:
   - Вычисляет индекс столбца: `col = i % numCols`.
   - Вычисляет индекс строки: `row = i / numCols`.
   - Убедится, что массив `rowHeights` имеет достаточно элементов для текущей строки, используя `rows[len(rowHeights)%len(rows)].Resolve(contentWidth)` как запасную высоту строки.
   - Вычисляет координату X через `gridOffsetX`.
   - Вычисляет координату Y через `gridOffsetY`.
   - Сохраняет оригинальную ширину дочернего элемента.
   - Устанавливает ширину дочернего элемента равной `colWidths[col]`.
   - Вызывает `layoutBox` для позиционирования и расчёта размеров дочернего элемента.
   - Восстанавливает оригинальную ширину дочернего элемента.
   - Если `child.Margin.Height > rowHeights[row]`, обновляет высоту строки.
3. Возвращает `rowHeights`.

**Замечания:**

- Массив `rowHeights` расширяется динамически, если нужно больше строк, чем определено в `GridTemplateRows`.
- При вычислении запасной высоты используется `rows[len(rowHeights)%len(rows)]` — возможно, для циклического доступа к строкам по умолчанию.

### gridOffsetX

```go
func gridOffsetX(base float64, colWidths []float64, gap float64, col int) float64
```

Вычисляет X-координату начала столбца.

**Формула:**

```
x = base
for c := 0 to col-1:
    x = x + colWidths[c] + gap
return x
```

### gridOffsetY

```go
func gridOffsetY(base float64, rowHeights []float64, gap float64, row int) float64
```

Вычисляет Y-координату начала строки.

**Формула:**

```
y = base
for r := 0 to row-1:
    y = y + rowHeights[r] + gap
return y
```

### sumRowHeights

```go
func sumRowHeights(rowHeights []float64, gap float64) float64
```

Вычисляет общую высоту сетки.

**Формула:**

```
totalHeight = 0
for h in rowHeights:
    totalHeight = totalHeight + h + gap
if len(rowHeights) > 0:
    totalHeight = totalHeight - gap
return totalHeight
```

**Замечание:**

- Отступ `gap` добавляется между каждой парой строк, но не для первой строки.
- Формула эквивалентна: `sum(rowHeights) + (len(rowHeights) - 1) * gap`.

## Поведение

- **Grid-верстка:** Поддерживает CSS Grid Layout с гибким распределением ширины между столбцами (auto, percent, fr).
- **Динамические строки:** Строки могут расширяться динамически, если их не было достаточно в определении `GridTemplateRows`.
- **Высота строк:** Высота строки определяется максимальным `Margin.Height` среди всех элементов в этой строке.
- **Позиционирование:** Дочерние элементы позиционируются по умолчанию с расчётом отступов между ячейками.

## Ограничения

- Fr-сегменты получают оставшуюся ширину пропорционально, но отступы между столбцами не учитываются при вычислении `remainingWidth`.
- Если `remainingWidth < 0`, он обнуляется, что может привести к сжатию контента.
- Запасная высота строки вычисляется циклически из `rows` массива, что может не соответствовать задуманному поведению.
- Ширина дочернего элемента временно изменяется и восстанавливается, что может быть неоптимально для производительности.

## Зависимости

- `github.com/starframe/abyss/internal/css` — предоставляет типы `LayoutBox` и `css.Length`.

## Применимость

Функция используется для вычисления сеточной верстки в CSS box model, где контейнер имеет свойство `display: grid` с заданными или вычисленными значениями `grid-template-columns`, `grid-template-rows` и `gap`.
