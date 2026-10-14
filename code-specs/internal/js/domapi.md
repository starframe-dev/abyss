# Specification: DOM API Bridge (`internal/js/domapi.go`)

## Обзор

Файл реализует мост между Go и JavaScript для предоставления API DOM в среде Abyss. Код представляет собой Go-программу, которая динамически генерирует JavaScript-код, который выполняется в браузере или TUI-окружении Abyss.

## Архитектура

### Структура пакета

```
package js
```

Пакет содержит два файла:

1. **domapi.go** — объявление публичного API
2. **domapi_embed.go** — реализация и встраивание JS-фрагментов

### Публичные функции

#### `DOMAPIScript() string`

**Описание**: Возвращает исходный код JavaScript для моста DOM API.

**Поведение**:
- Функция должна вызываться **после** вызова `RegisterBridgeFunctions()`
- Возвращает содержимое константы `domapiJSSource`
- Код представляет собой IIFE (Immediately Invoked Function Expression), который при выполнении регистрирует все функции-мосты в глобальной области видимости

**Возвращаемое значение**: строка с исходным кодом JavaScript

---

## Внутренняя реализация

### Константа `domapiJSSource`

Хранит сгенерированный JS-код. Формируется в пакете `domapi_embed.go` через чтение файлов JS-фрагментов и их соединение.

### Переменная `domapiFragments`

Массив названий JS-фрагментов в порядке загрузки:
- `style` — API стилей и CSS-свойств
- `element` — методы элементов DOM
- `global` — глобальные объекты (document, window, console)

### Функция `init()`

**Цель**: Динамическое формирование `domapiJSSource` путём чтения файлов JS-фрагментов.

**Поведение**:
1. Определяет путь к директории `domapi` относительно текущего файла
2. Пытается прочитать каждый фрагмент в порядке из `domapiFragments`
3. При ошибке чтения паникует с понятным сообщением
4. Соединяет все фрагменты в IIFE-выражение

**Путь к файлам**:
- В production: `internal/js/domapi/<fragment_name>.js`
- В тестах: `internal/js/domapi/<fragment_name>.js` (относительно рабочей директории)

---

## JS-фрагменты

### 1. Фрагмент `style` (style.js)

**Назначение**: API стилей элемента.

**Реализация**:
- Функция `makeStyle(el)` создаёт объект `CSSStyleDeclaration`-подобный для элемента
- Содержит словарь стилей `data` для хранения значений
- Метод `_buildStyle()` формирует строку CSS из всех свойств
- Метод `_flush()` применяет накопленные стили через `__go_setAttribute`

**CSS-свойства** (задаются через `Object.defineProperty`):
- Базовые: `color`, `backgroundColor`, `display`, `width`, `height`
- Отступы: `margin`, `marginTop`, `marginRight`, `marginBottom`, `marginLeft`
- Отрихты: `padding`, `paddingTop`, `paddingRight`, `paddingBottom`, `paddingLeft`
- Границы: `border`, `borderTop`, `borderRight`, `borderBottom`, `borderLeft`
- `borderColor`, `borderWidth`, `borderStyle`, `borderRadius`
- Шрифт: `fontSize`, `fontWeight`, `fontFamily`
- Текст: `textAlign`, `lineHeight`, `textDecoration`, `textOverflow`, `whiteSpace`, `wordBreak`
- Расположение: `position`, `top`, `right`, `bottom`, `left`, `zIndex`
- Flexbox/Grid: `flex`, `flexDirection`, `flexWrap`, `justifyContent`, `alignItems`, `gap`, `gridTemplateColumns`, `gridTemplateRows`, `gridGap`
- Размеры: `minWidth`, `minHeight`, `maxWidth`, `maxHeight`
- Эффекты: `boxShadow`, `transform`, `transition`, `animation`
- Прочее: `opacity`, `visibility`, `overflow`, `cursor`, `listStyle`, `verticalAlign`, `float`, `clear`, `outline`, `boxSizing`

### 2. Фрагмент `element` (element.js)

**Назначение**: Методы элементов DOM.

**Реализация**:

#### Функция `makeElement(raw)`

Создаёт JS-объект элемента из Go-структуры. Содержит:
- `__go_id` — идентификатор элемента в Go
- `nodeType` — тип узла (1 для элемента)
- `tagName` — имя тега
- Слияние всех методов через `mergeMethods()`

#### Методы атрибутов (`makeAttrMethods`)

| Метод | Описание |
|-------|----------|
| `setAttribute(name, value)` | Устанавливает атрибут |
| `getAttribute(name)` | Получает атрибут |
| `removeAttribute(name)` | Удаляет атрибут |
| `hasAttribute(name)` | Проверяет наличие атрибута |
| `toggleAttribute(name)` | Переключает атрибут |
| `className` (getter/setter) | Класс элемента |
| `id` (getter/setter) | ID элемента |
| `classList` | Объект управления классами |

#### Методы контента (`makeContentMethods`)

| Свойство | Описание |
|----------|----------|
| `outerHTML` | Получение/установка полного HTML элемента |
| `innerHTML` | Получение/установка HTML дочерних элементов |
| `textContent` | Получение/установка текстового контента |
| `style` | Стили элемента (lazy-init через `makeStyle`) |

#### Методы навигации (`makeNavigationMethods`)

| Свойство | Описание |
|----------|----------|
| `parentNode` | Родительский узел |
| `children` | Массив дочерних элементов |
| `firstChild` / `lastChild` | Первый/последний дочерний узел |
| `nextSibling` / `previousSibling` | Следующий/предыдущий узел |
| `previousElementSibling` / `nextElementSibling` | Следующий/предыдущий элемент |
| `firstElementChild` / `lastElementChild` | Первый/последний дочерний элемент |
| `childElementCount` | Количество дочерних элементов |

#### Методы запросов (`makeQueryMethods`)

| Метод | Описание |
|-------|----------|
| `querySelector(sel)` | Поиск первого элемента по селектору |
| `querySelectorAll(sel)` | Поиск всех элементов по селектору |
| `addEventListener(type, handler)` | Регистрация обработчика |
| `removeEventListener(type, handler)` | Удаление обработчика |
| `getBoundingClientRect()` | Получение bounding client rect |
| `matches(sel)` | Проверяет соответствие селектору |
| `closest(sel)` | Находит ближайшего предка по селектору |
| `focus()` / `blur()` | Заглушки (Abyss не имеет фокуса) |
| `scrollIntoView()` | Заглушка |
| `click()` | Заглушка |

#### Методы манипуляций (`makeChildMethods`)

| Метод | Описание |
|-------|----------|
| `appendChild(child)` | Добавляет дочерний элемент |
| `removeChild(child)` | Удаляет дочерний элемент |
| `insertBefore(newNode, refNode)` | Вставляет перед reference node |
| `replaceChild(newNode, oldNode)` | Заменяет узел |
| `cloneNode(deep)` | Клонирует элемент |
| `insertAdjacentHTML(position, html)` | Вставляет HTML в позицию |
| `insertAdjacentText(position, text)` | Вставляет текст |
| `insertAdjacentElement(position, el)` | Вставляет элемент |
| `remove()` | Удаляет элемент из дерева |
| `contains(child)` | Проверяет вложенность |

#### Методы вставки (`makeInsertionMethods`)

| Метод | Описание |
|-------|----------|
| `before(...nodes)` | Вставляет узлы перед элементом |
| `after(...nodes)` | Вставляет узлы после элемента |
| `replaceWith(...nodes)` | Заменяет элемент на новые узлы |
| `prepend(...nodes)` | Вставляет узлы в начало |
| `append(...nodes)` | Вставляет узлы в конец |

**Вспомогательные функции**:
- `mergeMethods(target, source)` — копирует дескрипторы свойств с сохранением привязки `this` для геттеров/сеттеров
- `makeClassList(el)` — создаёт объект управления CSS-классами с методами `add`, `remove`, `toggle`, `contains`
- `makeChildMethods()` — возвращает объект с методами манипуляции детьми
- `makeInsertionMethods()` — возвращает объект с методами вставки
- `makeAttrMethods()` — возвращает объект с методами атрибутов
- `makeContentMethods()` — возвращает объект с методами контента
- `makeNavigationMethods()` — возвращает объект с методами навигации
- `makeQueryMethods()` — возвращает объект с методами запросов

### 3. Фрагмент `global` (global.js)

**Назначение**: Глобальные объекты DOM API.

**Реализация**:

#### Объект `document`

| Свойство/Метод | Тип | Описание |
|----------------|-----|----------|
| `nodeType` | const | 9 (DOCUMENT_NODE) |
| `createElement(tagName)` | method | Создаёт элемент |
| `createTextNode(text)` | method | Создаёт текстовый узел |
| `body` | getter | Корневой элемент body |
| `documentElement` | getter | Корневой элемент документа |
| `getElementById(id)` | method | Поиск по ID |
| `getElementsByTagName(tagName)` | method | Поиск всех элементов по тегу |
| `getElementsByClassName(className)` | method | Поиск всех элементов по классу |
| `querySelector(sel)` | method | Поиск первого элемента |
| `querySelectorAll(sel)` | method | Поиск всех элементов |
| `createDocumentFragment()` | method | Создаёт фрагмент документа |
| `addEventListener(type, handler)` | method | Регистрация обработчика на root |
| `removeEventListener(type, handler)` | method | Удаление обработчика |
| `cookie` | getter/setter | Заглушка |
| `forms`, `images`, `links`, `scripts`, `styleSheets` | get | Заглушки, возвращают пустые массивы |
| `_title` / `title` | prop/setter | Заголовок документа |
| `_url` / `location` | prop/setter | Объект location |
| `write(html)` | method | Записывает HTML в body |
| `writeln(html)` | method | Записывает HTML с переносом |

#### Объект `window`

| Свойство/Метод | Тип | Описание |
|----------------|-----|----------|
| `document` | getter | Объект document |
| `location` | getter/setter | Объект location |
| `innerWidth` / `innerHeight` | getter | Размеры окна (800x600 по умолчанию) |
| `navigator` | getter | Объект navigator |
| `screen` | getter | Объект screen |
| `console` | getter | Объект console |
| `requestAnimationFrame(cb)` | method | Заглушка (вызывает cb сразу) |
| `cancelAnimationFrame(id)` | method | Заглушка |
| `setTimeout(cb, ms)` | method | Таймер |
| `setInterval(cb, ms)` | method | Интервал |
| `clearTimeout(id)` | method | Очистка таймера |
| `clearInterval(id)` | method | Очистка интервала |
| `matchMedia(query)` | method | Заглушка |
| `self`, `top`, `parent`, `frames`, `window` | getter | Ссылки на window |

#### Объект `console`

| Метод | Описание |
|-------|----------|
| `log(...args)` | Вывод в Go-логи |
| `warn(...args)` | Вывод с префиксом [WARN] |
| `error(...args)` | Вывод с префиксом [ERROR] |
| `info(...args)` | Вывод с префиксом [INFO] |

#### Объект `Node`

Константы типов узлов:
- `ELEMENT_NODE = 1`
- `ATTRIBUTE_NODE = 2`
- `TEXT_NODE = 3`
- `CDATA_SECTION_NODE = 4`
- `ENTITY_REFERENCE_NODE = 5`
- `ENTITY_NODE = 6`
- `PROCESSING_INSTRUCTION_NODE = 7`
- `COMMENT_NODE = 8`
- `DOCUMENT_NODE = 9`
- `DOCUMENT_TYPE_NODE = 10`
- `DOCUMENT_FRAGMENT_NODE = 11`
- `NOTATION_NODE = 12`

#### Экспорт в глобальную область

Все объекты экспортируются через `this`:
- `this.document`, `this.window`, `this.console`, `this.Node`
- `this.self`, `this.top`, `this.parent`, `this.frames`, `this.globalThis`
- Глобальные функции: `setTimeout`, `setInterval`, `clearTimeout`, `clearInterval`, `requestAnimationFrame`, `cancelAnimationFrame`, `matchMedia`

---

## Взаимодействие с Go

### Функции-мосты (Go side)

JS-код обращается к следующим функциям Go через префикс `__go_`:

| Функция Go | Описание |
|------------|----------|
| `__go_setAttribute(element, name, value)` | Устанавливает атрибут |
| `__go_getAttribute(element, name)` | Получает атрибут |
| `__go_outerHTML(element)` | Получает outerHTML |
| `__go_setOuterHTML(element, html)` | Устанавливает outerHTML |
| `__go_getInnerHTML(element)` | Получает innerHTML |
| `__go_setInnerHTML(element, html)` | Устанавливает innerHTML |
| `__go_getTextContent(element)` | Получает textContent |
| `__go_setTextContent(element, text)` | Устанавливает textContent |
| `__go_getParentNode(element)` | Получает родителя |
| `__go_getChildren(element)` | Получает массив детей |
| `__go_querySelector(element, selector)` | Поиск по селектору |
| `__go_querySelectorAll(element, selector)` | Поиск всех по селектору |
| `__go_addEventListener(element, type, handler)` | Добавляет обработчик |
| `__go_removeEventListener(element, type, handler)` | Удалляет обработчик |
| `__go_getBoundingClientRect(element)` | Получает bounding client rect |
| `__go_matches(element, selector)` | Проверяет соответствие |
| `__go_closest(element, selector)` | Находит ближайшего предка |
| `__go_getElementById(id)` | Получает элемент по ID |
| `__go_getElementsByTagName(tagName)` | Получает элементы по тегу |
| `__go_getElementsByClassName(className)` | Получает элементы по классу |
| `__go_createElement(tagName)` | Создаёт элемент |
| `__go_createTextNode(text)` | Создаёт текстовый узел |
| `__go_getBody()` | Получает body |
| `__go_createDocumentFragment()` | Создаёт фрагмент |
| `__go_appendChild(parent, child)` | Добавляет дочерний узел |
| `__go_removeChild(parent, child)` | Удаляет дочерний узел |
| `__go_insertBefore(parent, newNode, refNode)` | Вставляет перед refNode |
| `__go_replaceChild(parent, newNode, oldNode)` | Заменяет узел |
| `__go_cloneNode(element, deep)` | Клонирует элемент |
| `__go_insertAdjacentHTML(element, position, html)` | Вставляет HTML |
| `__go_insertAdjacentText(element, position, text)` | Вставляет текст |
| `__go_insertAdjacentElement(element, element, position)` | Вставляет элемент |
| `__go_contains(element, child)` | Проверяет вложенность |
| `__go_firstElementChild(element)` | Получает первого дочернего элемента |
| `__go_lastElementChild(element)` | Получает последнего дочернего элемента |
| `__go_previousElementSibling(element)` | Получает предыдущего элемента-сестры |
| `__go_nextElementSibling(element)` | Получает следующего элемента-сестры |
| `__go_log(message)` | Вывод в консоль |
| `__go_setTimeout(callback, ms)` | Таймер |
| `__go_clearTimeout(id)` | Очистка таймера |
| `__go_setInterval(callback, ms)` | Интервал |
| `__go_clearInterval(id)` | Очистка интервала |
| `__go_getPreviousElementSibling(element)` | Предыдущий элемент-сестра |
| `__go_getNextElementSibling(element)` | Следующий элемент-сестра |

### Формат данных элементов

Каждый JS-объект элемента содержит скрытые поля:
- `__go_id` — числовой идентификатор в Go
- `nodeType` — тип узла
- `tagName` — имя тега (верхний регистр)

---

## Особенности реализации

### 1. Lazy Initialization

- `style` свойство элемента инициализируется лениво через `makeStyle(this)`
- `classList` свойство инициализируется лениво через `makeClassList(this)`

### 2. Обработка ошибок

- В production коде паника при ошибке чтения файлов JS-фрагментов
- Нет try-catch в JS-коде (Go должен обрабатывать ошибки на своей стороне)

### 3. Заглушки (No-ops)

Методы, которые не имеют реализации в Abyss:
- `focus()`, `blur()` — управление фокусом
- `scrollIntoView()` — скролл в viewport
- `click()` — клик по элементу
- `matchMedia()` — медиа-запросы
- `window.location` — управление URL
- `console` — через `__go_log`

### 4. IIFE-обертка

Все JS-коды оборачиваются в IIFE:
```javascript
(function() {
    // код
})()
```

Это предотвращает загрязнение глобальной области видимости и позволяет использовать локальные переменные.

---

## Взаимодействие

### Последовательность вызовов

1. `RegisterBridgeFunctions()` — регистрация Go-функций-мостов
2. `DOMAPIScript()` — получение JS-кода
3. Вызов полученного JS-кода в контексте (например, `eval()` или встраивание в HTML)

### Порядок загрузки фрагментов

Фрагменты подключаются в порядке:
1. `style.js` — API стилей
2. `element.js` — API элементов
3. `global.js` — глобальные объекты

---

## Обозначения

| Обозначение | Описание |
|-------------|----------|
| `__go_id` | Внутренний ID элемента в Go |
| `__go_*` | Функции-мосты к Go |
| `_style` | Внутреннее хранилище стилей |
| `_classList` | Внутреннее хранилище классов |
| `_title` | Внутреннее хранилище заголовка |
| `_url` | Внутреннее хранилище URL |

---

## Форматы данных

### Структура элемента

```javascript
{
    __go_id: number,
    nodeType: number,
    tagName: string,
    // методы и свойства добавляются через mergeMethods
}
```

### Структура CSSStyleDeclaration

```javascript
{
    _buildStyle(): string,
    _flush(),
    // свойства через Object.defineProperty
}
```

### Структура classList

```javascript
{
    add(cls: string),
    remove(cls: string),
    toggle(cls: string),
    contains(cls: string)
}
```

### Структура Node

```javascript
{
    ELEMENT_NODE: 1,
    ATTRIBUTE_NODE: 2,
    TEXT_NODE: 3,
    // ...
}
```
