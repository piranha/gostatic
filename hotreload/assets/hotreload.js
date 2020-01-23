(function() {
    if (!('WebSocket' in window)) {
        return;
    }

    var abort;

    var ws = new WebSocket(((window.location.protocol === "https:") ? "wss://" : "ws://") +
                           window.location.host +
                           "/.gostatic.hotreload");

    ws.onmessage = function(e) {
        console.log(e.data);
        if (e.data == "page") {
            if (abort) {
                abort.abort();
            }

            abort = new AbortController();
            fetch(window.location.href,
                  {mode: 'same-origin',
                   headers: {'X-With': 'hotreload'},
                   signal: abort.signal})
                .then((res) => res.text())
                .then((text) => {
                    document.documentElement.innerHTML = text;
                    var e = new Event('load', {'bubbles': true});
                    window.dispatchEvent(e);
                })
                .catch((e) => {
                    if (!(e.message == "The operation was aborted. ")) {
                        console.log(e);
                    }
                });
        } else if (e.data == "css") {
            // This snippet pinched from quickreload, under the MIT license:
            // https://github.com/bjoerge/quickreload/blob/master/client.js
            var killcache = '__gostatic=' + new Date().getTime();
            var stylesheets = Array.prototype.slice.call(
                document.querySelectorAll('link[rel="stylesheet"]')
            );
            stylesheets.forEach(function (el) {
                var href = el.href.replace(/(&|\?)__gostatic\=\d+/, '');
                el.href = '';
                el.href = href + (href.indexOf("?") == -1 ? '?' : '&') + killcache;
            });
        }
    }

    window.addEventListener('beforeunload', function(e) {
        ws.close();
    });
})();
