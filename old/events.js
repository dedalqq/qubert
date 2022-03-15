let events = {
    events: null,

    showEvent: function(title, text) {
        let toast = ui.element("div", null, ["toast"]);

        let header = ui.element("div", toast, ["toast-header"]);
        header.append(ui.iconText("info-circle", title));

        let button = ui.element("button", header, ["ml-2", "mb-1", "close"], {type: "button"});
        button.onclick = function() {
            toast.remove()
        };

        ui.element("span", button, ["toast-header"]).innerHTML = "&times;";

        ui.element("div", toast, ["toast-body"]).append(ui.text(text));

        this.events.append(toast);

        toast.style.opacity = "1";

        setTimeout(function() {
            toast.remove()
        }, 3000)
    }
};