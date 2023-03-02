/* jshint esversion: 6 */

(function() {
  function esconnect() {
    var es = new EventSource('/.gostatic.hotreload');
    es.onmessage = function(e) {
      // console.log(e);
      localStorage.hotreloaddebug && console.log(e.data);
      enqueue(e.data);
    }
    window.addEventListener('beforeunload', _ => es.close());
  }
  esconnect();

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
    start() {
      console.log('hotreload connection established');
    },
    page() {
      fetch(window.location.href, {mode:    'same-origin',
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
    css() {
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
