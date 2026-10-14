# Спецификация: cmd/cuetty_test/main.go

## Обзор

Файл представляет собой тестовый клиент, использующий библиотеку `cue-tty` для взаимодействия с TUI-браузером `abyss`. Программа запускает браузер, вводит URL `example.com`, ожидает загрузку страницы, выводит строки экрана и сырой PTY-лог и завершает работу, отправив клавишу `q`.

## Публичный API

### Тип `cue.App`

Представляет запущенное TUI-приложение.

#### Конструктор

```go
func Launch(command string, opts ...Option) (*App, error)
```

Запускает команду в псевотерминале с заданными параметрами.

**Параметры:**
- `command` — путь к исполняемому файлу
- `opts` — вариативные параметры запуска через функциональные опции

**Возвращает:**
- `*App` — экземпляр приложения
- `error` — ошибка запуска

**Вариативные параметры (`Option`):**

| Функция | Описание |
|---------|----------|
| `WithSize(cols, rows int)` | Задаёт размеры терминала (по умолчанию 80x24) |
| `WithDir(dir string)` | Задаёт рабочую директорию |
| `WithEnv(env ...string)` | Добавляет переменные окружения |
| `WithReplaceEnv(env ...string)` | Полностью заменяет окружение (не наследует от родителя) |
| `WithArgs(args ...string)` | Передаёт аргументы команде |
| `WithTimeout(d time.Duration)` | Задаёт общий таймаут |

#### Методы `App`

| Метод | Описание | Возврат |
|-------|----------|---------|
| `Page() *Page` | Возвращает экземпляр страницы для взаимодействия | `*Page` |
| `Close() error` | Завершает работу приложения и освобождаёт ресурсы | `error` |
| `Wait() error` | Блокируется до выхода процесса | `error` |
| `Resize(cols, rows int) error` | Меняет размер терминала | `error` |
| `RawLog() []byte` | Возвращает сырой буфер вывода PTY (для отладки) | `[]byte` |
| `RenderText() string` | Возвращает отрендерированный текст без ANSI-кодов | `string` |

### Тип `cue.Page`

Представляет страницу браузера для взаимодействия.

#### Методы `Page`

| Метод | Описание | Возврат |
|-------|----------|---------|
| `Type(text string) error` | Вводит текст по одному символу (с задержкой 5мс между символами) | `error` |
| `Press(key string) error` | Отправляет нажатие клавиши (поддерживает "Enter", "Up", "q" и т.п.) | `error` |
| `Text() (string, error)` | Возвращает видимое содержимое терминала как plain text | `(string, error)` |
| `Lines() ([]string, error)` | Возвращает строки экрана как слайс строк | `([]string, error)` |
| `Line(n int) (string, error)` | Возвращает конкретную строку по индексу (0-based) | `(string, error)` |
| `Cursor() (row, col int, err error)` | Возвращает координаты курсора | `(row, col, error)` |
| `WaitFor(text string, timeout time.Duration) error` | Ожидает появления текста на экране | `error` |
| `WaitStable(d time.Duration)` | Задерживает выполнение для стабилизации TUI (по умолчанию 100мс) | `nil` |
| `MouseClick(x, y int) error` | Симулирует клик мыши по координатам | `error` |
| `MouseMove(x, y int) error` | Перемещает курсор мыши | `error` |
| `MouseDown(button, x, y int) error` | Нажимаем кнопку мыши (0=левая, 1=центральная, 2=правая) | `error` |
| `MouseUp(button, x, y int) error` | Отпускаем кнопку мыши | `error` |
| `MouseScrollUp(x, y int) error` | Прокрутка вверх | `error` |
| `MouseScrollDown(x, y int) error` | Прокрутка вниз | `error` |
| `EnableMouse() error` | Включает отслеживание мыши | `error` |
| `DisableMouse() error` | Отключает отслеживание мыши | `error` |
| `SaveArtifact(dir, name string) error` | Сохраняет состояние экрана в директорию | `error` |

### Тип `cue.Expectation`

Предоставляет методы для асинхронных ожиданий/аSSERTий.

#### Методы `Expectation`

| Метод | Описание | Возврат |
|-------|----------|---------|
| `Expect(p *Page) *Expectation` | Создает экземпляр ожиданий | `*Expectation` |
| `ToContain(substr string) error` | Проверяет наличие подстроки на экране | `error` |
| `ToMatch(pattern string) error` | Проверяет соответствие регулярному выражению | `error` |
| `ToHaveLine(row int, expected string) error` | Проверяет содержимое строки | `error` |
| `ToHaveCursorAt(row, col int) error` | Проверяет позицию курсора | `error` |
| `ToNotContain(substr string) error` | Проверяет отсутствие подстроки | `error` |

### Тип `cue.Event`

Записывает событие в таймлайн.

```go
type Event struct {
    Timestamp int64 `json:"timestamp"`
    Action    string `json:"action"`
    Detail    string `json:"detail"`
}
```

#### Функции

| Функция | Описание | Возврат |
|---------|----------|---------|
| `NewEvent(action, detail string) Event` | Создает событие с текущим временем | `Event` |

### Тип `cue.Artifact`

Записывает артефакт состояния.

```go
type Artifact struct {
    Name     string  `json:"name"`
    Render   string  `json:"render"`
    ANSI     string  `json:"ansi"`
    Timeline []Event `json:"timeline"`
}
```

#### Функции

| Функция | Описание | Возврат |
|---------|----------|---------|
| `WriteTimeline(dir string, events []Event) error` | Записывает markdown-таймлайн в файл | `error` |

## Внутренние реализации

### `cue.Options`

```go
type Options struct {
    Args       []string
    Env        []string
    EnvReplaced bool
    Dir        string
    Width      int
    Height     int
    Timeout    time.Duration
}
```

### `cue.Page` (внутренние поля)

```go
type Page struct {
    app     *App
    dragBtn int // текущая нажатая кнопка мыши (-1 если нет)
    dragX   int // последняя X координата перетаскивания
    dragY   int // последняя Y координата перетаскивания
}
```

### `cue.App` (внутренние поля)

```go
type App struct {
    backend    *pty.Backend
    parser     *ansi.Parser
    screen     *ansi.Screen
    opts       Options
    mu         sync.Mutex
    closeOnce  sync.Once
    done       chan struct{}
    stopReader chan struct{}
    rawLog     []byte // сырой вывод PTY
}
```

### `ansi.Parser`

Парсит сырые байты терминала, поддерживая:
- ANSI escape-последовательности
- CSI-команды (движение курсора, SGR, очистка и т.п.)
- OSC-последовательности
- UTF-8 кодировку
- Буферизацию неполных последовательностей для корректной обработки

```go
type Parser struct {
    screen       *Screen
    state        parseState
    params       []int
    intermed     []byte
    pending      []byte
    csiPriv      bool
    utfBuf       [4]byte
    utfLen       int
    utfRemaining int
    pendingCSI   []byte
}
```

### `ansi.Screen`

Виртуальный терминальный экран, поддерживающий:
- Рендеринг текста и ANSI-кодов
- Стилизацию (цвет, жирный, курсив и т.п.)
- Управление курсором
- Прокрутку

```go
type Screen struct {
    *portalis.Screen
    Style Style
}
```

### `ansi.Color`

```go
type Color struct {
    R, G, B uint8
    Is256   bool
    Index   uint8
}
```

Поддерживает 16 цветовых палитр, 256-цветную палитру и true-color (RGB 0-255).

### `ansi.Style`

```go
type Style struct {
    Fg        Color
    Bg        Color
    Bold      bool
    Dim       bool
    Italic    bool
    Underline bool
    Reverse   bool
}
```

### `pty.Backend`

Обёртка над псевотерминалом.

```go
type Backend struct {
    master *os.File
    cmd    *exec.Cmd
}
```

### `cue.KeyMap`

Карта ключей для конвертации в терминальные последовательности:

```go
var KeyMap = map[string][]byte{
    "Enter":     {0x0D},
    "Return":    {0x0D},
    "Tab":       {0x09},
    "Backspace": {0x7F},
    "Escape":    {0x1B},
    "Space":     {0x20},
    "Up":        {0x1B, 0x5B, 0x41},
    "Down":      {0x1B, 0x5B, 0x42},
    "Right":     {0x1B, 0x5B, 0x43},
    "Left":      {0x1B, 0x5B, 0x44},
    "Home":      {0x1B, 0x5B, 0x48},
    "End":       {0x1B, 0x5B, 0x46},
    "PageUp":    {0x1B, 0x5B, 0x35, 0x7E},
    "PageDown":  {0x1B, 0x5B, 0x36, 0x7E},
    "Delete":    {0x1B, 0x5B, 0x33, 0x7E},
    "Insert":    {0x1B, 0x5B, 0x32, 0x7E},
    "F1":        {0x1B, 0x4F, 0x50},
    "F2":        {0x1B, 0x4F, 0x51},
    // ... и так далее до F12
}
```

### `cue.ResolveKey(key string)`

Конвертирует имя клавиши в терминальные байты. Поддерживает:
- Прямые ключи из `KeyMap`
- `Ctrl+<характер>` (возвращает 1-й контрольный символ)
- Односимвольные клавиши
- Возвращает `nil` для неизвестных клавиш

### `cue.MouseSGR(button, x, y int, release bool)`

Кодирует событие мыши по протоколу SGR 1006:
- `button`: 0=левая, 1=центральная, 2=правая, 64=скролл вверх, 65=скролл вниз
- `x, y`: 0-based координаты ячеек
- `release=true` отправляет событие отпускания кнопки

### `cue.MouseSGRMove(x, y int)`

Кодирует событие движения мыши (без события нажатия/отпускания).

## Пример использования (из main.go)

```go
app, err := cue.Launch("./bin/abyss",
    cue.WithSize(200, 60),
    cue.WithDir("/Users/a/Space/Projects/Starframe/abyss"),
    cue.WithEnv("TERM=xterm-256color"),
)
if err != nil {
    fmt.Printf("Launch error: %v\n", err)
    return
}
defer app.Close()

page := app.Page()
time.Sleep(1 * time.Second)

// Load example.com
page.Type("example.com")
time.Sleep(200 * time.Millisecond)
page.Press("Enter")

// Wait for page to load
time.Sleep(5 * time.Second)

// Dump page lines
lines, _ := page.Lines()
fmt.Printf("=== PAGE LINES (%d lines, 200x60) ===\n", len(lines))
for i, line := range lines {
    trimmed := line
    for len(trimmed) > 0 && trimmed[len(trimmed)-1] == ' ' {
        trimmed = trimmed[:len(trimmed)-1]
    }
    if trimmed == "" {
        fmt.Printf("[%2d] (empty)\n", i)
    } else {
        fmt.Printf("[%2d] %q\n", i, trimmed)
    }
}

// Also dump raw log for ANSI content
raw := app.RawLog()
fmt.Printf("\n=== RAW LOG (last 2000 bytes) ===\n")
start := len(raw) - 2000
if start < 0 {
    start = 0
}
for i := start; i < len(raw); i++ {
    b := raw[i]
    if b == '\033' {
        fmt.Printf("\\033")
    } else if b == '\n' {
        fmt.Printf("\\n\n")
    } else if b == '\r' {
    } else if b >= 32 && b <= 126 {
        fmt.Printf("%c", b)
    } else {
        fmt.Printf("\\x%02x", b)
    }
}
fmt.Println()

page.Press("q")
time.Sleep(200 * time.Millisecond)
```

## Поток выполнения

1. `cue.Launch()` создаёт PTY-сессию и запускает процесс `./bin/abyss` (размер 200x60, рабочая директория репозитория, `TERM=xterm-256color`).
2. В фоне запускается `app.readLoop()` для чтения вывода; парсинг ANSI-кодов происходит автоматически через `ansi.Parser`; `app.write()` отправляет данные в PTY.
3. `app.Page()` возвращает `*Page`; ожидание 1 с даёт время на инициализацию.
4. `page.Type("example.com")` вводит URL, задержка 200 мс, затем `page.Press("Enter")`.
5. Ожидание 5 с до загрузки страницы.
6. `page.Lines()` выводит строки экрана с отрезанием концовых пробелов; пустые строки печатаются как `[%2d] (empty)`, остальные — `[%2d] %q`.
7. `app.RawLog()` печатает последние 2000 байт сырого PTY-вывода с экранированием (ESC → `\033`, LF → `\n`, CR пропускается, печатные → как есть, остальные → `\xNN`).
8. `page.Press("q")` и задержка 200 мс; `defer app.Close()` завершает процесс и освобождает ресурсы.

## Ошибки

- Ошибки от `Launch()`, `Page().Type()`, `Page().Press()`, `Page().MouseClick()` и других методов взаимодействия возвращаются напрямую.
- `main()` проверяет только ошибку `Launch()`; ошибки `Type()`, `Press()` и `Lines()` игнорируются (пустое значение `_`).
- `Page().Type()` игнорирует ошибки от `app.write()` внутри цикла ввода.
- `Page().Lines()` и `Page().Line()` возвращают `nil` при ошибке из `Text()` вместо паники.
- `Page().Line()` возвращает `""` при выходу за границы, а не ошибку.

## Внутренние детали

### Чтение потока

- `readLoop()` читает данные из PTY в цикле
- Буферизация в 4096 байт
- Данные передаются через канал `dataCh`
- `rawLog` накапливает все байты для отладки
- Автоматические ответы на запросы курсора (`\x1b[6n`) и статуса (`\x1b[5n`)

### Парсинг

- `Parser.Feed()` обрабатывает сырые байты
- Неполные последовательности буферизуются для корректной обработки
- Поддерживаются все ANSI-кодировки (SGR, CSI, OSC, SS2, SS3)

### Стилизация

- `ansi.Style` хранится отдельно от `Screen`
- `syncStyle()` копирует стилизацию в `portalis.Screen.Cursor`
- Поддержка 16, 256 и true-color палитр