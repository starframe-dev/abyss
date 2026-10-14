# Спецификация: проверка ASCII-картинок через cue-tty

## Цель

Доказать, что `bin/abyss` (production-бинарь) рендерит изображения как ASCII art с truecolor,
и зафиксировать дефекты, которые были видны **только** при честной проверке через cue-tty.

Предшествующая проверка через `cmd/render_test` и `--headless-dump` эти дефекты не ловила:
они обходят слой `View()`, где и происходило повреждение вывода.

## Найденные дефекты

`[2026-08-31] Проблема: data:image PNG не рендерились — показывался плейсхолдер `[image] alt` →
Решение: `fetchImages` намеренно пропускает `data:` URL, поэтому кэш оставался пустым и
`cachedImage` возвращал `nil`. `cachedImage` теперь сам декодирует `data:image/*;base64`
(`encoding/base64` + `asciirender.Decode`) и кладёт результат в кэш.
Тесты: `TestCachedImage_DataURLReturnsDecodedImage`, `TestCachedImage_InvalidDataURLReturnsNil`.

`[2026-08-31] Проблема: плейсхолдер `[image] alt` рисовался ПОВЕРХ картинки и стирал её левые
колонки (в дампе было `[image] R##f###f###...`) →
Решение: `quantizeImageBox` после успешной квантизации обнуляет `box.TextLines`, которые
`layoutImage` засеял плейсхолдером. Порядок оверлеев в `ascii.go`: сначала `overlayAsciiCells`,
затем текст — поэтому текст обязан исчезнуть, когда есть реальные ячейки.
Тесты: `TestRender_ImgPopulatesAsciiCellsAndFramebuffer`,
`TestRenderANSI_ImagePlaceholderDoesNotErasePicture` (+ `...PlaceholderRemainsWhenImageMissing`
как страховка, что фолбэк не удалён).

`[2026-08-31] Проблема: картинка в cue-tty обрывалась на 7 колонках (`###L###[K`) →
Решение: `View()` резал ANSI-строку как обычный текст — `[]rune(line)[start:end]`.
Реальная строка картинки стоит ~1300 рун на 64 видимые колонки, поэтому окно в `m.Width` рун
показывало лишь ~6 колонок, а обрезанный посередине CSI оставлял висячий `ESC`, который
терминал трактовал как erase-to-end-of-line. Добавлен ANSI-осознанный клиппинг
(`internal/browser/ansi_clip.go`: `splitANSI`, `visibleWidth`, `clipANSI`).
Подтверждение из дампа: `ESC[K count: 0` — то есть `[K` генерировал именно наш обрыв,
а не парсер cue-tty.
Тесты: `TestClipANSI_KeepsEscapeSequencesIntact`, `TestClipANSI_NoDanglingEscape`,
`TestClipANSI_RealAbyssImageLine`.

`[2026-08-31] Проблема: после одного нажатия Right и возврата назад страница становилась
полностью пустой →
Решение: `visible := lines[yoff:end]` делит backing array с `tab.ANSILines`, и `visible[i] = ...`
писал в них на месте, безвозвратно портя исходные строки. Теперь обрезка собирается в новый
срез. Регрессионный тест `TestView_HorizontalScrollKeepsImageColumnsRoundTrip` на сломанной
версии даёт `64 → 54 → 54`, на исправленной — `64 → 54 → 64`.

## Критерии успеха

Загрузка `http://127.0.0.1:18768/compact.html` (три `data:image/png` с градиентом) в свежий
`bin/abyss` под cue-tty, терминал 120x40:

```
ASCII IMAGE CHECK
###L###f###f###f###f###f###f###t###t###t###t###t###t###1###1###1     ← красная
###t###t###t###t###t###1###1###1###1###1###1###i###i###i###i###i     ← зелёная
###L###f###f###f###f###f###f###t###t###t###t###t###1###1###1###1     ← синяя
```

- `[image]` отсутствует, `[K` отсутствует
- плотностная палитра `L → f → t → 1 → i` отражает градиент яркости
- прокрутка вправо режет по колонкам, откат влево восстанавливает вид побайтово

## Проверки

```
go build -o bin/abyss .
go vet ./...                                   # чисто
go test ./internal/browser/ -count=1           # ok
go test ./... -count=1 -p 1                    # все пакеты ok
```

Полный прогон `go test ./... -count=1` без `-p 1` нестабилен: `TestAbyss_E2E_LoadPage`
таймаутится на `example.com` из-за конкуренции нескольких пакетов за PTY и сеть
(изолированно и при `-p 1` проходит). К изменениям отношения не имеет.

## Артефакты

- `/tmp/abyss-img-fixture/compact.html` — компактный фикстур (картинки попадают в первый экран)
- `/tmp/abyss-compact.ansi` — ANSI-дамп текущего бинаря через `--headless-dump`
