# Specification: `cmd/cuetty_full_test/main.go`

## Описание

Файл представляет собой интеграционные тесты для TUI (Text User Interface) веб-браузера `cuetty`, построенного с использованием библиотеки `cue-tty`. Тесты проверяют основные сценарии работы приложения: загрузку страницы, ввод URL, прокрутку, обновление и обработку ошибок.

---

## Публичный API

### Тип `main` (программа)

**Функция `main()`**
- Точка входа в приложение
- Запускает цикл тестирования всех сценариев
- При ошибке выводит сообщение на `stderr` и завершает с кодом `1`

### Тип `cue` (библиотека)

#### Функция `cue.Launch()`

**Значение:**
```go
func Launch(
    binaryPath string,
    options ...cue.Option
) (*App, error)
```

**Описание:**
Запускает TUI приложение с указанным бинарным файлом и набором опций.

**Параметры:**
- `binaryPath` — путь к бинарному файлу приложения (например, `"./bin/abyss"`)

**Опции (type `cue.Option`):**
- `cue.WithSize(w, h int)` — задаёт размеры экрана в символах
- `cue.WithDir(path string)` — задаёт рабочую директорию
- `cue.WithEnv(key, value string)` — задаёт переменную окружения

**Возвращает:**
- `*App` — экземпляр приложения
- `error` — ошибка при запуске

**Пример:**
```go
app, err := cue.Launch("./bin/abyss",
    cue.WithSize(120, 40),
    cue.WithDir("/Users/a/Space/Projects/Starframe/abyss"),
    cue.WithEnv("TERM=xterm-256color"),
)
```

---

### Тип `App`

**Метод `Close()`**
```go
func (a *App) Close() error
```
Закрывает приложение и освобождает ресурсы.

**Метод `Page()`**
```go
func (a *App) Page() *Page
```
Возвращает объект текущей страницы.

---

### Тип `Page`

**Метод `WaitStable(duration time.Duration)`**
```go
func (p *Page) WaitStable(duration time.Duration)
```
Ожидает стабильного состояния страницы (RERender не изменяет вывод).

**Метод `Text()`**
```go
func (p *Page) Text() (string, error)
```
Возвращает текстовое представление страницы.

**Метод `Type(text string)`**
```go
func (p *Page) Type(text string)
```
Вводит текст в поле ввода (для адресной строки).

**Метод `Press(key string)`**
```go
func (p *Page) Press(key string)
```
Имитирует нажатие клавиши. Поддерживает:
- `"Enter"` — Enter
- `"Down"` — стрелка вниз (прокрутка)
- `"Escape"` — Esc (выход из адреса)
- `"q"` — Quit (завершение приложения)
- `"r"` — Refresh (обновление страницы)

**Метод `WaitFor(text string, duration time.Duration)`**
```go
func (p *Page) WaitFor(text string, duration time.Duration)
```
Ожидает появления указанного текста на странице.

---

### Тип `cue.Expect` (assertion helpers)

#### Функция `cue.Expect(page)`
```go
func Expect(page *Page) *AssertionChain
```
Возвращает цепочку утверждений для проверки состояния страницы.

#### Метод `ToContain(text string)`
```go
func (c *AssertionChain) ToContain(text string) error
```
Проверяет, содержит ли страница указанный текст.

**Возвращает:**
- `error` — nil, если текст найден; ошибка, если не найден

**Пример:**
```go
if err := cue.Expect(page).ToContain("Example Domain"); err != nil {
    // обработка ошибки
}
```

---

## Сценарии тестирования

### Test 1: Initial screen (Начальный экран)
- Запуск приложения с заданными размерами
- Ожидание стабильного состояния (2 сек)
- Проверка наличия заголовка "Abyss"
- Проверка наличия поля ввода "Enter URL"
- Вывод содержимого экрана

### Test 2: Load example.com (Загрузка example.com)
- Ввод адреса "example.com"
- Нажатие Enter
- Ожидание появления "Example Domain" (15 сек)
- Проверка отображения страницы
- Вывод содержимого

### Test 3: Scrolling (Прокрутка)
- Сохранение текущего текста экрана
- Нажатие 8 раз клавиши Down
- Ожидание стабильного состояния (300 мс)
- Проверка изменения контента
- Ожидание, что экран изменился

### Test 4: Manual refresh (r key) (Ручное обновление)
- Нажатие клавиши "r"
- Ожидание стабильного состояния (2 сек)
- Проверка сохранения контента после обновления
- Вывод содержимого

### Test 5: Error page (Страница ошибки)
- Нажатие Escape (выход из адреса)
- Ввод невалидного URL
- Ожидание появления сообщения "Error:" (10 сек)
- Проверка наличия ошибки

### Test 6: Quit (Завершение)
- Нажатие клавиши "q"
- Ожидание (500 мс)
- Вызов `app.Close()`
- Проверка отсутствия ошибок при закрытии

---

## Структура данных

### App
| Поле | Тип | Описание |
|------|-----|----------|
| (внутреннее состояние) | private | Управление запуском/завершением |

### Page
| Поле | Тип | Описание |
|------|-----|----------|
| (внутреннее состояние) | private | Буфер вывода, состояние RERender |

### AssertionChain
| Поле | Тип | Описание |
|------|-----|----------|
| (внутреннее состояние) | private | Стек условий для проверки |

### Option
| Поле | Тип | Описание |
|------|-----|----------|
| (внутреннее состояние) | private | Функция настройки |

---

## Поток выполнения

```
main()
├─ Launch(binaryPath, options...)
│  ├─ WithSize(120, 40)
│  ├─ WithDir(path)
│  ├─ WithEnv("TERM=xterm-256color")
│  └─ Returns: app, error
├─ defer app.Close()
├─ Test 1: Initial screen
│  ├─ page := app.Page()
│  ├─ page.WaitStable(2 * time.Second)
│  ├─ text, _ := page.Text()
│  ├─ cue.Expect(page).ToContain("Abyss")
│  └─ cue.Expect(page).ToContain("Enter URL")
├─ Test 2: Load example.com
│  ├─ page.Type("example.com")
│  ├─ page.WaitStable(300 * time.Millisecond)
│  ├─ page.Press("Enter")
│  ├─ page.WaitFor("Example Domain", 15*time.Second)
│  ├─ text, _ := page.Text()
│  ├─ cue.Expect(page).ToContain("Example Domain")
│  └─ ...
├─ Test 3: Scrolling
│  ├─ before, _ := page.Text()
│  ├─ 8 × page.Press("Down")
│  ├─ page.WaitStable(300 * time.Millisecond)
│  ├─ after, _ := page.Text()
│  └─ before != after
├─ Test 4: Manual refresh (r key)
│  ├─ page.Press("r")
│  ├─ page.WaitStable(2 * time.Second)
│  ├─ text, _ := page.Text()
│  ├─ cue.Expect(page).ToContain("Example Domain")
│  └─ ...
├─ Test 5: Error page
│  ├─ page.Press("Escape")
│  ├─ page.WaitStable(200 * time.Millisecond)
│  ├─ page.Type("not-a-valid-url-12345.xyz")
│  ├─ page.WaitStable(100 * time.Millisecond)
│  ├─ page.Press("Enter")
│  ├─ page.WaitFor("Error:", 10*time.Second)
│  └─ cue.Expect(page).ToContain("Error:")
├─ Test 6: Quit
│  ├─ page.Press("q")
│  ├─ page.WaitStable(500 * time.Millisecond)
│  └─ app.Close()
└─ Print success message
```

---

## Ошибки

Все ошибки обрабатываются через `os.Exit(1)` при обнаружении:
- Ошибки запуска приложения
- Ошибки ожидания текста на странице
- Ошибки проверки содержимого
- Ошибок при закрытии приложения

Ошибки выводятся на `stderr` с префиксом `FAIL:`.

---

## Внешние зависимости

- `github.com/starframe-dev/cue-tty` — основная библиотека TUI
- `github.com/charmbracelet/bubbles` — UI компоненты
- `github.com/charmbracelet/bubbletea` — TUI фреймворк
- `github.com/charmbracelet/lipgloss` — стилизация текста
- `github.com/stretchr/testify` — assertion helpers (возможно, используется внутри cue)
- `golang.org/x/net` — сетевые операции (возможно, для DNS)

---

## Условия завершения

### Успешное завершение
- ALL TESTS PASSED сообщение
- Код возврата 0 (default)

### Неудачное завершение
- Вывод FAIL сообщения
- Код возврата 1 через `os.Exit(1)`
