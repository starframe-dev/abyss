    // Style API: element.style backed by the Go style bridge.
    function makeStyle(el) {
        var data = {}
        var styleObj = {
            _buildStyle: function() {
                var parts = []
                for (var key in data) {
                    if (Object.prototype.hasOwnProperty.call(data, key)) {
                        parts.push(key + ':' + data[key] + ';')
                    }
                }
                return parts.join('')
            },
            _flush: function() {
                __go_setAttribute(el, 'style', this._buildStyle())
            }
        }

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
                get: function() { return data[prop] || '' },
                set: function(value) {
                    data[prop] = String(value)
                    styleObj._flush()
                },
                enumerable: true,
                configurable: true
            })
        })

        return styleObj
    }
