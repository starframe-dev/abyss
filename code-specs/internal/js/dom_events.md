# dom_events

## Обзор

Эта реализация предоставляет систему событий для JavaScriptCore (JSC) с поддержкой:

- Событий узлов DOM (addEventListener, removeEventListener, dispatchEvent)
- Таймеров (setTimeout, setInterval, clearTimer)

Реализация использует нативный C API JavaScriptCore и работает через C-границу Go.

## Публичный API

### Инициализация

```go
func InitEventSystem(ctx C.JSGlobalContextRef)
```

Инициализирует систему событий с указанным контекстом JSC.

**Параметры:**
- `ctx` — указатель на глобальный контекст JavaScriptCore

**Эффекты:**
- Устанавливает глобальный контекст в `eventCtx`
- Инициализирует map слушателей событий
- Очищает таймеры (сбрасывает `timers` и `nextTimerID`)
- Вызывает `MarkClean()` для обновления состояния сборки

---

### Слушатели событий

#### addEventListener

```go
func addEventListener(nodeID NodeID, eventType string, callback C.JSObjectRef)
```

Регистрирует обратный вызов для события на узле.

**Параметры:**
- `nodeID` — идентификатор узла
- `eventType` — тип события (строка)
- `callback` — JS объект обратного вызова

**Эффекты:**
- Создаёт map слушателей для узла, если его нет
- Добавляет callback в список слушателей для типа события
- Защищает callback от сбора мусора через `_protectValue`

#### removeEventListener

```go
func removeEventListener(nodeID NodeID, eventType string, callback C.JSObjectRef)
```

Удаляет обратный вызов для события на узле.

**Параметры:**
- `nodeID` — идентификатор узла
- `eventType` — тип события (строка)
- `callback` — JS объект обратного вызова для удаления

**Эффекты:**
- Удаляет callback из списка слушателей для типа события
- Освобождает callback через `_unprotectValue`
- Если callback не найден, ничего не происходит (без ошибки)

#### dispatchEvent

```go
func dispatchEvent(nodeID NodeID, eventType string)
```

Выпускает событие на узле, вызывая всех зарегистрированных слушателей.

**Параметры:**
- `nodeID` — идентификатор узла
- `eventType` — тип события (строка)

**Поведение:**
1. Создаёт объект события через `_makeEventObject`
2. Вызывает все слушатели с объектом события
3. Проверяет свойство `_propagationStopped` между вызовами
4. Останавливает дальнейшую обработку, если propagation остановлена

---

### Объекты событий

#### makeEventObject

```go
func makeEventObject(eventType string) C.JSValueRef
```

Создаёт JS объект события с заданным типом.

**Параметры:**
- `eventType` — тип события

**Возвращает:**
- JS объект события или null в случае ошибки

**Свойства объекта события:**
- `type` (string) — тип события
- `defaultPrevented` (boolean) — флаг предотвращения по умолчанию

**Методы объекта события:**
- `preventDefault()` — устанавливает `defaultPrevented = true`
- `stopPropagation()` — устанавливает `_propagationStopped = true`

---

### Таймеры

#### setTimeout

```go
func setTimeout(callback C.JSObjectRef, delayMs int) int
```

Регистрирует обратный вызов, который будет выполнен после указанной задержки.

**Параметры:**
- `callback` — JS объект обратного вызова
- `delayMs` — задержка в миллисекундах

**Возвращает:**
- ID таймера (int)

**Эффекты:**
- Защищает callback от сбора мусора
- Добавляет таймер в список `timers`
- Устанавливает `Repeat = false`

#### setInterval

```go
func setInterval(callback C.JSObjectRef, delayMs int) int
```

Регистрирует обратный вызов, который будет выполнен каждые указанное количество времени.

**Параметры:**
- `callback` — JS объект обратного вызова
- `delayMs` — интервал в миллисекундах

**Возвращает:**
- ID таймера (int)

**Эффекты:**
- Защищает callback от сбора мусора
- Добавляет таймер в список `timers`
- Устанавливает `Repeat = true`
- При повторении таймер перезапускается с новым дедлайном

#### clearTimer

```go
func clearTimer(id int)
```

Удаляет таймер по ID.

**Параметры:**
- `id` — ID таймера для удаления

**Эффекты:**
- Удаляет таймер из списка `timers`
- Освобождает callback через `_unprotectValue`
- Ничего не происходит, если таймер не найден

---

### Обработка таймеров

#### ProcessTimers

```go
func ProcessTimers() bool
```

Выполняет все готовые таймеры. Должен вызываться с активным контекстом.

**Возвращает:**
- `true` если был выполнен хотя бы один таймер
- `false` если готовых таймеров нет

**Поведение:**
1. Делит таймеры на готовые и неготовые
2. Выполняет готовые таймеры через `_callAsFunction`
3. Для повторяющихся таймеров:
   - Пересчитывает дедлайн
   - Добавляет обратно в список таймеров
4. Для одноразовых таймеров:
   - Освобождает callback через `_unprotectValue`

---

## Типы данных

### таймер

```go
type timer struct {
    ID       int
    Callback C.JSObjectRef
    Deadline time.Time
    Interval time.Duration
    Repeat   bool
}
```

**Свойства:**
- `ID` — уникальный идентификатор таймера
- `Callback` — JS объект обратного вызова
- `Deadline` — время, когда таймер должен сработать
- `Interval` — интервал для повторяющихся таймеров
- `Repeat` — флаг повторяющегося таймера

---

## Внутренние структуры

### eventListeners

```go
eventListeners = map[NodeID]map[string][]C.JSObjectRef{}
```

Map слушателей событий:
- Ключ первого уровня: `NodeID`
- Ключ второго уровня: тип события (строка)
- Значение: массив JS объектов обратных вызовов

### timers

```go
timers = []timer{}
```

Список активных таймеров.

### eventCtx

Глобальный указатель на контекст JavaScriptCore.

### eventMu

Мьютекс для защиты `eventListeners`.

### timerMu

Мьютекс для защиты `timers` и `nextTimerID`.

### nextTimerID

Счётчик для генерации уникальных ID таймеров.

### MarkClean()

Функция для обновления состояния сборки (вероятно, атомарный обмен флагом сборки).

---

## Взаимодействие с C

### C._protectValue

```c
extern void _protectValue(JSContextRef ctx, JSValueRef value);
```

Защищает JS объект от сбора мусора. Должен вызываться при добавлении callback в систему.

### C._unprotectValue

```c
extern void _unprotectValue(JSContextRef ctx, JSValueRef value);
```

Освобождает JS объект для сбора мусора. Должен вызываться при удалении callback.

### C._callAsFunction

```c
extern JSValueRef _callAsFunction(JSContextRef ctx, JSObjectRef func, JSObjectRef thisObj, size_t argCount, const JSValueRef args[], JSValueRef* exception);
```

Выполняет JS функцию с указанными аргументами.

### C._makeString

```c
extern JSValueRef _makeString(JSContextRef ctx, const char* value);
```

Создаёт JS строку из C строки.

### C._evaluate

```c
extern JSValueRef _evaluate(JSContextRef ctx, const char* script, const char* sourceURL, JSValueRef* exception);
```

Оценка JS скрипта в контексте.

---

## Важные детали реализации

### Управление памятью

- Все callback защищаются через `_protectValue` при добавлении
- Callback освобождается через `_unprotectValue` при удалении
- При выполнении одноразовых таймеров callback освобождается после вызова

### Защита от гонки данных

- `eventMu` защищает `eventListeners` и `eventCtx`
- `timerMu` защищает `timers` и `nextTimerID`
- При `dispatchEvent` копируются callback перед освобождением мьютекса для вызова JS кода

### Обработка ошибок

- Ошибки JS возвращаются через `exc` параметр в C функциях
- При создании объекта события возвращается `nil` в случае ошибки оценки
- Нет активного управления ошибками — предполагается, что JSCore обработает ошибки

### Поведение при отсутствии слушателей

- Если `eventListeners[nodeID]` не существует — ничего не происходит
- Если `eventListeners[nodeID][eventType]` не существует — ничего не происходит
- Если список callback пуст — ничего не происходит

### Повторяющиеся таймеры

- При `Repeat = true` таймер добавляется обратно в список после выполнения
- Новый `Deadline` вычисляется как `now + Interval`
- Callback остаётся защищённым

### Массив таймеров

- Используется `append(timers, t)` вместо `push_back` для добавления
- При фильтрации используются срезы: `timers[:i]` и `timers[i+1:]`

---

## Ограничения

1. **Нет очистки старых таймеров** — старые выполненные таймеры не удаляются из списка
2. **Нет обработки ошибок JS** — предполагается, что JSCore обработает ошибки
3. **Нет проверки типов событий** — события проверяются только по строке
4. **Нет поддержки аргументов для слушателей** — слушатели вызываются только с объектом события
5. **Нет поддержки `this` для слушателей** — `thisObj` всегда `nil`
6. **ID таймеров не сбрасываются** — `nextTimerID` только увеличивается

---

## Примеры использования

### Регистрация слушателя

```go
addEventListener(nodeID, "click", myCallback)
dispatchEvent(nodeID, "click")
removeEventListener(nodeID, "click", myCallback)
```

### Таймеры

```go
id := setTimeout(callback, 1000)
ProcessTimers() // выполнить таймеры
clearTimer(id) // очистить таймер
```

### Событие с предотвращением

```go
addEventListener(nodeID, "submit", handle)
handle(event) {
    event.preventDefault() // остановить по умолчанию
    event.stopPropagation() // остановить распространение
}
```

---

## Связанные файлы

- `internal/js/*.go` — основной код JSCore bridge
- `code-specs/internal/js/dom_events.md` — эта спецификация
