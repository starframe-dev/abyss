    function makeAttrMethods() {
        return {
            setAttribute: function(name, value) {
                __go_setAttribute(this, String(name), String(value))
            },

            getAttribute: function(name) {
                return __go_getAttribute(this, String(name))
            },

            removeAttribute: function(name) {
                __go_setAttribute(this, String(name), '')
            },

            hasAttribute: function(name) {
                var val = __go_getAttribute(this, String(name))
                return val !== null && val !== undefined && val !== ''
            },

            toggleAttribute: function(name, force) {
                name = String(name)
                var present = this.hasAttribute(name)
                var shouldAdd = arguments.length > 1 ? !!force : !present
                if (shouldAdd !== present) {
                    __go_setAttribute(this, name, '')
                }
                return shouldAdd
            },

            get className() {
                return __go_getAttribute(this, 'class') || ''
            },

            set className(val) {
                __go_setAttribute(this, 'class', String(val))
            },

            get id() {
                return __go_getAttribute(this, 'id') || ''
            },

            set id(val) {
                __go_setAttribute(this, 'id', String(val))
            },

            get classList() {
                if (!this._classList) {
                    this._classList = makeClassList(this)
                }
                return this._classList
            }
        }
    }

    function makeContentMethods() {
        return {
            get outerHTML() {
                return __go_outerHTML(this)
            },

            set outerHTML(html) {
                __go_setOuterHTML(this, String(html))
            },

            get innerHTML() {
                return __go_getInnerHTML(this)
            },

            set innerHTML(html) {
                __go_setInnerHTML(this, String(html))
            },

            get textContent() {
                return __go_getTextContent(this)
            },

            set textContent(text) {
                __go_setTextContent(this, String(text))
            },

            get style() {
                if (!this._style) {
                    this._style = makeStyle(this)
                }
                return this._style
            },

            set style(val) {
                if (typeof val === 'string') {
                    __go_setAttribute(this, 'style', val)
                }
            }
        }
    }

    function makeNavigationMethods() {
        return {
            get parentNode() {
                var raw = __go_getParentNode(this)
                return raw && raw.__go_id > 0 ? makeElement(raw) : null
            },

            get parentElement() {
                var parent = this.parentNode
                return parent && parent.nodeType === 1 ? parent : null
            },

            get children() {
                var raw = __go_getChildren(this)
                var result = []
                for (var i = 0; i < raw.length; i++) {
                    if (raw[i] && raw[i].__go_id > 0) result.push(makeElement(raw[i]))
                }
                return result
            },

            get childNodes() {
                return this.children
            },

            get firstChild() {
                var kids = this.children
                return kids.length ? kids[0] : null
            },

            get lastChild() {
                var kids = this.children
                return kids.length ? kids[kids.length - 1] : null
            },

            get nextSibling() {
                var parent = this.parentNode
                if (!parent) return null
                var kids = parent.children
                for (var i = 0; i < kids.length; i++) {
                    if (kids[i].__go_id === this.__go_id) return i + 1 < kids.length ? kids[i + 1] : null
                }
                return null
            },

            get previousSibling() {
                var parent = this.parentNode
                if (!parent) return null
                var kids = parent.children
                for (var i = 0; i < kids.length; i++) {
                    if (kids[i].__go_id === this.__go_id) return i > 0 ? kids[i - 1] : null
                }
                return null
            },

            get previousElementSibling() { return __go_previousElementSibling(this) },
            get nextElementSibling() { return __go_nextElementSibling(this) },
            get firstElementChild() { return __go_firstElementChild(this) },
            get lastElementChild() { return __go_lastElementChild(this) },

            get childElementCount() {
                var count = 0
                var children = this.children
                for (var i = 0; i < children.length; i++) {
                    if (children[i].nodeType === 1) count++
                }
                return count
            }
        }
    }

    function makeQueryMethods() {
        return {
            querySelector: function(sel) {
                var raw = __go_querySelector(this, String(sel))
                return raw && raw.__go_id >= 0 ? makeElement(raw) : null
            },

            querySelectorAll: function(sel) {
                var raw = __go_querySelectorAll(this, String(sel))
                var result = []
                for (var i = 0; i < raw.length; i++) {
                    if (raw[i] && raw[i].__go_id >= 0) result.push(makeElement(raw[i]))
                }
                return result
            },

            addEventListener: function(type, handler) {
                if (typeof handler === 'function') __go_addEventListener(this, String(type), handler)
            },

            removeEventListener: function(type, handler) {
                if (typeof handler === 'function') __go_removeEventListener(this, String(type), handler)
            },

            getBoundingClientRect: function() { return __go_getBoundingClientRect(this) },
            matches: function(sel) { return __go_matches(this, String(sel)) },
            closest: function(sel) { return __go_closest(this, String(sel)) },
            focus: function() {},
            blur: function() {},
            scrollIntoView: function() {},
            click: function() {}
        }
    }

    function mergeMethods(target, source) {
        Object.keys(source).forEach(function(key) {
            Object.defineProperty(target, key, Object.getOwnPropertyDescriptor(source, key))
        })
    }

    function makeElement(raw) {
        var id = raw && raw.__go_id !== undefined ? raw.__go_id : (typeof raw === 'number' ? raw : -1)
        var nodeType = raw && raw.nodeType !== undefined ? raw.nodeType : 1
        var tagName = raw && raw.tagName !== undefined ? raw.tagName : ''
        var el = { __go_id: id, nodeType: nodeType, tagName: tagName }
        mergeMethods(el, makeChildMethods())
        mergeMethods(el, makeInsertionMethods())
        mergeMethods(el, makeAttrMethods())
        mergeMethods(el, makeContentMethods())
        mergeMethods(el, makeNavigationMethods())
        mergeMethods(el, makeQueryMethods())
        return el
    }
