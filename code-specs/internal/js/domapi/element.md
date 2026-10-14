# Спецификация: `internal/js/domapi/element.js`

## Описание

Фрагмент JS DOM API: методы элемента (`appendChild`, атрибуты, `querySelector`, события). Встраивается в тот же IIFE, что `style.js` и `global.js`.

## Ответственность

- прототип элемента и привязка к native-bridge;
- без собственного глобального scope — общий замыкающий IIFE.
