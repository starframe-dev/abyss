# Specification: cmd/external_test/main.go

## Обзор

`cmd/external_test/main.go` — это CLI-приложение, которое загружает веб-страницу по HTTP-адресу, выполняет JavaScript-код внутри неё с использованием встроенной реализации DOM, вычисляет CSS-стили, выполняет макетирование, рендерит результат в буфер кадров и печатает вывод в stdout в виде ASCII-графики с поддержкой цветов.

## Поведение программы

### Основной цикл

1. **Загрузка страницы**: Программа запрашивает URL (`http://localhost:9999/test_external_js.html`) через `fetcher.Fetch(url)` с использованием общего HTTP-клиента `fetcher.DefaultClient` с таймаутом 10 секунд (локальная константа `defaultClientTimeout`).

2. **Парсинг HTML**: Полученный HTML-документ парсится в DOM-дерево с помощью `dom.Parse()`.

3. **Инициализация JavaScript-контекста**: Создаётся новый контекст исполнения JavaScript через `js.NewContext()` (JavaScriptCore, macOS). Контекст закрывается через `defer jsc.Close()`.

4. **Регистрация мостов**: Реализации Go-функций регистрируются как мосты для вызова из JavaScript через `jsc.RegisterBridgeFunctions()`.

5. **Инициализация DOM**: Финальное DOM-дерево инициализируется через `js.InitDOM(doc)`.

6. **Инициализация DOM API**: Выполняется скрипт `js.DOMAPIScript()` в контексте через `jsc.Evaluate()`, который инициализирует глобальные объекты DOM API.

7. **Выполнение скриптов**: Выполняются `<script>`-элементы из загруженного документа через `js.ExecuteScripts(jsc, doc, url)`.

8. **Получение финального DOM**: Финальное DOM-дерево получается через `js.GetFinalDOM()`.

9. **Вычисление стилей**: Для всех узлов документа вычисляются итоговые CSS-стили с учётом каскада и инлайн-стилей через `css.Compute()`.

10. **Построение макета**: Строится дерево макета (`layout.Build`) с вычислением геометрии всех элементов при ширине `viewportWidth = 800`.

11. **Рендеринг**: Дерево макета рендерится в буфер кадров (`render.Render`), затем в ASCII-графику (`ascii.Render`).

12. **Вывод**: ASCII-строки печатаются построчно в stdout, после чего печатается маркер `=== END ===`.

### Обработка ошибок

- Ошибки сети: выводятся на stderr, программа завершается с кодом 1.
- Ошибки инициализации мостов: выводятся на stderr, программа завершается с кодом 1.
- Ошибки выполнения DOM API: выводятся на stderr, программа завершается с кодом 1.
- Ошибки выполнения скриптов: выводятся на stderr, программа завершается с кодом 1.

## Публичный API

### Функция `main()`

**Подпись**: `func main()`

**Описание**: Запускает основной цикл обработки страницы.

**Параметры**: none

**Возвращает**: void

**Побочные эффекты**:
- Вывод ASCII-графики и маркера `=== END ===` в stdout
- Ошибки выводятся в stderr
- Завершение программы с кодом выхода 0 (успех) или 1 (ошибка)

## Типы данных

### Внутренние типы (из зависимостей)

| Тип | Пакет | Описание |
|-----|-------|----------|
| `ascii.Cell` | `github.com/starframe/abyss/internal/ascii` | Ячейка терминала с символом и цветами |
| `dom.Node` | `github.com/starframe/abyss/internal/dom` | Узел DOM-дерева |
| `css.Color` | `github.com/starframe/abyss/internal/css` | Цвет в формате RGB |
| `css.ComputedStyle` | `github.com/starframe/abyss/internal/css` | Итоговый CSS-стиль узла |
| `layout.LayoutBox` | `github.com/starframe/abyss/internal/layout` | Узел макета с геометрией |
| `framebuffer.Framebuffer` | `github.com/starframe/abyss/internal/framebuffer` | Буфер пикселей RGB |
| `js.Context` | `github.com/starframe/abyss/internal/js` | Контекст исполнения JavaScript |

### Константы и значения

- `viewportWidth = 800` — локальная константа файла: стандартная ширина виджета (совпадает с `layout.ViewportWidth`)
- `defaultClientTimeout = 10 * time.Second` — локальная константа: стандартный таймаут HTTP-клиента
- `defaultClient = fetcher.DefaultClient` — локальная переменная: общий HTTP-клиент с таймаутом 10 секунд

## Зависимости

### Внешние

- `golang.org/x/net/html` — парсер HTML (используется в `dom`)

### Внутренние

- `github.com/starframe/abyss/internal/ascii` — ASCII-рендеринг
- `github.com/starframe/abyss/internal/css` — вычисление CSS-стилей
- `github.com/starframe/abyss/internal/dom` — DOM-дерево
- `github.com/starframe/abyss/internal/fetcher` — загрузка HTTP
- `github.com/starframe/abyss/internal/js` — JavaScript-контекст и мосты
- `github.com/starframe/abyss/internal/layout` — макетирование
- `github.com/starframe/abyss/internal/render` — рендеринг в буфер
- `github.com/starframe/abyss/internal/framebuffer` — тип буфера кадров, возвращаемого `render.Render`

## Поток данных

```
fetcher.Fetch(url) → html, err
  ↓
dom.Parse(html) → doc
  ↓
js.NewContext() → jsc (defer jsc.Close())
  ↓
jsc.RegisterBridgeFunctions()
  ↓
js.InitDOM(doc)
  ↓
jsc.Evaluate(js.DOMAPIScript())
  ↓
js.ExecuteScripts(jsc, doc, url)
  ↓
js.GetFinalDOM() → finalDOM
  ↓
css.Compute(finalDOM) → styles
  ↓
layout.Build(finalDOM, styles, 800) → box
  ↓
render.Render(box, 800) → fb
  ↓
ascii.Render(fb, box, 800) → lines
  ↓
печать lines и "=== END ===" в stdout
```

## Инварианты

1. **DOM-дерево**: После парсинга HTML документ имеет корневой узел `DocumentNode` с вложенными элементами и текстовыми узлами.

2. **JavaScript-контекст**: Контекст закрывается через `defer` после создания для корректного освобождения ресурсов.

3. **Кэш стилей**: Результаты вычисления стилей кэшируются на уровне документа. `css.InvalidateCache()` вызывается при модификации DOM через JS.

4. **Макет**: Дерево макета содержит только видимые элементы (элементы с `display: none` игнорируются).

5. **Буфер кадров**: Высота сетки ASCII-рендера вычисляется как `fb.Height / 2`, плюс один ряд при нечётной высоте.

## Границы модуля

- **Входные данные**: URL строка в формате http:// или https://
- **Выходные данные**: ASCII-строки и маркер `=== END ===` в stdout, ошибки в stderr
- **Побочные эффекты**: Вывод в stderr, завершение программы

## Важные детали реализации

### JavaScriptCore интеграция

- Используется CGO для связи с JavaScriptCore на macOS
- Функции мостов экспортируются через `goBridgeDispatch`
- Все JS-значения проходят через C-обёртки (`JSValueRef`)

### CSS-каскад

- Стили вычисляются с учётом специфики селекторов
- Инлайн-стили имеют наивысший приоритет
- Результаты кэшируются для оптимизации

### Макетирование

- Используется box model (margin, border, padding, content)
- Поддержка flexbox и grid
- Текст обёртывается по ширине контейнера

### ASCII-рендеринг

- 2 CSS-пикселя = 1 строка терминала
- Поддержка ANSI-цветов
- Текст накладывается поверх цветового фона

## Примеры вызова

### Основной вызов

```go
func main() {
    url := "http://localhost:9999/test_external_js.html"
    html, err := fetcher.Fetch(url)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fetch error: %v\n", err)
        os.Exit(1)
    }
    // ... обработка
}
```

### Запуск через `go run`

```bash
go run cmd/external_test/main.go
```

## Тесты

В самом файле тестового кода нет. Тесты локальных констант (`viewportWidth`, `defaultClientTimeout`, `defaultClient`) и конвейера обработки находятся в `main_test.go` в том же пакете.
