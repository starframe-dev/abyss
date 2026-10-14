# Specification: cmd/debug_render/main.go

## Обзор

Этот модуль представляет собой утилиту для отладочного рендеринга веб-страниц в формате ASCII-арт. Программа загружает HTML-документ, выполняет JavaScript-скрипты, вычисляет CSS-стили и преобразует результат в текстовый рендеринг.

## Входящие зависимости

- `github.com/starframe-dev/abyss/internal/ascii` — модуль рендеринга ASCII-графики
- `github.com/starframe-dev/abyss/internal/css` — модуль вычисления CSS-стилей
- `github.com/starframe-dev/abyss/internal/dom` — модуль парсинга DOM
- `github.com/starframe-dev/abyss/internal/fetcher` — модуль загрузки HTML-контента
- `github.com/starframe-dev/abyss/internal/js` — модуль выполнения JavaScript
- `github.com/starframe-dev/abyss/internal/layout` — модуль построения макета
- `github.com/starframe-dev/abyss/internal/render` — модуль текстового рендеринга

## Публичный API

### `main()`

**Описание:** Запускает основной цикл обработки веб-страницы и сохраняет результат в файлы.

**Поведение:**
1. Загружает HTML-документ из `https://example.com`
2. Парсит HTML в DOM-дерево
3. Создаёт контекст JavaScript
4. Инициализирует DOM
5. Регистрирует функции моста между Go и JavaScript
6. Выполняет скрипты DOM API
7. Выполняет пользовательские скрипты страницы
8. Вычисляет стили на основе финального DOM
9. Строит макет страницы
10. Рендерит макет в текстовую форму
11. Преобразует ASCII-арт в строки
12. Сохраняет полный результат в `/tmp/abyss_full_ansi.txt`
13. Сохраняет обрезанную версию (до 120 символов) в `/tmp/abyss_short_ansi.txt`

**Побочные эффекты:**
- Вывод ошибок на stderr через `fmt.Fprintf`
- Вывод сообщений о завершении на stdout
- Запись файлов в `/tmp/`

**Обработка ошибок:**
- Ошибки загрузки: вывод на stderr и завершение с кодом 1
- Ошибки регистрации функций моста: вывод на stderr и завершение с кодом 1
- Ошибки выполнения DOM API: вывод на stderr и завершение с кодом 1
- Ошибки выполнения скриптов: вывод на stdout (программа продолжает работу)

### `stripANSI(s string) string`

**Описание:** Удаляет коды ANSI-escape из строки.

**Поведение:**
- Итерируется по символу строки
- При обнаружении `\033` (начало ANSI-кода) пропускает последовательность до `m`
- Возвращает очищенную строку

**Параметры:**
- `s string` — входная строка с возможным ANSI-кодом

**Возвращает:**
- `string` — строка без ANSI-кодов

## Внутренние типы

### Контекст JavaScript (через `internal/js`)

Модуль предоставляет:
- `NewContext() Context` — создание нового контекста JS
- `Close() error` — закрытие контекста
- `RegisterBridgeFunctions() error` — регистрация функций моста
- `Evaluate(script string) error` — выполнение скрипта
- `ExecuteScripts(context, doc, url) error` — выполнение скриптов страницы
- `GetFinalDOM() DOM` — получение финального DOM-дерева

### DOM API

Модуль предоставляет:
- `Parse(html string) DOM` — парсинг HTML в DOM
- `InitDOM(doc) error` — инициализация DOM

### CSS

Модуль предоставляет:
- `Compute(dom) Styles` — вычисление стилей для DOM

### Layout

Модуль предоставляет:
- `Build(dom, styles, width int) Box` — построение макета

### Render

Модуль предоставляет:
- `Render(box, width int) Framebuffer` — рендеринг макета

### ASCII

Модуль предоставляет:
- `Render(framebuffer, box, width int) []string` — рендеринг в ASCII-строки

## Поток выполнения

```
main()
├── fetcher.Fetch(url) → html
├── dom.Parse(html) → doc
├── js.NewContext() → jsc
│   └── js.InitDOM(doc)
│   └── jsc.RegisterBridgeFunctions()
│   └── jsc.Evaluate(js.DOMAPIScript())
│   └── js.ExecuteScripts(jsc, doc, url)
├── js.GetFinalDOM() → finalDOM
├── css.Compute(finalDOM) → styles
├── layout.Build(finalDOM, styles, 800) → box
├── render.Render(box, 800) → fb
└── ascii.Render(fb, box, 800) → lines
    ├── stripANSI(line) → clean line
    └── os.WriteFile("/tmp/abyss_short_ansi.txt", ...)
```

## Утилитарные функции

### `stripANSI()`

Удаляет последовательности ANSI-escape кодов из текста. ANSI-коды имеют формат `\033[m` или `\033[<код>m`. Функция пропускает всё от `\033` до `m`, оставляя только видимые символы.

## Файлы вывода

1. `/tmp/abyss_full_ansi.txt` — полный ASCII-арт с 800 символами на строку
2. `/tmp/abyss_short_ansi.txt` — обрезанная версия с максимумом 120 символов на строку

## Примечания

- Программа использует фиксированную ширину 800 символов для рендеринга
- Полная версия сохраняет все ANSI-коды, короткая — только очищенные
- Короткая версия обрезает строки, превышающие 120 символов
- Линии с пробельным контентом выводятся с номером в формате `[  #] текст`
