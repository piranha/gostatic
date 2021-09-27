/* jshint esversion: 6 */

(function() {
  if (!('WebSocket' in window)) return;

  var proto = (location.protocol === "https:") ? "wss://" : "ws://";
  var ws = new WebSocket(proto + location.host + "/.gostatic.hotreload");

  ws.onmessage = function(e) {
    localStorage.hotreloaddebug && console.log(e.data);
    enqueue(e.data);
  };

  window.addEventListener('beforeunload', function(e) {
    ws.close();
  });

  var MODES = new Set();
  var timeout, timeoutS;

  function enqueue(mode) {
    MODES.add(mode);
    if (!timeout) timeoutS = 32;
    if (timeout) {
      clearTimeout(timeout);
      timeoutS = Math.min(timeoutS * 2, 1000);
    }
    timeout = setTimeout(hotreload, timeoutS);
  }
  function hotreload() {
    localStorage.hotreloaddebug && console.log('reload', MODES);
    MODES.forEach(mode => RELOADERS[mode]());
    MODES = new Set();
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
