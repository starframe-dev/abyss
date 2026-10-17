'use strict'

globalThis.document = {
    createElement(tagName) {
        if (typeof tagName !== 'string' || tagName.trim() === '') {
            throw new TypeError('tagName must be a non-empty string')
        }

        return __go_createElement(tagName)
    },

    getElementById(id) {
        if (typeof id !== 'string') {
            throw new TypeError('id must be a string')
        }

        return __go_getElementById(id)
    },
}


