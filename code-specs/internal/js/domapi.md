# Спецификация: DOM API Bridge (`internal/js/domapi.go`)

## Обзор

Файл `domapi.go` предоставляет `DOMAPIScript() string`, возвращающий JavaScript-код моста DOM API. Реализация формирования исходника находится в `domapi_embed.go`; JS-фрагменты встроены в исполняемый файл через `embed.FS` и директиву `//go:embed domapi/*.js`.

## Go API

`DOMAPIScript()` возвращает значение `domapiJSSource` без преобразований. Вызывать её следует после `RegisterBridgeFunctions()`, поскольку исполняющийся JS использует зарегистрированные функции Go `__go_*`.

`domapiJSSource` инициализируется при загрузке пакета вызовом `loadDOMAPI()`. Фрагменты в `domapiFragments` соединяются в порядке `style`, `element`, `global`, разделяются переводами строки и оборачиваются в `(function() { ... })()` с завершающим переводом строки. Файлы читаются из встроенной файловой системы по путям `domapi/<имя>.js`. При ошибке чтения функция паникует с сообщением о фрагменте и исходной ошибке. Файлы не читаются с диска во время исполнения.

## Состав JavaScript API

Содержимое API определяется фрагментами `internal/js/domapi/style.js`, `element.js` и `global.js`, встроенными в бинарный файл. Обёртка предоставляет общий scope фрагментам; глобальные объекты и функции экспортируются фрагментом `global` через `this`.

### Стили и classList

`makeStyle(el)` возвращает объект с хранилищем значений, `_buildStyle()` (собирает пары `ключ:значение;`) и `_flush()` (записывает строку в атрибут `style` через `__go_setAttribute`). Доступные свойства: `color`, `backgroundColor`, `background`, `display`, `width`, `height`, `margin`, `marginTop`, `marginRight`, `marginBottom`, `marginLeft`, `padding`, `paddingTop`, `paddingRight`, `paddingBottom`, `paddingLeft`, `border`, `borderTop`, `borderRight`, `borderBottom`, `borderLeft`, `borderColor`, `borderWidth`, `borderStyle`, `borderRadius`, `fontSize`, `fontWeight`, `fontFamily`, `textAlign`, `lineHeight`, `opacity`, `visibility`, `overflow`, `position`, `top`, `right`, `bottom`, `left`, `zIndex`, `flex`, `flexDirection`, `flexWrap`, `justifyContent`, `alignItems`, `gap`, `gridTemplateColumns`, `gridTemplateRows`, `gridGap`, `minWidth`, `minHeight`, `maxWidth`, `maxHeight`, `boxShadow`, `transform`, `transition`, `animation`, `whiteSpace`, `wordBreak`, `textOverflow`, `cursor`, `listStyle`, `textDecoration`, `fontStyle`, `verticalAlign`, `float`, `clear`, `outline`, `boxSizing`. Чтение отсутствующего/пустого значения возвращает пустую строку. Свойство `style` элемента создаётся лениво; присваивание строки `style` записывает её непосредственно в атрибут.

`classList` также создаётся лениво и предоставляет `add`, `remove`, `toggle`, `contains`, работающие с атрибутом `class`.

### Элементы

`makeElement(raw)` создаёт объект с `__go_id`, `nodeType`, `tagName` (значения берутся из raw, с указанными в коде запасными значениями) и объединяет методы атрибутов, контента, навигации, запросов, манипуляции детьми и вставки. `mergeMethods` копирует дескрипторы свойств, включая геттеры и сеттеры.

Поддерживаются методы атрибутов `setAttribute`, `getAttribute`, `removeAttribute`, `hasAttribute`, `toggleAttribute`; свойства `className`, `id`, `classList`. Контент: `outerHTML`, `innerHTML`, `textContent`, `style`. Навигация: `parentNode`, `children`, `firstChild`, `lastChild`, `nextSibling`, `previousSibling`, `previousElementSibling`, `nextElementSibling`, `firstElementChild`, `lastElementChild`, `childElementCount`. Запросы и события: `querySelector`, `querySelectorAll`, `addEventListener`, `removeEventListener`, `getBoundingClientRect`, `matches`, `closest`; `focus`, `blur`, `scrollIntoView`, `click` являются пустыми заглушками.

Манипуляции деревом: `appendChild`, `removeChild`, `insertBefore`, `replaceChild`, `cloneNode`, `insertAdjacentHTML`, `insertAdjacentText`, `insertAdjacentElement`, `remove`, `contains`. Методы `before`, `after`, `replaceWith`, `prepend`, `append` обрабатывают переданные строки как текстовые узлы посредством `document.createTextNode`; фактические операции выполняются через bridge-функции и методы элементов.

### document, window и Node

`global.js` создаёт и экспортирует через `this` объекты `document`, `window`, `console`, `Node`, а также `self`, `top`, `parent`, `frames`, `globalThis` и глобальные функции таймеров, `requestAnimationFrame`, `cancelAnimationFrame`, `matchMedia`.

`document` предоставляет `nodeType`, создание элементов/текстовых узлов/фрагментов, свойства `body` и `documentElement`, поиск по ID, тегу, классу и селектору, регистрацию событий на корне, заглушки `cookie`, `forms`, `images`, `links`, `scripts`, `styleSheets`, свойство `title`, `location`, а также `write` и `writeln`. Реальное поведение и граничные условия определяются кодом `global.js`; в частности, `location` возвращает объект с пустыми полями и `reload`, а запись cookie и установка `window.location` ничего не делают.

`window` предоставляет `document`, `location`, размеры 800x600, `navigator`, `screen`, `console`, таймеры через Go bridge, немедленный вызов callback в `requestAnimationFrame`, пустые `cancelAnimationFrame` и заглушку `matchMedia`. `console.log/warn/error/info` передают строки в `__go_log` с соответствующими префиксами для warn/error/info.

`Node` содержит числовые константы `ELEMENT_NODE` (1), `ATTRIBUTE_NODE` (2), `TEXT_NODE` (3), `CDATA_SECTION_NODE` (4), `ENTITY_REFERENCE_NODE` (5), `ENTITY_NODE` (6), `PROCESSING_INSTRUCTION_NODE` (7), `COMMENT_NODE` (8), `DOCUMENT_NODE` (9), `DOCUMENT_TYPE_NODE` (10), `DOCUMENT_FRAGMENT_NODE` (11), `NOTATION_NODE` (12).

## Взаимодействие с Go

JS-фрагменты вызывают `__go_*` bridge-функции, включая операции атрибутов и HTML/текста, навигации и запросов, событий, создания и изменения узлов, таймеров и логирования. Их регистрация и точная сигнатура реализуются вне описываемого файла; перечень фактически используемых имён определяется содержимым трёх встроенных JS-фрагментов.

## Последовательность использования

1. Зарегистрировать bridge-функции через `RegisterBridgeFunctions()`.
2. Получить строку через `DOMAPIScript()`.
3. Выполнить строку в JS-контексте, где доступны зарегистрированные функции `__go_*`.
