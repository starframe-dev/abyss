# Спецификация bridge_query.go

## Назначение
Файл пакета `js` содержит cgo-декларации JavaScriptCore и функции-мосты, преобразующие аргументы JS в вызовы DOM-запросов Go, а результаты — в JS-значения. В файле реализованы функции `getBodyJS`, `getElementByIdJS`, `querySelectorJS`, `querySelectorAllJS`, `getAttributeJS`, `getParentNodeJS`, `getChildrenJS`, `getElementsByTagNameJS`, `getElementsByClassNameJS`, `matchesJS`, `closestJS`, `previousElementSiblingJS`, `nextElementSiblingJS`, `firstElementChildJS` и `lastElementChildJS`.

## Cgo-декларации
В C-блоке подключены `<JavaScriptCore/JavaScriptCore.h>` и `<stdlib.h>`. Объявлены внешние функции `makeGoFunction`, `_getType`, `_valueToString`, `_makeArray`, `_fprintf_stderr`, `_protectValue`, `_unprotectValue`, `_callAsFunction`, `_makeString` и `_makeBoolean` с сигнатурами, указанными в этом блоке исходного файла. В данном Go-файле эти функции напрямую используются только `_makeArray`; остальные являются декларациями.

## Общее поведение
- Все мосты принимают `*Context` и `[]C.JSValueRef` и возвращают `C.JSValueRef`.
- При нехватке аргументов применяют указанное ниже значение по умолчанию.
- Ошибки `getArgString` явно проверяются только в `matchesJS` и `closestJS`. В остальных функциях ошибка игнорируется, а возвращённая строка используется.
- Узловые результаты преобразуются посредством `nodeToJS`; отсутствие узла обычно возвращает `null`.
- Результаты-коллекции преобразуются в массив через `C._makeArray`; пустая коллекция передаётся как `nil`, размер `0`. Для непустой используется указатель на первый элемент. Переменная исключения передаётся в C-функцию, но результат создания массива и исключение отдельно не проверяются.

## Функции

### `getBodyJS`
Вызывает `getBody()`. При `nil` возвращает `ctx.MakeNull()`, иначе — `nodeToJS(ctx, body)`. Аргументы не использует.

### `getElementByIdJS`
Если аргументов меньше одного — `null`. Иначе получает строку первого аргумента через `getArgString` (ошибку игнорирует), вызывает `getElementByID`; отсутствие узла даёт `null`, иначе возвращает `nodeToJS`.

### `querySelectorJS`
Требует два аргумента, иначе `null`. Первый преобразуется `jsToNodeID`; ID меньше нуля даёт `null`. Второй преобразуется в строку через `getArgString` (ошибка игнорируется), затем вызывается `querySelector`. Отсутствующий узел даёт `null`, иначе возвращается `nodeToJS`.

### `querySelectorAllJS`
Требует два аргумента, иначе возвращает `undefined`. Отрицательный NodeID также даёт `undefined`. Строка селектора получается через `getArgString` без проверки ошибки. Вызывает `querySelectorAll`, преобразует узлы посредством `nodeToJS` и возвращает массив через `_makeArray`, включая пустой массив.

### `getAttributeJS`
Требует два аргумента, иначе `null`. Отрицательный NodeID даёт `null`. Имя атрибута берётся через `getArgString` без проверки ошибки; результат `getAttribute` всегда оборачивается в `ctx.MakeString`, в том числе если строка пуста.

### `getParentNodeJS`
Требует один аргумент, иначе `null`. Отрицательный NodeID или отсутствие родителя дают `null`; иначе возвращается `nodeToJS` родителя.

### `getChildrenJS`
Требует один аргумент, иначе `undefined`. Отрицательный NodeID также даёт `undefined`. Преобразует результат `getChildren` через `nodeToJS` и возвращает массив через `_makeArray`, включая пустой массив.

### `getElementsByTagNameJS` и `getElementsByClassNameJS`
Каждая функция требует один аргумент, иначе возвращает `undefined`. Первый аргумент преобразуется через `getArgString` без проверки ошибки. Вызывает соответствующую функцию поиска, преобразует узлы и возвращает массив через `_makeArray`, включая пустой массив.

### `matchesJS`
Требует два аргумента; недостаток аргументов, отрицательный NodeID или ошибка преобразования селектора возвращают `false`. Иначе вызывает `matches(id, selector)` и возвращает результат как JS boolean.

### `closestJS`
Требует два аргумента; недостаток аргументов, отрицательный NodeID или ошибка преобразования селектора возвращают `null`. Иначе вызывает `closest`; отсутствие узла возвращает `null`, найденный узел преобразуется `nodeToJS`.

### `previousElementSiblingJS`, `nextElementSiblingJS`, `firstElementChildJS`, `lastElementChildJS`
Каждая функция требует один аргумент; при нехватке аргументов, отрицательном NodeID или отсутствии соответствующего узла возвращает `null`. Иначе вызывает одноимённую функцию поиска и преобразует узел через `nodeToJS`.

## Зависимости от других частей пакета
Файл использует `Context` (`MakeNull`, `MakeUndefined`, `MakeString`, `MakeBoolean`, поле `ctx`), а также `getArgString`, `jsToNodeID`, `nodeToJS` и перечисленные выше функции DOM-запросов. Их реализация в этом файле не содержится.
