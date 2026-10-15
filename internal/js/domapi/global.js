'use strict'

const document = {
    createElement(tagName) {
        if (typeof tagName !== 'string' || tagName.trim() === '') {
            throw new TypeError('tagName must be a non-empty string')
        }

        return globalThis.document.createElement(tagName)
    },

    getElementById(id) {
        if (typeof id !== 'string') {
            throw new TypeError('id must be a string')
        }

        return globalThis.document.getElementById(id)
    },
}

const window = globalThis.window

if (typeof window === 'undefined' || typeof globalThis.document === 'undefined') {
    throw new Error('DOM API requires a browser environment')
}
