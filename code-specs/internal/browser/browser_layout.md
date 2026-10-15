# Спецификация: browser_layout

## Обзор

Модуль строит карту прямоугольников узлов DOM по дереву layout.

## Функции

### `buildLayoutBoxMap(root *layout.LayoutBox) map[*dom.Node]js.Rect`

Создаёт пустую карту и рекурсивно обходит layout-дерево, начиная с нулевых смещений `(0, 0)`. Для каждого box с ненулевыми `box` и `box.Node` в карту записывается прямоугольник:

- `X = box.Content.X + offsetX`
- `Y = box.Content.Y + offsetY`
- `Width = box.Content.Width`
- `Height = box.Content.Height`

Для каждого дочернего box передаются координаты текущего прямоугольника как смещения. Если корневой box или его Node равен nil, возвращается пустая карта.

### `buildLayoutBoxRecursive(box *layout.LayoutBox, offsetX, offsetY float64, result map[*dom.Node]js.Rect)`

Внутренняя вспомогательная функция. При nil `box` или `box.Node` немедленно возвращает управление и не обходит дочерние узлы. Иначе записывает прямоугольник текущего box и рекурсивно обрабатывает элементы `Children` в исходном порядке.

### `BuildLayoutBoxMapForTest(root *layout.LayoutBox) map[*dom.Node]js.Rect`

Экспортная обёртка для тестирования, возвращает результат `buildLayoutBoxMap(root)` без дополнительной обработки.

## Типы и зависимости

Используются `dom.Node`, `js.Rect` и `layout.LayoutBox` из соответствующих внутренних пакетов проекта. `js.Rect` заполняется полями `X`, `Y`, `Width`, `Height`.

## Свойства обхода

При обходе дерева из n посещённых box время работы составляет O(n); дополнительная память — O(n) для карты и O(h) для стека рекурсии, где h — глубина дерева. Предполагается отсутствие циклов в `Children`. Если разные box ссылаются на один `*dom.Node`, более поздняя запись заменяет предыдущую.
