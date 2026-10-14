# Спецификация: `browser_update.go`

## Обзор

Файл реализует функцию `Update`, которая обрабатывает события в модели Bubble Tea для TUI-браузера `Abyss`. Функция обрабатывает сообщения о размере окна, нажатия клавиш, таймеры и загрузку страниц, возвращая обновлённую модель и команду.

## Общественный API

### Функция `Update`

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
```

**Описание:**  
Основной обработчик событий модели Bubble Tea. Реализует паттерн `Update` из библиотеки `bubbletea`.

**Параметры:**
- `m Model` — текущее состояние модели браузера
- `msg tea.Msg` — входящее сообщение

**Возвращаемое значение:**
- `tea.Model` — обновлённая модель браузера
- `tea.Cmd` — команда для выполнения (может быть nil)

---

## Обработка сообщений

### 1. `tea.WindowSizeMsg`

**Обработка:**
- Обновляет размеры модели и виджета адреса
- Корректирует размер viewport (высота минус хедер и табы)
- При режиме `ModeView` обновляет контент viewport содержимым ASCII-строк

**Реализация:**
```go
case tea.WindowSizeMsg:
    m.Width = msg.Width
    m.Height = msg.Height
    m.Address.Width = msg.Width - 4
    m.Viewport.Width = msg.Width
    m.Viewport.Height = msg.Height - 4
    if m.Mode == ModeView {
        m.Viewport.SetContent(strings.Join(m.ASCIILines(), "\n"))
    }
```

---

### 2. `tea.KeyMsg`

Обработка нажатий клавиш разделена по режимам (`Mode`).

#### Режим `ModeAddress` (редактор URL)

| Клавиша | Действие |
|---------|----------|
| `enter` | Переход по URL: сохраняет активную вкладку, устанавливает режим `ModeLoading`, обнуляет зум и смещение, загружает страницу |
| `q` | Сохранение сессии, закрытие JS-контекста, выход (`tea.Quit`) |

**Логика `enter`:**
- Получает URL из поля ввода
- Если URL пустой — игнорирует
- Добавляет префикс `https://`, если нет `http://` или `https://`
- Сохраняет состояние активной вкладки
- Переключает в режим `ModeLoading`
- Устанавливает зум в 1.0 и смещение в 0
- Вызывает команду `loadPage(url, m.Width)`

**Логика `q`:**
- Вызывает `saveSession(m.Tabs)`
- Вызывает `js.CloseCurrentContext()`
- Возвращает `m, tea.Quit`

#### Режим `ModeView` (просмотр страницы)

| Клавиша | Действие |
|---------|----------|
| `ctrl+l` | Сохраняет вкладку, переключает в `ModeAddress`, фокусирует поле ввода, возвращает `textinput.Blink` |
| `r` | Перерисовка страницы (команда `reRenderCmd`) |
| `+`, `=` | Увеличение зума (`ZoomFactor + 0.5`, максимум 3.0) |
| `-`, `_` | Уменьшение зума (`ZoomFactor - 0.5`, минимум 0.5) |
| `left` | Смещение влево (`OffsetX - 10`, минимум 0) |
| `right` | Смещение вправо (`OffsetX + 10`) |
| `t` | Создание новой вкладки: сохраняет текущую, добавляет новую вкладку в режиме `ModeAddress`, фокусирует ввод, очищает контент |
| `w` | Закрытие текущей вкладки (если это не последняя) |
| `tab` | Переключение на следующую вкладку (циклично) |
| `shift+tab` | Переключение на предыдущую вкладку (циклично) |
| `q` | Сохранение сессии, закрытие JS-контекста, выход |
| *по умолчанию* | Передаётся в `Viewport.Update()` |

**Детали действий:**

- **Управление зумом:**
  ```go
  case "+", "=":
      m.active().ZoomFactor = math.Min(3.0, m.ZoomFactor()+0.5)
      return m, reRenderCmd(m.Width, m.ZoomFactor())
  ```

- **Управление горизонтальной прокруткой:**
  ```go
  case "left":
      m.active().OffsetX = max(0, m.OffsetX()-10)
  case "right":
      m.active().OffsetX = m.OffsetX() + 10
  ```

- **Управление вкладками:**
  ```go
  case "t":
      m.saveActiveTab()
      m.Tabs = append(m.Tabs, Tab{Mode: ModeAddress})
      m.ActiveTab = len(m.Tabs) - 1
      m.Mode = ModeAddress
      m.Address.SetValue("")
      m.Address.Focus()
      m.Viewport.SetContent("")
      return m, textinput.Blink
  ```

- **Закрытие вкладки:**
  ```go
  case "w":
      if len(m.Tabs) <= 1 {
          break // Не закрываем последнюю вкладку
      }
      m.Tabs = append(m.Tabs[:m.ActiveTab], m.Tabs[m.ActiveTab+1:]...)
      if m.ActiveTab >= len(m.Tabs) {
          m.ActiveTab = len(m.Tabs) - 1
      }
      m.restoreActiveTab()
  ```

- **Переключение вкладок:**
  ```go
  case "tab":
      m.saveActiveTab()
      m.ActiveTab = (m.ActiveTab + 1) % len(m.Tabs)
      m.restoreActiveTab()
  case "shift+tab":
      m.saveActiveTab()
      m.ActiveTab = (m.ActiveTab - 1 + len(m.Tabs)) % len(m.Tabs)
      m.restoreActiveTab()
  ```

#### Режим `ModeError` (ошибка загрузки)

| Клавиша | Действие |
|---------|----------|
| `ctrl+l` | Переключение в `ModeAddress`, фокусировка ввода |
| `q` | Сохранение сессии, закрытие JS-контекста, выход |

**Реализация:**
```go
case tea.KeyMsg:
    switch m.Mode {
    case ModeAddress:
        // ...
    case ModeView:
        // ...
    case ModeError:
        switch msg.String() {
        case "ctrl+l":
            m.saveActiveTab()
            m.Mode = ModeAddress
            m.Address.Focus()
            return m, textinput.Blink
        case "q":
            saveSession(m.Tabs)
            js.CloseCurrentContext()
            return m, tea.Quit
        }
    }
```

---

### 3. `tickMsg`

**Обработка:**
- Если режим `ModeLoading` — добавляет команду `tickCmd()` (не перерисовывает)
- Если JS-контекст активен и DOM «грязный» — перерисовка и команда `tickCmd()`

**Логика:**
```go
case tickMsg:
    if m.Mode == ModeLoading {
        cmds = append(cmds, tickCmd())
        break
    }
    if jsc, _ := js.GetCurrentContext(); jsc != nil {
        js.ProcessTimers()
        if js.IsDirty() {
            cmds = append(cmds, reRenderCmd(m.Width, m.ZoomFactor()))
        }
    }
    cmds = append(cmds, tickCmd())
```

**Назначение `tickCmd()`:**
- Тик каждые 500 мс
- Используется для проверки таймеров JS и перерисовки

---

### 4. `pageLoadedMsg`

**Обработка:**
- Переключение в `ModeView`
- Сохранение URL в активной вкладке
- Обновление ANSI- и ASCII-строк
- Установка контента viewport
- Переход наверх viewport (`GotoTop()`)
- Добавление `tickCmd()`

**Реализация:**
```go
case pageLoadedMsg:
    m.Mode = ModeView
    m.active().URL = msg.URL
    m.active().ANSILines = msg.ANSI
    m.active().ASCIILines = msg.Lines
    m.active().Mode = ModeView
    m.Viewport.SetContent(strings.Join(msg.Lines, "\n"))
    m.Viewport.GotoTop()
    cmds = append(cmds, tickCmd())
```

---

### 5. `pageErrorMsg`

**Обработка:**
- Переключение в `ModeError`
- Сохранение ошибки в активной вкладке

**Реализация:**
```go
case pageErrorMsg:
    m.Mode = ModeError
    m.active().Mode = ModeError
    m.active().Err = msg.Err
```

---

### 6. `default`

**Обработка:**
- Передаёт сообщение в `Address.Update()` (режим `ModeAddress`)
- Передаёт сообщение в `Viewport.Update()` (режим `ModeView`)

---

## Команды (tea.Cmd)

### `tea.Quit`

Завершает работу приложения. Вызовается при нажатии `ctrl+c` или `q`.

### `textinput.Blink`

Мигающая фокусировка поля ввода. Возвращается при переключении в `ModeAddress` или создании новой вкладки.

### `loadPage(url, viewportWidth)`

Команда загрузки страницы. Выполняет:
1. Запрос к URL через `fetcher.Fetch()`
2. Парсинг HTML через `dom.Parse()`
3. Инициализацию JS-контекста
4. Регистрация мостов
5. Инициализацию системы событий
6. Выполнение скриптов и API
7. Загрузку внешних CSS
8. Вычисление стилей, верстку, рендеринг
9. Возврат `pageLoadedMsg` или `pageErrorMsg`

### `reRenderCmd(viewportWidth, zoomFactor)`

Команда перерисовки страницы. Выполняет CSS/layout/render pipeline заново для текущего DOM. Учитывает зум viewport.

### `tickCmd()`

Таймер на 500 мс. Используется для проверки JS таймеров и перерисовки.

---

## Сохранение и восстановление сессии

### `saveSession(tabs []Tab)`

Сохраняет URL всех вкладок в файл `.abyss/session.json`. Используется для восстановления вкладок при запуске.

### `loadSession()`

Загружает сохранённые URL из сессии. Используется при инициализации модели.

---

## Вспомогательные методы

### `saveActiveTab()`

Копирует состояние текущего вида в активную вкладку:
- URL из поля адреса
- Режим
- ANSI- и ASCII-строки
- Зум и смещение

### `restoreActiveTab()`

Восстанавливает состояние модели из активной вкладки:
- Режим
- URL в поле адреса
- Контент viewport

### `active() *Tab`

Возвращает указатель на активную вкладку. Возвращает nil, если индекс недопустим.

---

## Модели и сообщения

### `Model`

```go
type Model struct {
    Mode       Mode
    Address    textinput.Model
    Viewport   viewport.Model
    Tabs       []Tab
    ActiveTab  int
    Width      int
    Height     int
    loadedPageContent string
}
```

### `Tab`

```go
type Tab struct {
    URL        string
    Mode       Mode
    ANSILines  []string
    ASCIILines []string
    Err        error
    ZoomFactor float64
    OffsetX    int
}
```

### `Mode`

```go
type Mode int

const (
    ModeAddress Mode = iota  // 0
    ModeLoading             // 1
    ModeView                // 2
    ModeError               // 3
)
```

### `pageLoadedMsg`

```go
type pageLoadedMsg struct {
    Lines []string
    ANSI  []string
    URL   string
}
```

### `pageErrorMsg`

```go
type pageErrorMsg struct {
    Err error
}
```

### `tickMsg`

```go
type tickMsg struct{}
```

---

## Важные детали реализации

1. **Защита от закрытия последней вкладки:**
   - В режиме `ModeView` при нажатии `w` проверяется `len(m.Tabs) <= 1`

2. **Коррекция индекса активной вкладки после удаления:**
   - Если `ActiveTab` выходит за границы после удаления, устанавливается последний индекс

3. **Циклическое переключение вкладок:**
   - Используется формула `(ActiveTab + 1) % len(m.Tabs)`

4. **Обработка загрузки:**
   - Во время загрузки (`ModeLoading`) перерисовка блокируется

5. **Управление JS-контекстом:**
   - При выходе (`tea.Quit`) всегда вызывается `js.CloseCurrentContext()`
   - При загрузке страницы создается новый контекст через `js.NewContext()`

6. **Хранение сессии:**
   - Файл: `~/.abyss/session.json`
   - Структура: `{"urls": ["url1", "url2", ...]}`

7. **Обработка ошибок:**
   - Ошибки `fetch`, JS-инициализации игнорируются и возвращаются через `pageErrorMsg`

---

## Диаграмма состояний (модалитетов)

```
┌─────────────────┐
│   ModeAddress   │──enter──┐
└─────────────────┘        │
                           ▼
┌─────────────────┐      ┌─────────────────┐
│   ModeLoading   │──pageLoaded──┐        │
└─────────────────┘              │        │
        │ pageError              │        │
        ▼                        │        │
┌─────────────────┐              │        │
│   ModeError     │◄─────────────┘        │
└─────────────────┘                      │
        │                                │
        │ ctrl+l                         │
        └─────────────────────────────────┘
```

---

## Пример использования

```go
// Инициализация модели
m := browser.New()

// Запуск цикла Bubble Tea
p := tea.NewProgram(m)
p.Run()
```

## Зависимости

- `github.com/charmbracelet/bubbles/textinput` — поле ввода URL
- `github.com/charmbracelet/bubbles/viewport` — прокручиваемая область контента
- `github.com/charmbracelet/bubbletea` — модель событий
- `github.com/starframe/abyss/internal/js` — JS-движок и DOM
- `github.com/starframe/abyss/internal/dom` — парсер HTML
- `github.com/starframe/abyss/internal/ascii` — рендеринг ASCII/ANSI
- `github.com/starframe/abyss/internal/css` — вычисление стилей
- `github.com/starframe/abyss/internal/layout` — построение верстки
- `github.com/starframe/abyss/internal/render` — рендеринг
- `github.com/starframe/abyss/internal/fetcher` — HTTP-запросы
- `github.com/starframe/abyss/internal/dom` — парсинг DOM
