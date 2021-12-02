let ws = {
    socket: null,

    connect: function(token) {
        let s = new WebSocket(`ws://${window.location.host}/ws`, ["a2", token]);

        s.onopen = ws.onopen;
        s.onclose = ws.onclose;
        s.onmessage = ws.onmessage;
        s.onerror = function(e) {
            console.log(e)
        }

        ws.socket = s
    },

    onmessage: function(message) {
        let data = JSON.parse(message.data);

        console.log(data)

        switch (data.type) {
            case "reload":
                core.renderModule(false).then()
                break;
            case "update":
                core.update(data.options.id, data.options.element, data.options.data).then()
                break;
        }
    },

    onopen: function() {
        console.log("onopen")
    },

    onclose: function() {
        console.log("onclose")

        core.initLoginPage()
    },

    setLocation: function(module, args) {
        ws.socket.send(JSON.stringify({
            module: module,
            args: args,
        }))
    },
}
