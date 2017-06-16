(function() {

  function Env(ws) {
    this._error = null;
    ws.onerror = (e) => {
      this._error = e;
    };
    ws.onclose = () => {
      this._error = this._error || 'connection closed';
    };
    this._ws = ws;

    this._runningCall = false;
    this._pendingCalls = [];
  }

  Env.prototype.reset = function() {
    return this._call({type: 'reset'});
  };

  Env.prototype.step = function(events) {
    var actions = actionsForEvents(events);
    return this._call({type: 'step', actions: actions});
  };

  Env.prototype._call = function(msg) {
    if (!this._runningCall) {
      this._runningCall = true;
      var res = this._callPromise(msg);
      var callNext = () => {
        if (this._pendingCalls.length > 0) {
          var first = this._pendingCalls[0];
          this._pendingCalls.splice(0, 1);
          first();
        } else {
          this._runningCall = false;
        }
      };
      res.then(callNext).catch(callNext);
      return res;
    }

    return new Promise((resolve, reject) => {
      this._pendingCalls.push(() => {
        this._runningCall = false;
        this._call(msg).then(resolve).catch(reject);
      });
    });
  };

  Env.prototype._callPromise = function(msg) {
    return new Promise((resolve, reject) => {
      if (this._error) {
        reject(this._error);
        return;
      }
      var onError, onMessage, onClose;
      var removeEvents = () => {
        this._ws.removeEventListener('error', onError);
        this._ws.removeEventListener('message', onMessage);
        this._ws.removeEventListener('close', onClose);
      };
      onError = (e) => {
        removeEvents();
        reject(e);
      };
      onMessage = (m) => {
        removeEvents();
        var parsed;
        try {
          parsed = JSON.parse(m.data);
        } catch (e) {
          reject(e);
          return;
        }
        if (parsed.type === 'error') {
          reject(parsed.error);
        } else {
          resolve(parsed);
        }
      };
      onClose = () => {
        removeEvents();
        reject('connection closed');
      };
      this._ws.addEventListener('error', onError);
      this._ws.addEventListener('message', onMessage);
      this._ws.addEventListener('close', onClose);
      try {
        this._ws.send(JSON.stringify(msg));
      } catch (e) {
        removeEvents();
        reject(e);
        return;
      }
    });
  };

  window.connectEnv = function(name) {
    return new Promise((resolve, reject) => {
      var wsProto = 'ws://';
      if (location.protocol === 'https:') {
        wsProto = 'wss://';
      }
      var sock = new WebSocket(wsProto + location.host + '/env/' + name);
      sock.onopen = function() {
        sock.onerror = () => false;
        resolve(new Env(sock));
      };
      sock.onerror = function(e) {
        reject(e);
      };
    });
  };

  function actionsForEvents(events) {
    var actions = [];
    events.forEach((e) => {
      var mouseTypes = {
        'mousedown': 'mousePressed',
        'mouseup': 'mouseReleased',
        'mousemove': 'mouseMoved'
      };
      var keyTypes = {
        'keydown': 'keyDown',
        'keyup': 'keyUp'
      };
      if (mouseTypes[e.type] && e.button === 0) {
        actions.push({
          mouseEvent: {
            type: mouseTypes[e.type],
            x: e.offsetX,
            y: e.offsetY,
            button: 'left',
            clickCount: e.detail
          }
        });
      } else if (keyTypes[e.type] && !e.shiftKey && !e.metaKey &&
                 !e.ctrlKey && !e.altKey) {
        var text = (e.key.length === 1 ? e.key : '');
        actions.push({
          keyEvent: {
            type: keyTypes[e.type],
            text: text,
            unmodifiedText: text,
            keyIdentifier: e.keyIdentifier,
            code: e.code,
            key: e.key,
            windowsVirtualKeyCode: e.keyCode,
            nativeVirtualKeyCode: e.keyCode
          }
        });
      }
    });
    return actions;
  }

})();
