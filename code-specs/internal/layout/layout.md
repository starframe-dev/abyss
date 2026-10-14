# Specification: internal/layout/layout.go

## Обзор пакета

Пакет `layout` реализует движок вычисления положений и размеров DOM-узлов по модели CSS box model. Пакет предназначен для рендеринга HTML-документов в текстовом терминале с использованием блоков `█▄▀` (терминальная графика).

## Основные типы

### `ViewportWidth`

```go
const ViewportWidth = 800
```

Ширина вьюпорта по умолчанию в пикселях (800px).

### `Rect`

```go
type Rect struct {
    X, Y     float64 // координаты левого верхнего угла
    Width    float64 // ширина прямоугольника
    Height   float64 // высота прямоугольника
}
```

Представляет прямоугольник с позицией и размером. Используется для описания content, padding, border, margin областей LayoutBox.

### `TextLine`

```go
type TextLine struct {
    Text string // текст строки
    X, Y float64 // координаты в терминале
}
```

Представляет строку текста с её позицией в терминальном рендеринге.

### `LayoutBox`

```go
type LayoutBox struct {
    Node      *dom.Node     // ссылка на узел DOM
    Parent    *LayoutBox    // родительский LayoutBox (если есть)
    Style     css.ComputedStyle // вычисленный стиль узла
    Content   Rect          // content area (относительно parent content)
    Padding   Rect          // padding area
    Border    Rect          // border area
    Margin    Rect          // margin area
    Children  []*LayoutBox  // дочерние LayoutBox
    TextLines []TextLine    // строки текста для inline-контента
}
```

Представляет вычисленную геометрию элемента. Каждая область (Content, Padding, Border, Margin) имеет свои прямоугольные границы.

## Публичный API

### `Build`

```go
func Build(doc *dom.Node, styles map[*dom.Node]css.ComputedStyle, viewportWidth float64) *LayoutBox
```

Создает дерево layout из DOM-документа и вычисленных стилей.

**Параметры:**
- `doc` — корневой узел DOM-документа
- `styles` — карта узлов на их вычисленные стили
- `viewportWidth` — ширина вьюпорта

**Возвращает:**
- Корневой `*LayoutBox` или `nil`, если дерево не может быть построено

**Поведение:**
1. Проходит рекурсивно по всем узлам DOM
2. Исключает элементы `style`, `script`, `head` — они не создают LayoutBox
3. Если `display: none` — игнорирует элемент
4. Если узел `TextNode` — пропускает (текст обрабатывается позже)
5. Строит дерево LayoutBox
6. Вызывает `Layout()` для вычисления геометрии

**Пример:**
```go
root := Build(doc, styles, 800)
```

### `Layout`

```go
func Layout(box *LayoutBox, containerWidth float64)
```

Вычисляет геометрию дерева layout.

**Параметры:**
- `box` — корневой LayoutBox
- `containerWidth` — ширина контейнера (в пикселях)

**Поведение:**
1. Вычисляет ширину контента с учётом padding, border
2. Обработчики display: Flex, Grid, Block, Inline
3. Вычисляет высоту контента
4. Заполняет поля Content, Padding, Border, Margin
5. Обрабатывает текстовый контент с переносом строк

## Внутренние функции

### `buildNode`

```go
func buildNode(node *dom.Node, styles map[*dom.Node]css.ComputedStyle) *LayoutBox
```

Создаёт LayoutBox для одного DOM-узла.

**Логика:**
1. Возвращает `nil` для `TextNode`
2. Возвращает `nil` для `style`, `script`, `head` элементов
3. Проверяет `display: none`
4. Создаёт новый LayoutBox
5. Рекурсивно строит дочерние узлы
6. Устанавливает связи Parent-Children

### `layoutBox`

```go
func layoutBox(box *LayoutBox, containerWidth, offsetX, offsetY float64, isRoot bool)
```

Вычисляет геометрию одного LayoutBox.

**Параметры:**
- `box` — LayoutBox для вычисления
- `containerWidth` — ширина контейнера
- `offsetX`, `offsetY` — смещение (margin-box origin)
- `isRoot` — флаг корневого узла

**Поведение:**
1. Вычисляет ширину через `resolveWidth()`
2. Извлекает значения margin, padding, border
3. Вычисляет width content area
4. Вычисляет позицию для positioned элементов (absolute/relative)
5. Устанавливает Rect для Content, Padding, Border, Margin
6. Вызывает соответствующий layout-функцию в зависимости от display
7. Вычисляет высоту через `resolveHeight()`

### `resolveWidth`

```go
func resolveWidth(style css.ComputedStyle, containerWidth float64) float64
```

Вычисляет ширину элемента.

**Поведение:**
1. Если `width` задан — возвращает вычисленное значение
2. Если `width: auto` — заполняет контейнер минус margins
3. Защита от отрицательных значений

### `resolveHeight`

```go
func resolveHeight(style css.ComputedStyle, contentHeight float64) float64
```

Вычисляет высоту элемента.

**Поведение:**
1. Если `height` задан — возвращает вычисленное значение
2. Иначе возвращает `contentHeight`

### `layoutBlockFlow`

```go
func layoutBlockFlow(box *LayoutBox, contentWidth float64) float64
```

Обработка элементов с `display: block`.

**Поведение:**
1. Проходит по дочерним элементам вертикально
2. Суммирует высоты с margin
3. Если нет детей — делегирует в `layoutInlineFlow()`

### `layoutInlineFlow`

```go
func layoutInlineFlow(box *LayoutBox, contentWidth float64) float64
```

Обработка элементов с `display: inline` и текстового контента.

**Поведение:**
1. Проверяет, находится ли в `<style>` или `<script>`
2. Собирает inline-текст через `collectInlineText()`
3. Преобразует текст в слова через `strings.Fields()`
4. Выполняет word wrap через `wordWrap()`
5. Возвращает высоту контента

### `layoutFlex` / `layoutGrid`

Функции для layout flex и grid (реализация отсутствует в коде).

### `layoutBlockFlow` / `layoutInlineFlow`

Переключатели для display properties.

### `collectInlineText`

```go
func collectInlineText(n *dom.Node) []string
```

Собирает текстовый контент из inline-элементов.

**Поведение:**
1. Пропускает `<style>` и `<script>`
2. Рекурсивно собирает текст из TextNode
3. Убирает лишние пробелы

### `wordWrap`

```go
func wordWrap(words []string, box *LayoutBox, contentWidth float64) []TextLine
```

Выполняет перенос строк для текстового контента.

**Параметры:**
- `words` — список слов
- `box` — LayoutBox (для доступа к координатам)
- `contentWidth` — доступная ширина

**Поведение:**
1. `lineHeight = 2.0` (1 терминальная строка = 2 CSS пикселя)
2. `wordWidth = len(word)` (1 пиксель на символ)
3. Пробелы занимают 1 пиксель
4. Если строка не влезает — создаёт новую строку
5. Возвращает срезку TextLine

**Примечание:** Использует `strings.Fields()` — оригинальный HTML уже имеет правильные пробелы.

### `inStyleOrScript`

```go
func inStyleOrScript(n *dom.Node) bool
```

Проверяет, находится ли узел внутри `<style>` или `<script>`.

### `collectVisibleText`

```go
func collectVisibleText(n *dom.Node) string
```

Возвращает видимый текстовый контент, игнорируя `<style>` и `<script>`.

### `findPositionedAncestor`

```go
func findPositionedAncestor(box *LayoutBox) *LayoutBox
```

Находит ближайшего родителя с `position: relative` или `absolute`.

### `estimateTextWidth`

```go
func estimateTextWidth(text string) float64
```

Оценивает ширину текста в пикселях (1 пиксель на символ).

## Модель Box Model

Каждый элемент имеет 4 области:

```
┌─────────────────────────────────────────┐
│  Margin                                 │
│ ┌─────────────────────────────────────┐ │
│ │   Border                             │ │
│ │ ┌─────────────────────────────────┐ │ │
│ │ │   Padding                       │ │ │
│ │ │ ┌─────────────────────────────┐ │ │ │
│ │ │ │   Content                    │ │ │ │
│ │ │ └─────────────────────────────┘ │ │ │
│ │ └─────────────────────────────────┘ │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

### Порядок вычислений

1. Content width вычисляется из ширины минус padding, border
2. Padding расширяется внутрь (не накладывается на предшествующие siblings)
3. Border окружает padding
4. Margin окружает border

### Root vs Non-root

- Root элемент: margin прижат к 0, border берётся от marginLeft/marginTop
- Non-root элемент: margin считается от родителя

## Positioned Elements

При `position: absolute` или `position: relative`:

1. Находит positioned ancestor (ближайшего родителя с position)
2. Вычисляет позицию относительно контента ancestor:
   - `top` относительно `refY + ancestor.Content.Height`
   - `left` относительно `refX + ancestor.Content.Width`
   - `right`/`bottom` вычисляется зеркально

## Текстовый рендеринг

### Шкала

- 1 терминальная строка = 2 CSS пикселя
- 1 терминальный столбец = 1 CSS пиксель
- Блоки `█▄▀` используются для графического рендеринга

### Word Wrap

- Словоразбивание через `strings.Fields()`
- Пробел занимает 1 пиксель
- Перенос при превышении `contentWidth`
- Каждая строка TextLine сохраняет координаты X, Y

## Исключения

Элементы не создаются в следующих случаях:

1. `node.Type == dom.TextNode` — текст обрабатывается отдельно
2. `node.TagName` в `["style", "script", "head"]` — стили и скрипты игнорируются
3. `style.Display == css.DisplayNone` — скрытые элементы

## Зависимости

- `github.com/starframe-dev/abyss/internal/css` — вычисленные стили, margin/padding/border resolution
- `github.com/starframe-dev/abyss/internal/dom` — типы DOM, узлы, текст

## Важные примечания

1. **Root layout**: `Build()` вызывает `Layout()` только если root != nil
2. **Auto width/height**: Auto-значения вычисляются относительно контейнера
3. **Inline text**: Текст внутри block-элементов обрабатывается как inline-flow
4. **Flex/Grid**: Функции `layoutFlex` и `layoutGrid` объявлены, но реализация отсутствует
5. **TextLines**: Заполняется только для элементов с inline-текстом
6. **Positioned**: Если ancestor не найден — элемент позиционируется относительно viewport
7. **Error handling**: Нет явной обработки ошибок — предполагается корректный вход

## Схема потока данных

```
Build()
  └─> buildNode() (построение дерева LayoutBox)
       └─> Layout() (вычисление геометрии)
            └─> layoutBox()
                 ├─> resolveWidth() / resolveHeight()
                 ├─> findPositionedAncestor()
                 ├─> layoutBlockFlow() / layoutInlineFlow()
                 │    ├─> collectInlineText()
                 │    └─> wordWrap()
                 └─> resolveHeight()
```

## Пример использования

```go
root := Build(doc, styles, 800)
if root != nil {
    Layout(root, 800)
    // Теперь root.Content, root.Padding, root.Border, root.Margin
    // и root.TextLines заполнены
}
```
