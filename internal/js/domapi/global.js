    // document object
var document = {
        nodeType: 9,

        createElement: function(tagName) {
            var raw = __go_createElement(String(tagName))
            var el = makeElement(raw)
            if (el) el.tagName = String(tagName).toUpperCase()
            return el
        },

        createTextNode: function(text) {
            return makeElement(__go_createTextNode(String(text)))
        },

        get body() {
            var raw = __go_getBody()
            if (raw && raw.__go_id > 0) {
                var el = makeElement(raw)
                el.tagName = 'BODY'
                return el
            }
            return null
        },

        get documentElement() {
            return this.body ? this.body.parentNode : null
        },

        getElementById: function(id) {
            var raw = __go_getElementById(String(id))
            if (raw && raw.__go_id > 0) {
                return makeElement(raw)
            }
            return null
        },

        getElementsByTagName: function(tagName) {
            var raw = __go_getElementsByTagName(String(tagName))
            var result = []
            for (var i = 0; i < raw.length; i++) {
                if (raw[i] && raw[i].__go_id > 0) {
                    result.push(makeElement(raw[i]))
                }
            }
            return result
        },

        getElementsByClassName: function(className) {
            var raw = __go_getElementsByClassName(String(className))
            var result = []
            for (var i = 0; i < raw.length; i++) {
                if (raw[i] && raw[i].__go_id > 0) {
                    result.push(makeElement(raw[i]))
                }
            }
            return result
        },

        querySelector: function(sel) {
            var root = this.documentElement || this.body
            if (!root || root.__go_id < 0) return null
            var raw = __go_querySelector(root, String(sel))
            if (raw && raw.__go_id >= 0) {
                return makeElement(raw)
            }
            return null
        },

        querySelectorAll: function(sel) {
            var root = this.documentElement || this.body
            if (!root || root.__go_id < 0) return []
            var raw = __go_querySelectorAll(root, String(sel))
            var result = []
            for (var i = 0; i < raw.length; i++) {
                if (raw[i] && raw[i].__go_id >= 0) {
                    result.push(makeElement(raw[i]))
                }
            }
            return result
        },

        createDocumentFragment: function() {
            return makeElement(__go_createDocumentFragment())
        },

        addEventListener: function(type, handler) {
            if (handler && typeof handler === 'function') {
                var root = this.documentElement || this.body
                if (root) __go_addEventListener(root, String(type), handler)
            }
        },
        removeEventListener: function(type, handler) {
            if (handler && typeof handler === 'function') {
                var root = this.documentElement || this.body
                if (root) __go_removeEventListener(root, String(type), handler)
            }
        },

        get cookie() { return ''; },
        set cookie(val) {},

        get forms() { return []; },
        get images() { return []; },
        get links() { return []; },
        get scripts() { return []; },
        get styleSheets() { return []; },

        _title: '',
        get title() { return this._title || ''; },
        set title(val) { this._title = String(val); },

        _url: '',
        get location() {
            return {
                href: '',
                protocol: '',
                host: '',
                hostname: '',
                port: '',
                pathname: '',
                search: '',
                hash: '',
                reload: function() {}
            }
        },

        write: function(html) {
            var body = this.body
            if (body) {
                var current = __go_getInnerHTML(body)
                __go_setInnerHTML(body, current + String(html))
            }
        },

        writeln: function(html) {
            this.write(String(html) + '\n')
        }
    }

    // window object
    var console = {
        log: function() {
            var args = Array.prototype.slice.call(arguments)
            __go_log(args.join(' '))
        },
        warn: function() {
            var args = Array.prototype.slice.call(arguments)
            __go_log('[WARN] ' + args.join(' '))
        },
        error: function() {
            var args = Array.prototype.slice.call(arguments)
            __go_log('[ERROR] ' + args.join(' '))
        },
        info: function() {
            var args = Array.prototype.slice.call(arguments)
            __go_log('[INFO] ' + args.join(' '))
        }
    }

    var window = {
        document: document,

        get location() { return document.location; },
        set location(val) {},

        get innerWidth() { return 800; },
        get innerHeight() { return 600; },

        get navigator() {
            return {
                userAgent: 'Abyss/1.0',
                platform: 'macOS',
                language: 'en-US',
                cookieEnabled: false
            }
        },

        get screen() {
            return {
                width: 800,
                height: 600,
                availWidth: 800,
                availHeight: 600,
                colorDepth: 24
            }
        },

        get console() {
            return console
        },

        requestAnimationFrame: function(cb) {
            if (typeof cb === 'function') {
                // Call immediately — Abyss doesn't have a frame loop
                cb(0)
            }
        },

        cancelAnimationFrame: function(id) {
            // No-op
        },

        setTimeout: function(cb, ms) {
            if (typeof cb !== 'function') return 0
            return __go_setTimeout(cb, Number(ms) || 0)
        },

        setInterval: function(cb, ms) {
            if (typeof cb !== 'function') return 0
            return __go_setInterval(cb, Number(ms) || 0)
        },

        clearTimeout: function(id) {
            __go_clearTimeout(id)
        },

        clearInterval: function(id) {
            __go_clearInterval(id)
        },

        matchMedia: function(query) {
            return {
                matches: false,
                media: String(query),
                onchange: null,
                addListener: function() {},
                removeListener: function() {},
                addEventListener: function() {},
                removeEventListener: function() {},
                dispatchEvent: function() {}
            }
        },

        get self() { return window; },
        get top() { return window; },
        get parent() { return window; },
        get frames() { return window; },
        get window() { return window; }
    }

    document.nodeType = 9

    var Node = {
        ELEMENT_NODE: 1,
        ATTRIBUTE_NODE: 2,
        TEXT_NODE: 3,
        CDATA_SECTION_NODE: 4,
        ENTITY_REFERENCE_NODE: 5,
        ENTITY_NODE: 6,
        PROCESSING_INSTRUCTION_NODE: 7,
        COMMENT_NODE: 8,
        DOCUMENT_NODE: 9,
        DOCUMENT_TYPE_NODE: 10,
        DOCUMENT_FRAGMENT_NODE: 11,
        NOTATION_NODE: 12
    }

    // Expose to global scope
    this.document = document
    this.window = window
    this.console = console
    this.Node = Node
    this.self = window
    this.top = window
    this.parent = window
    this.frames = window
    this.globalThis = window

    // Expose common window functions as globals
    this.setTimeout = window.setTimeout
    this.setInterval = window.setInterval
    this.clearTimeout = window.clearTimeout
    this.clearInterval = window.clearInterval
    this.requestAnimationFrame = window.requestAnimationFrame
    this.cancelAnimationFrame = window.cancelAnimationFrame
    this.matchMedia = window.matchMedia
