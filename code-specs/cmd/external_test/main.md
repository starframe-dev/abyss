# Спецификация: cmd/external_test/main.go

## Обзор

`cmd/external_test/main.go` — CLI-программа, загружающая фиксированную страницу по HTTP, выполняющая связанные с ней JavaScript-скрипты, затем вычисляющая стили и рендерящая результат в ASCII-графику.

## Поведение

1. Вызывает `fetcher.Fetch("http://localhost:9999/test_external_js.html")`. При ошибке печатает `Fetch error: ...` в stderr и завершает процесс с кодом 1.
2. Парсит полученный HTML вызовом `dom.Parse(string(html))`.
3. Создаёт JavaScript-контекст `js.NewContext()` и регистрирует `defer jsc.Close()`.
4. Регистрирует bridge-функции. При ошибке печатает `Bridge init error: ...` в stderr и завершает процесс с кодом 1.
5. Инициализирует DOM вызовом `js.InitDOM(doc)`.
6. Выполняет `js.DOMAPIScript()` через `jsc.Evaluate`. При ошибке печатает `DOM API error: ...` в stderr и завершает процесс с кодом 1.
7. Выполняет скрипты документа через `js.ExecuteScripts(jsc, doc, url)`. При ошибке печатает `Script error: ...` в stderr и завершает процесс с кодом 1.
8. Получает итоговый DOM через `js.GetFinalDOM()`.
9. Вычисляет стили вызовом `css.Compute(finalDOM)`, строит макет `layout.Build(finalDOM, styles, viewportWidth)` и рендерит его `render.Render(box, viewportWidth)`.
10. Преобразует результат вызовом `ascii.Render(fb, box, viewportWidth)`, печатает каждую полученную строку через `fmt.Println`, а затем печатает `=== END ===`.

Ширина viewport задана локальной константой `viewportWidth = 800`.

## Ошибки и завершение

Описанные четыре ошибки (загрузка, регистрация bridge-функций, выполнение DOM API и скриптов) выводятся в stderr и приводят к `os.Exit(1)`. Успешный путь печатает ASCII-вывод и завершающий маркер в stdout. Другой обработки ошибок в этом файле нет.

## Объявления файла

Файл объявляет `defaultClientTimeout = 10 * time.Second` и `defaultClient = fetcher.DefaultClient`, однако ни одна из этих переменных не используется в `main`; фактическая загрузка делегирована `fetcher.Fetch` без передачи клиента или таймаута из этого файла.
