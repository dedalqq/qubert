let ui = {
    left: null,
    main: null,

    element: function(name, parent, classes, options) {
        let el = document.createElement(name);

        if (classes) {
            for (const cl of classes) {
                el.classList.add(cl)
            }
        }

        if (options) {
            if (options.id !== undefined) { el.id = options.id }
            if (options.for !== undefined) { el.setAttribute("for", options.for) }
            if (options.role !== undefined) { el.role = options.role }
            if (options.href !== undefined) { el.href = options.href }
            if (options.type !== undefined) { el.type = options.type }
            if (options.name !== undefined) { el.name = options.name }
            if (options.value !== undefined) { el.value = options.value }
            if (options.scope !== undefined) { el.scope = options.scope }
            if (options.placeholder !== undefined) { el.placeholder = options.placeholder }
            if (options.tabindex !== undefined) {
                el.setAttribute("tabindex", options.tabindex)
            }
            if (options.autocomplete !== undefined) {
                el.setAttribute("autocomplete", options.autocomplete)
            }
        }

        if (parent) {
            parent.append(el)
        }

        return el
    },

    text: function(text, parent) {
        let textNode = document.createTextNode(text);

        if (parent) {
            parent.append(textNode)
        }

        return textNode
    },

    link: function(title, handler, classes) {
        let link = ui.element("a", null, classes, {href: "#"});

        link.append(title);
        link.onclick = function() {
            handler(); return false;
        };

        return link
    },

    icon: function(icon, parent) {
        return this.element("i", parent, ["fas", `fa-${icon}`])
    },

    iconText: function(icon, text, classes) {
        let span = this.element("span", null, classes);

        this.icon(icon, span);

        if (text) {
            this.text(" ", span);
            this.text(text, span);
        }

        return span
    },

    spinner: function(text) {
        let div = this.element("div");

        this.element("div", div, ["spinner-border", "spinner-border-sm"], {role: "status"});
        this.text(" ", div);
        this.text(text, div);

        return div
    },

    pageTitle: function(text) {
        let title = this.element("h3", null, [], {});
        this.text(text, title);
        return title
    },

    appendLine: function(el, parent) {
        this.element("div", parent, ["element_line"]).append(el)
    },

    append: function(el, parent) {
        parent.append(el)
    },

    clear: function(el) {
        while (el.firstChild) {
            el.removeChild(el.lastChild);
        }
    }
};