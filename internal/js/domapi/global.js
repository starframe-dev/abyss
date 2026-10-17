'use strict'

globalThis.document = {
    createElement(tagName) {
        if (typeof tagName !== 'string' || tagName.trim() === '') {
            throw new TypeError('tagName must be a non-empty string')
        }

        var created = __go_createElement(tagName)
        return created && created.__go_id >= 0 ? makeElement(created) : null
    },

    getElementById(id) {
        if (typeof id !== 'string') {
            throw new TypeError('id must be a string')
        }

        var found = __go_getElementById(id)
        return found && found.__go_id >= 0 ? makeElement(found) : null
    },

    querySelector(selector) {
        var found = __go_documentQuerySelector(String(selector))
        return found && found.__go_id >= 0 ? makeElement(found) : null
    },
}


