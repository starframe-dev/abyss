# fetcher — Спецификация пакета

## Обзор

`fetcher` загружает веб-ресурсы через HTTP GET с ограничением размера, таймаутом и поддержкой отмены контекста.

## Публичный API

### Константы и ошибки

- `ErrResponseTooLarge` — sentinel-ошибка, возвращаемая, когда размер ответа превышает лимит.
- `MaxHTMLBytes` — 8 MiB.
- `MaxImageBytes` — 16 MiB.
- `MaxCSSBytes` — 2 MiB.
- `MaxJSBytes` — 4 MiB.
- `UserAgent` — `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36`.
- `DefaultClient` — изменяемый общий `*http.Client` с таймаутом 10 секунд.

### `Fetch(url string) (string, error)`
Загружает HTML с лимитом `MaxHTMLBytes`, используя `context.Background()`.

### `FetchContext(ctx context.Context, url string) (string, error)`
Загружает HTML с тем же лимитом; запрос поддерживает отмену через контекст.

### `FetchBytes(url string) ([]byte, string, error)`
Загружает бинарный ресурс с лимитом `MaxImageBytes` и возвращает тело и значение заголовка `Content-Type`. Использует `context.Background()`.

### `FetchBytesContext(ctx context.Context, url string) ([]byte, string, error)`
То же, что `FetchBytes`, с поддержкой переданного контекста.

### `FetchCSS(ctx context.Context, url string) (string, error)`
Загружает CSS с лимитом `MaxCSSBytes`.

### `FetchJS(ctx context.Context, url string) (string, error)`
Загружает JavaScript с лимитом `MaxJSBytes`.

### `ResponseReadError`
Структура содержит `BytesRead int64` и исходную ошибку `Err error`. Оборачивает ошибку чтения тела; `Error()` возвращает текст вложенной ошибки, `Unwrap()` позволяет извлечь её.

## Поведение запросов

Для каждого запроса создаётся HTTP GET с переданным контекстом, заголовками `User-Agent` (значение `UserAgent`) и `Accept`. Значения `Accept`:

- HTML: `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`
- бинарный ресурс: `image/avif,image/webp,image/png,image/jpeg,image/*;q=0.8,*/*;q=0.5`
- CSS: `text/css,*/*;q=0.1`
- JavaScript: `application/javascript,text/javascript,*/*;q=0.1`

Используется `DefaultClient`. Принимается только статус `200 OK`. Если известный `Content-Length` больше лимита, загрузка сразу завершается ошибкой. Иначе тело читается не более чем до `limit+1` байта; превышение лимита возвращает ошибку, оборачивающую `ErrResponseTooLarge`. При успехе возвращаются данные и `Content-Type` (пустая строка, если заголовок отсутствует). Для функций, возвращающих строку, байты преобразуются напрямую в строку.

## Ошибки

Ошибки создания запроса имеют вид `fetch <url>: create request: <err>`. Ошибки выполнения запроса — `fetch <url>: <err>`. Некорректный статус — `fetch <url>: <status>`. Ошибки чтения оборачиваются в `ResponseReadError` с сообщением `fetch <url>: read body: <err>` и количеством успешно считанных байт. Ошибки превышения размера оборачивают `ErrResponseTooLarge` и указывают URL и лимит.
