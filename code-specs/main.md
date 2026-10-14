# Abyss — Спецификация проекта

## Обзор

Abyss — это TUI (текстовый интерфейс пользователя) браузер, отображающий веб-страницы в виде ASCII-графики с поддержкой CSS, JavaScript и интерактивного просмотра. Приложение использует библиотеку Bubble Tea для создания TUI.

## Публичный API

### `main()`

Запускает TUI браузер:

- Создаёт модель браузера через `browser.New()`
- Настраивает альтернативный экран
- Запускает программу Bubble Tea
- При ошибке выводит сообщение и завершает программу

### `browser.New() Model`

Создаёт новую модель браузера с:

- Поле ввода для адреса (placeholder: "Enter URL...")
- Виджет viewport (80x24 символов)
- Массив вкладок (по умолчанию одна вкладка в режиме адреса)
- Восстановлением сессии из файла `~/.abyss/session.json`

## Типы данных

### `Mode`

Перечисление режимов интерфейса:

- `ModeAddress` — ввод URL
- `ModeLoading` — загрузка страницы
- `ModeView` — просмотр страницы
- `ModeError` — ошибка загрузки

### `Tab`

Структура вкладки браузера:

- `URL` — адрес вкладки
- `Mode` — режим вкладки
- `ANSILines` — строки с ANSI-кодами
- `ASCIILines` — строки без ANSI-кодов (plain text)
- `Err` — ошибка загрузки
- `ZoomFactor` — фактор масштабирования
- `OffsetX` — смещение по горизонтали

### `Model`

Модель Bubble Tea:

- `Mode` — текущий режим
- `Address` — текст ввода URL
- `Viewport` — виджет viewport
- `Tabs` — список вкладок
- `ActiveTab` — индекс активной вкладки
- `Width` — ширина терминала
- `Height` — высота терминала
- `loadedPageContent` — загруженный контент страницы

### `pageLoadedMsg`

Сообщение о загрузке страницы:

- `Lines` — строки ASCII-отображения
- `ANSI` — строки с ANSI-кодами
- `URL` — загруженный URL

### `pageErrorMsg`

Сообщение об ошибке загрузки:

- `Err` — ошибка загрузки

### `tickMsg`

Сообщение о таймере (для периодических обновлений)

### `sessionData`

Структура для сохранения сессии:

- `URLs` — массив сохранённых URL

## Внутренние пакеты

### `internal/browser`

Реализует логику TUI браузера:

#### `loadPage(url, viewportWidth) tea.Cmd`

Загружает HTML-страницу и выполняет:

1. Фетчинг HTML через `fetcher.Fetch()`
2. Парсинг DOM через `dom.Parse()`
3. Инициализацию JS контекста
4. Регистрация JS-бриджей
5. Выполнение пользовательских скриптов
6. Вычисление стилей (CSS)
7. Построение макета (layout)
8. Отрисовку framebuffer
9. Преобразование в ASCII/ANSI

#### `reRenderCmd(viewportWidth, zoomFactor) tea.Cmd`

Перерисовывает страницу при изменении стилей или макета без перезагрузки:

- Считывает текущий DOM
- Применяет фактор масштабирования
- Пересчитывает CSS, layout, framebuffer
- Возвращает новое содержимое viewport

#### `fetchExternalCSS(doc, baseURL) []string`

Загружает внешние CSS-файлы:

- Находит теги `<link rel="stylesheet">`
- Формирует абсолютный URL
- Загружает CSS с таймаутом 2 секунды
- Возвращает до 3 стилейheets

#### `resolveURL(base, rel) string`

Формирует абсолютный URL из относительного

### `internal/fetcher`

#### `DefaultClient`

Глобальный HTTP-клиент с таймаутом 10 секунд

#### `Fetch(url) (string, error)`

Загружает HTML-контент:

- Устанавливает User-Agent Mozilla
- Проверяет HTTP 200
- Возвращает body как строку

### `internal/ascii`

#### `Render(fb, tree, viewportWidth) []string`

Преобразует framebuffer в plain text строки:

- Делит framebuffer на 2 строки терминала (1px = 2 терминальных строки)
- Накладывает текстовый контент
- Возвращает plain text

#### `RenderANSI(fb, tree, viewportWidth) []string`

Преобразует framebuffer в ANSI цветной текст:

- Генерирует ANSI-коды для фона и текста
- Следит за изменениями цветов
- Возвращает ANSI-строки

### `internal/render`

#### `Render(tree, viewportWidth) *framebuffer.Framebuffer`

Отрисовывает layout в framebuffer:

- Собирает все LayoutBox с учётом z-index
- Сдвигает координаты (исправляет negative offset)
- Вычисляет высоту framebuffer
- Очищает с цветом фона
- Рисует все коробки по z-index

### `internal/layout`

#### `Build(doc, styles, viewportWidth) *LayoutBox`

Создаёт layout-дерево из DOM:

- Игнорирует `<style>`, `<script>`, `<head>`
- Применяет CSS-стили
- Возвращает root LayoutBox

#### `Layout(box, containerWidth)`

Вычисляет геометрию layout-дерева

### `internal/framebuffer`

#### `Framebuffer`

RGB буфер пикселей:

- `Width`, `Height` — размеры
- `Pixels` — массив цветов

Методы:

- `Set(x, y, c)` — установить пиксель
- `Get(x, y)` — получить пиксель
- `Clear(c)` — очистить буфер
- `FillRect(x, y, w, h, c)` — заполнить прямоугольник
- `DrawBorder(...)` — нарисовать границы
- `DrawRoundedRect(...)` — нарисовать скруглённый прямоугольник
- `clipCorner(...)` — скруглить угол

### `internal/dom`

#### `Parse(htmlStr) *Node`

Парсит HTML и возвращает root DOM-дерево

#### `Node`

Узел DOM:

- `Type` — тип узла (Element, Text, Comment, Document)
- `TagName` — имя тега
- `Attributes` — атрибуты
- `Children` — дети
- `Text` — текст (для text nodes)
- `Parent` — родитель

Методы:

- `QueryAll(selector)` — найти все узлы по селектору
- `QuerySelector(selector)` — найти первый узел
- `Matches(selector)` — проверить совпадение селектора
- `Closest(selector)` — найти ближайшего предка по селектору
- `GetAttr(name)` — получить атрибут
- `RemoveChild(child)` — удалить ребёнка
- `AppendChild(child)` — добавить ребёнка
- `Clone(deep)` — клонировать узел
- `OuterHTML()` — сериализовать в HTML
- `FindStyleTags()` — найти `<style>` теги
- `FindLinkStylesheets()` — найти CSS линки
- `TextContent()` — получить весь текст

### `internal/css`

#### `ComputedStyle`

Вычисленные CSS-стили узла:

- `Display`, `Position`, `Overflow`
- `Width`, `Height`, `MinWidth`, `MaxWidth`
- `Top`, `Left`, `Right`, `Bottom`
- `Margin`, `Padding`
- `BorderTop/Right/Bottom/Left`
- `BackgroundColor`, `Color`
- `ZIndex`
- `FlexDirection`, `JustifyContent`, `AlignItems`
- `GridTemplateColumns/Rows`, `Gap`
- `BorderRadius*`

#### `Compute(doc) map[*dom.Node]ComputedStyle`

Вычисляет стили для всех узлов документа с кэшированием

#### `ComputeWithExternal(doc, externalCSS) map[*dom.Node]ComputedStyle`

Вычисляет стили с учётом внешних CSS-строок

### `internal/fetcher`

#### `Fetch(url) (string, error)`

Загружает HTML-страницу

## Управление состоянием

### Режимы работы

1. **ModeAddress** — ввод URL в адресной строке
2. **ModeLoading** — отображение "Loading..."
3. **ModeView** — отображение отрисованной страницы
4. **ModeError** — отображение ошибки

### Клавиши

- `Ctrl+C` / `Q` — выход
- `Enter` — загрузить URL
- `Ctrl+L` — перейти в режим адреса
- `+`, `=` — увеличить масштаб
- `-`, `_` — уменьшить масштаб
- `R` — перерисовать страницу
- `←`, `→` — горизонтальная прокрутка
- `T` — новая вкладка
- `W` — закрыть вкладку
- `Tab` / `Shift+Tab` — переключение вкладок

### Сессия

Сессия сохраняется в `~/.abyss/session.json`:

```json
{
  "urls": ["https://example.com", "https://google.com"]
}
```

## Ошибки

- Ошибки фетчинга передаются в `pageErrorMsg.Err`
- Ошибки JS-инициализации оборачиваются в `fmt.Errorf("js bridge init: %w", err)`
- Ошибки отображаются в режиме `ModeError`

## Внешние зависимости

- `github.com/charmbracelet/bubbletea` — TUI фреймворк
- `github.com/charmbracelet/bubbles/textinput` — текст инпут
- `github.com/charmbracelet/bubbles/viewport` — viewport виджет
- `github.com/charmbracelet/lipgloss` — стилизация UI
- `golang.org/x/net/html` — парсинг HTML

## Архитектурные границы

- `internal/browser/` — TUI логика (Bubble Tea model)
- `internal/dom/` — DOM парсинг и манипуляции
- `internal/css/` — CSS парсинг и вычисления
- `internal/layout/` — CSS макетирование
- `internal/render/` — отрисовка в framebuffer
- `internal/ascii/` — преобразование в ASCII/ANSI
- `internal/framebuffer/` — пиксельный буфер
- `internal/fetcher/` — HTTP фетчинг
- `internal/js/` — JS контекст и DOM API
- `main.go` — точка входа
