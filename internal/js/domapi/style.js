    // Style API: element.style backed by the CSSStyleDeclaration-like object.
function makeStyle(el) {
        var data = {}
        var styleObj = {
            _buildStyle: function() {
                var parts = []
                for (var key in data) {
                    parts.push(key + ':' + data[key] + ';')
                }
                return parts.join('')
            },
            _flush: function() {
                __go_setAttribute(el, 'style', this._buildStyle())
            }
        }

        // Define common CSS properties
        var props = ['color','backgroundColor','background','display','width','height',
            'margin','marginTop','marginRight','marginBottom','marginLeft',
            'padding','paddingTop','paddingRight','paddingBottom','paddingLeft',
            'border','borderTop','borderRight','borderBottom','borderLeft',
            'borderColor','borderWidth','borderStyle','borderRadius',
            'fontSize','fontWeight','fontFamily','textAlign','lineHeight',
            'opacity','visibility','overflow','position','top','right','bottom','left',
            'zIndex','flex','flexDirection','flexWrap','justifyContent','alignItems',
            'gap','gridTemplateColumns','gridTemplateRows','gridGap',
            'minWidth','minHeight','maxWidth','maxHeight',
            'boxShadow','transform','transition','animation',
            'whiteSpace','wordBreak','textOverflow','cursor',
            'listStyle','textDecoration','fontStyle','verticalAlign',
            'float','clear','outline','boxSizing']

        props.forEach(function(prop) {
            Object.defineProperty(styleObj, prop, {
                get: function() { return data[prop] || ''; },
                set: function(val) {
                    data[prop] = val
                    this._flush()
                },
                enumerable: true,
                configurable: true
            })
        })

        return styleObj
    }

    function makeClassList(el) {
        return {
            add: function(cls) {
                var current = __go_getAttribute(el, 'class') || ''
                var classes = current ? current.split(/\s+/) : []
                if (classes.indexOf(cls) === -1) {
                    classes.push(cls)
                    __go_setAttribute(el, 'class', classes.join(' '))
                }
            },
            remove: function(cls) {
                var current = __go_getAttribute(el, 'class') || ''
                var classes = current ? current.split(/\s+/) : []
                var idx = classes.indexOf(cls)
                if (idx !== -1) {
                    classes.splice(idx, 1)
                    __go_setAttribute(el, 'class', classes.join(' '))
                }
            },
            toggle: function(cls) {
                var current = __go_getAttribute(el, 'class') || ''
                var classes = current ? current.split(/\s+/) : []
                var idx = classes.indexOf(cls)
                if (idx !== -1) {
                    classes.splice(idx, 1)
                } else {
                    classes.push(cls)
                }
                __go_setAttribute(el, 'class', classes.join(' '))
            },
            contains: function(cls) {
                var current = __go_getAttribute(el, 'class') || ''
                var classes = current ? current.split(/\s+/) : []
                return classes.indexOf(cls) !== -1
            }
        }
    }

    function makeChildMethods() {
        return {
            appendChild: function(child) {
                if (child && child.__go_id !== undefined) {
                    __go_appendChild(this, child)
                }
                return child
            },

            removeChild: function(child) {
                if (child && child.__go_id !== undefined) {
                    __go_removeChild(this, child)
                }
                return child
            },

            insertBefore: function(newNode, refNode) {
                if (newNode && newNode.__go_id !== undefined) {
                    __go_insertBefore(this, newNode, refNode || null)
                }
                return newNode
            },

            replaceChild: function(newNode, oldNode) {
                if (newNode && newNode.__go_id !== undefined && oldNode && oldNode.__go_id !== undefined) {
                    __go_replaceChild(this, newNode, oldNode)
                }
                return oldNode
            },

            cloneNode: function(deep) {
                return __go_cloneNode(this, !!deep)
            },

            insertAdjacentHTML: function(position, html) {
                __go_insertAdjacentHTML(this, String(position), String(html))
            },

            insertAdjacentText: function(position, text) {
                __go_insertAdjacentText(this, String(position), String(text))
            },

            insertAdjacentElement: function(position, el) {
                if (el && el.__go_id !== undefined) {
                    __go_insertAdjacentElement(this, el, String(position))
                }
                return el
            },

            remove: function() {
                if (this.parentNode) {
                    this.parentNode.removeChild(this)
                }
            },

            contains: function(child) {
                if (!child || child.__go_id === undefined) return false
                return __go_contains(this, child)
            }
        }
    }

    // toTextNode coerces a string argument to a text node.
    function toTextNode(node) {
        return typeof node === 'string' ? document.createTextNode(node) : node
    }

    // insertOne inserts a node before a reference using the given parent op.
    function insertOne(parent, node, ref) {
        if (node && node.__go_id !== undefined) {
            parent.insertBefore(node, ref)
        }
    }

    function makeInsertionMethods() {
        return {
            before: function() {
                var parent = this.parentNode
                if (!parent) return
                for (var i = 0; i < arguments.length; i++) {
                    insertOne(parent, toTextNode(arguments[i]), this)
                }
            },

            after: function() {
                var parent = this.parentNode
                if (!parent) return
                var ref = this.nextSibling
                for (var i = 0; i < arguments.length; i++) {
                    insertOne(parent, toTextNode(arguments[i]), ref)
                }
            },

            replaceWith: function() {
                var parent = this.parentNode
                if (!parent) return
                var ref = this.nextSibling
                for (var i = 0; i < arguments.length; i++) {
                    insertOne(parent, toTextNode(arguments[i]), ref)
                }
                parent.removeChild(this)
            },

            prepend: function() {
                var first = this.firstChild
                for (var i = 0; i < arguments.length; i++) {
                    insertOne(this, toTextNode(arguments[i]), first)
                }
            },

            append: function() {
                for (var i = 0; i < arguments.length; i++) {
                    var node = toTextNode(arguments[i])
                    if (node && node.__go_id !== undefined) {
                        this.appendChild(node)
                    }
                }
            }
        }
    }
