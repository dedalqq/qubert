let client = {
    accessToken: null,

    setToken: function(token) {
        client.accessToken = token
    },

    get: async function(url) {
        return client.fetch("GET", url, null)
    },

    post: async function(url, data) {
        return client.fetch("POST", url, data)
    },

    fetch: async function(method, url, data=null) {
        let headers = {
            "Content-Type": "application/json",
        }

        if (client.accessToken !== null) {
            headers["X-access-token"] = client.accessToken
        }

        let opt = {
            method: method,
            headers: headers,
        }

        if (method === "POST") {
            opt["body"] = JSON.stringify(data)
        }

        let resp = await fetch(url, opt)

        if (resp.status !== 200) {
            throw new Error("wrong code")
        }

        return await resp.json()
    }
}
