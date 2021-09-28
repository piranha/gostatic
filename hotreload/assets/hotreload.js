/* jshint esversion: 6 */

(function() {
  if (!('WebSocket' in window)) return;

  var proto = (location.protocol === "https:") ? "wss://" : "ws://";

  function connect() {
    var ws = new WebSocket(proto + location.host + "/.gostatic.hotreload");

    ws.onmessage = function(e) {
      localStorage.hotreloaddebug && console.log(e.data);
      enqueue(e.data);
    };

    ws.addEventListener('close', e => setTimeout(connect, 1000));
    window.addEventListener('beforeunload', e => ws.close());
  }
  connect();


  var MESSAGES = new Set();
  var timeout, timeoutMs;

  function enqueue(msg) {
    MESSAGES.add(msg);
    // start with 32ms and double every message up to 1000
    timeoutMs = timeout ? Math.min(timeoutMs * 2, 1000) : 32;
    clearTimeout(timeout);
    timeout = setTimeout(execute, timeoutMs);
  }

  function execute() {
    localStorage.hotreloaddebug && console.log('reload', MESSAGES);
    MESSAGES.forEach(mode => RELOADERS[mode]());
    MESSAGES.clear();
    timeout = null;
  }


  var RELOADERS = {
    page: function reloadpage() {
      fetch(window.location.href,
            {mode:    'same-origin',
             headers: {'X-With': 'hotreload'}})
        .then(res => res.text())
        .then(text => {
          morphdom(document.documentElement, text);
          // document.documentElement.innerHTML = text;
          var e = new Event('hotreload', {'bubbles': true});
          window.dispatchEvent(e);
        })
        .catch(e => {
          if (e.message != "The operation was aborted. ") {
            console.log(e);
          }
        });
    },
    css: function reloadcss() {
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
    }};
})();
