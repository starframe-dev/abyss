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

            toggleAttribute: function(name) {
                var val = __go_getAttribute(this, String(name))
                if (val !== null && val !== undefined && val !== '') {
                    __go_setAttribute(this, String(name), '')
                    return false
                } else {
                    __go_setAttribute(this, String(name), '')
                    return true
                }
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
                if (raw && raw.__go_id > 0) {
                    return makeElement(raw)
                }
                return null
            },

            get children() {
                var raw = __go_getChildren(this)
                var result = []
                for (var i = 0; i < raw.length; i++) {
                    var child = raw[i]
                    if (child && child.__go_id > 0) {
                        result.push(makeElement(child))
                    }
                }
                return result
            },

            get firstChild() {
                var kids = this.children
                return kids.length > 0 ? kids[0] : null
            },

            get lastChild() {
                var kids = this.children
                return kids.length > 0 ? kids[kids.length - 1] : null
            },

            get nextSibling() {
                if (!this.parentNode) return null
                var kids = this.parentNode.children
                for (var i = 0; i < kids.length; i++) {
                    if (kids[i].__go_id === this.__go_id && i + 1 < kids.length) {
                        return kids[i + 1]
                    }
                }
                return null
            },

            get previousSibling() {
                if (!this.parentNode) return null
                var kids = this.parentNode.children
                for (var i = 0; i < kids.length; i++) {
                    if (kids[i].__go_id === this.__go_id && i > 0) {
                        return kids[i - 1]
                    }
                }
                return null
            },

            get previousElementSibling() {
                return __go_previousElementSibling(this)
            },

            get nextElementSibling() {
                return __go_nextElementSibling(this)
            },

            get firstElementChild() {
                return __go_firstElementChild(this)
            },

            get lastElementChild() {
                return __go_lastElementChild(this)
            },

            get childElementCount() {
                var count = 0
                for (var i = 0; i < this.children.length; i++) {
                    if (this.children[i].nodeType === 1) count++
                }
                return count
            }
        }
    }

    function makeQueryMethods() {
        return {
            querySelector: function(sel) {
                var raw = __go_querySelector(this, String(sel))
                if (raw && raw.__go_id >= 0) {
                    return makeElement(raw)
                }
                return null
            },

            querySelectorAll: function(sel) {
                var raw = __go_querySelectorAll(this, String(sel))
                var result = []
                for (var i = 0; i < raw.length; i++) {
                    if (raw[i] && raw[i].__go_id >= 0) {
                        result.push(makeElement(raw[i]))
                    }
                }
                return result
            },

            addEventListener: function(type, handler) {
                if (handler && typeof handler === 'function') {
                    __go_addEventListener(this, String(type), handler)
                }
            },
            removeEventListener: function(type, handler) {
                if (handler && typeof handler === 'function') {
                    __go_removeEventListener(this, String(type), handler)
                }
            },

            getBoundingClientRect: function() {
                return __go_getBoundingClientRect(this)
            },

            matches: function(sel) {
                return __go_matches(this, String(sel))
            },

            closest: function(sel) {
                return __go_closest(this, String(sel))
            },

            focus: function() {
                // No-op: Abyss doesn't have focus management
            },

            blur: function() {
                // No-op
            },

            scrollIntoView: function() {
                // No-op: Abyss viewport is always at top
            },

            click: function() {}
        }
    }

    // mergeMethods copies property descriptors (including getters/setters)
    // from source onto target, so getters keep `this` bound to target. This is
    // required because Object.assign eagerly invokes source getters.
    function mergeMethods(target, source) {
        Object.keys(source).forEach(function(key) {
            Object.defineProperty(target, key, Object.getOwnPropertyDescriptor(source, key))
        })
    }

    function makeElement(raw) {
        var id = (raw && raw.__go_id !== undefined) ? raw.__go_id : (typeof raw === 'number' ? raw : -1)
        var nodeType = (raw && raw.nodeType !== undefined) ? raw.nodeType : 1
        var tagName = (raw && raw.tagName !== undefined) ? raw.tagName : ''
        var el = {
            __go_id: id,
            nodeType: nodeType,
            tagName: tagName
        }
        mergeMethods(el, makeChildMethods())
        mergeMethods(el, makeInsertionMethods())
        mergeMethods(el, makeAttrMethods())
        mergeMethods(el, makeContentMethods())
        mergeMethods(el, makeNavigationMethods())
        mergeMethods(el, makeQueryMethods())
        return el
    }
