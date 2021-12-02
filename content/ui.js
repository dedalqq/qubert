let ui = {
    build: function (obj) {
        if (obj.tag === undefined) {
            if (obj.text !== undefined) {
                return document.createTextNode(obj.text)
            }

            console.warn("incorrect object")
        }

        let el

        if (obj.tag === "svg") {
            el = ui.svgElement("svg");

            if (obj.use) {
                let use = ui.svgElement("use");
                use.setAttributeNS(null, "href", obj.use)
                el.appendChild(use)
            }

            for (let opt of ["width", "height"]) {
                if (obj[opt] !== undefined) {
                    el.setAttributeNS(null, opt, obj[opt])
                }
            }
        } else {
            el = document.createElement(obj.tag);

            if (obj.tag === "a" && obj.href === undefined) {
                el.href = "#"
            }

            for (let opt of ["id", "for", "role", "href", "type", "name", "value", "scope", "placeholder", "tabindex", "autocomplete", "width", "height", "list", "src"]) {
                if (obj[opt] !== undefined) {
                    el.setAttribute(opt, obj[opt])
                }
            }

            if (obj.text !== undefined) {
                if (obj.tag === "textarea" || obj.tag === "pre") {
                    el.append(document.createTextNode(obj.text))
                } else {
                    let text = obj.text.split("\n")
                    for (let i=0; i<text.length; i++) {
                        if (i !== 0) {
                            el.append(document.createElement("br"))
                        }
                        el.append(document.createTextNode(text[i]))
                    }
                }
            }

            if (obj.tag === "form") {
                el.onsubmit = function(e) { return false }
            }
        }

        if (obj.classes) {
            for (const cl of obj.classes) {
                el.classList.add(cl)
            }
        }

        for (const f of ["onmousedown", "onblur", "onclick", "onkeypress", "oninput"]) {
            if (obj[f] !== undefined) {
                el[f] = function(e) {
                    return obj[f](e, el)
                }
            }
        }

        if (obj.el !== undefined) {
            for (let e of obj.el) {
                el.append(ui.build(e))
            }
        }

        if (obj.parent !== undefined) {
            obj.parent.append(el)
        }

        if (obj.cb !== undefined) {
            obj.cb(el)
        }

        return el
    },

    text: function(text) {
        return document.createTextNode(text);
    },

    clear: function(el) {
        while (el.firstChild) {
            el.removeChild(el.lastChild);
        }
    },

    svgElement: function(name) {
        return document.createElementNS("http://www.w3.org/2000/svg", name);
    }
}
