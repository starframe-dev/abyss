# Specification: cmd/manual/main.go

## Обзор

Файл `cmd/manual/main.go` реализует CLI-утилиту для демонстрации pipeline рендеринга HTML-страницы с inline JavaScript в ASCII-выходе. Программа создаёт тестовую HTML-страницу, парсит её в DOM, исполняет JavaScript-код, применяет CSS-стилизацию, вычисляет верстку и рендерит результат в виде ASCII-картинки.

## Вводные данные

- **Входные данные**: Константа `testPage` — строка с полным HTML-документом.
- **Выходные данные**: ASCII-текст, выведенный через `fmt.Println` (стандартный выход `stdout`).
- **Сторонние эффекты**:
  - Создание временного файла в `os.TempDir()`.
  - Удаление временного файла через `defer`.
  - Вывод `fmt.Println` и `fmt.Printf` (stdout, stderr).
  - Sleep на 2 секунды в конце.

## Публичный API

### Функция `main()`

**Цель**: Точка входа в программу.

**Последовательность действий**:
1. Формирует путь к временному HTML-файлу (`filepath.Join(os.TempDir(), "abyss_test_page.html")`).
2. Записывает содержимое `testPage` в файл через `os.WriteFile`.
3. Планирует удаление файла через `defer os.Remove(filePath)`.
4. Выводит заголовок и путь к файлу.
5. Парсит HTML через `dom.Parse(testPage)`.
6. Создаёт контекст JavaScript через `js.NewContext()`.
7. Инициализирует DOM в контексте JS через `js.InitDOM(doc)`.
8. Настраивает writer для логов JS на `os.Stderr`.
9. Регистрирует мостовые функции через `jsc.RegisterBridgeFunctions()`.
10. Инициализирует DOM API через `jsc.Evaluate(js.DOMAPIScript())`.
11. Выполняет inline JS-скрипты через `js.ExecuteScripts(jsc, doc, "")`.
12. Получает финальный DOM через `js.GetFinalDOM()`.
13. Если финальный DOM не nil, выводит количество детей корня и вызывает `printDOMTree`.
14. Вычисляет CSS-стили через `css.Compute(finalDOM)`.
15. Строит layout через `layout.Build(finalDOM, styles, 800)`.
16. Рендерит framebuffer через `render.Render(box, 800)`.
17. Конвертирует framebuffer в ASCII-строки через `ascii.Render(fb, box, 800)`.
18. Выводит размер framebuffer и количество строк.
19. Выводит все ASCII-строки с префиксом `=== OUTPUT ===`.
20. Спит 2 секунды для наблюдения пользователем.

**Обработка ошибок**:
- При записи файла — `panic(err)`.
- При инициализации bridge — печать ошибки и `os.Exit(1)`.
- При инициализации DOM API — печать ошибки и `os.Exit(1)`.
- При выполнении JS — печать ошибки, но продолжение.

**Внутренние функции**:

#### `printDOMTree(root *dom.Node)`

**Цель**: Рекурсивно выводит дерево DOM-элементов.

**Логика**:
- Проходит по дочерним узлам корня.
- Для элементов (`dom.ElementNode`) выводит тег и количество детей.
- Если тег `"html"`, вызывает `printHTMLBody`.

#### `printHTMLBody(htmlEl *dom.Node)`

**Цель**: Выводит содержимое тега `<body>` и его детей.

**Логика**:
- Фильтрует только элементы `<body>`.
- Для элементов выводит тег, атрибут `class` и `TextContent()`.
- Для текстовых узлов выводит обрезанный текст (до 40 символов + "...").

## Типы

### Константа `testPage`

**Тип**: `string`.

**Описание**: Полная HTML-страница с inline JavaScript и CSS. Содержит:
- `<style>` для flexbox, grid, цветовых блоков, карточек.
- `<script>` с тестами на:
  - `__go_log` (bridge-функция для вывода в stderr).
  - `document.createElement`, `appendChild`.
  - `className`, `classList.add`.
  - `style` (backgroundColor, color, padding, borderRadius).
  - `setAttribute`/`getAttribute`.
  - `parentNode`, `children`, `firstChild`, `lastChild`.
  - `innerHTML`.
  - `document.write`.
  - `getElementById`.
  - `getElementsByTagName`.
  - `removeChild`.
- `<body>` с HTML-структурой:
  - Flex-row с красным, зелёным, синими блоками.
  - Grid-контейнер с фиолетовым и жёлтым блоками.
  - Flex-row с карточками (red + green + текст).

### Внешние зависимости

- `github.com/starframe-dev/abyss/internal/ascii` — рендеринг framebuffer в ASCII.
- `github.com/starframe-dev/abyss/internal/css` — вычисление стилей из DOM.
- `github.com/starframe-dev/abyss/internal/dom` — парсинг и представление DOM.
- `github.com/starframe-dev/abyss/internal/js` — контекст выполнения JavaScript, мостовые функции.
- `github.com/starframe-dev/abyss/internal/layout` — вычисление box-model верстки.
- `github.com/starframe-dev/abyss/internal/render` — рендеринг framebuffer из layout.

## Алгоритм

1. **Подготовка**:
   - Создание временного HTML-файла.
   - Вывод заголовка.

2. **Парсинг**:
   - `dom.Parse(testPage)` → `*dom.Node` (корневой элемент).
   - Вывод количества детей корня.

3. **Инициализация JS**:
   - `js.NewContext()` → `*js.Context`.
   - `js.InitDOM(doc)`.
   - `jsc.RegisterBridgeFunctions()`.
   - `jsc.Evaluate(js.DOMAPIScript())`.

4. **Выполнение JS**:
   - `js.ExecuteScripts(jsc, doc, "")`.
   - Получение финального DOM через `js.GetFinalDOM()`.

5. **Анализ DOM**:
   - Если финальный DOM не nil:
     - `len(finalDOM.Children)` → вывод количества.
     - `printDOMTree(finalDOM)`.

6. **Рендеринг pipeline**:
   - `css.Compute(finalDOM)` → `[]css.Style`.
   - `layout.Build(finalDOM, styles, 800)` → `layout.Box`.
   - `render.Render(box, 800)` → `render.Framebuffer`.
   - `ascii.Render(fb, box, 800)` → `[]string`.

7. **Вывод**:
   - `fb.Width`, `fb.Height`.
   - `len(lines)`.
   - Цикл по строкам с `fmt.Println`.

8. **Завершение**:
   - `time.Sleep(2 * time.Second)`.
   - Конец программы.

## Инварианты

- Временный файл всегда удаляется через `defer`.
- `os.Exit(1)` вызывается только при критических ошибках (bridge, DOM API).
- Non-critical ошибки (JS execution) не останавливают программу.
- Финальный DOM может быть nil (в этом случае вывод пропускается).

## Побочные эффекты

- Создание/удаление временного файла.
- Вывод на `stdout` и `stderr`.
- Sleep на 2 секунды.

## Ограничения

- Размер экрана фиксирован: 800px.
- Временный файл хранится в `os.TempDir()`.
- Нет проверки валидности `testPage` перед парсингом (полагается на корректность константы).

## Сценарии использования

- Демонстрация pipeline рендеринга HTML → DOM → JS → CSS → Layout → ASCII.
- Тестирование inline JavaScript в контролируемой среде.
- Визуализация DOM-дерева через ASCII.

## Безопасность

- Временный файл создаётся только с правами 0644.
- Вывод на `stderr` через `js.LogWriter = os.Stderr` позволяет перехватить логи JS.
- Нет использования внешних файлов (кроме временного HTML).

## Поддерживаемые операции

- Парсинг HTML.
- Выполнение JavaScript с DOM API.
- Применение CSS.
- Вычисление layout.
- Рендеринг в ASCII.
- Вывод результата.
