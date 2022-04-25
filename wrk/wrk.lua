local json = require "json"

function init(args)
    wrk.method = "POST"
    wrk.body = json.encode({ names = { "foo", "bar", "gopher", "spam", "goland" } })
    wrk.headers["Content-Type"] = "application/x-www-form-urlencoded"
end

-- wrk -c 4 -d 10s --latency -s wrk.lua http://127.0.0.1:8080/?sleep=100ms
