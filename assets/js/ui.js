(function() {

  var MOUSE_EVENTS = ['mousedown', 'mouseup', 'mousemove'];
  var KEY_EVENTS = ['keydown', 'keyup'];

  function UI() {
    this._state = UI.INITIALIZING;
    this._episode = null;

    this._canvas = document.getElementById('canvas');
    this._error = document.getElementById('error');
    this._score = document.getElementById('score');

    this._eventHandler = (e) => this._gotEvent(e);

    var butToEvt = {
      'paused-overlay': () => this._onPlay(),
      'game-over-overlay': () => this._onReset(),
      'pause-button': () => this._onPause(),
    };
    Object.keys(butToEvt).forEach((id) => {
      document.getElementById(id).addEventListener('click', butToEvt[id]);
    });

    window.connectEnv(window.spec.name).then((env) => {
      this._env = env;
      this._setState(UI.DONE);
      this._onReset();
    }).catch((e) => {
      this._handleError(e);
    });
  }

  UI.INITIALIZING = 'state-initializing';
  UI.RESETTING = 'state-resetting';
  UI.PLAYING = 'state-playing';
  UI.PAUSED = 'state-paused';
  UI.ERROR = 'state-error';
  UI.DONE = 'state-done';

  UI.prototype._onReset = function() {
    if (this._state === UI.PLAYING ||
        this._state === UI.PAUSED) {
      this._episode.close();
      this._episode = null;
      this._unregisterEvents();
    }
    if (this._state === UI.DONE || this._state === UI.PLAYING ||
        this._state === UI.PAUSED) {
      this._setState(UI.RESETTING);
      this._env.reset().then((obj) => {
        this._setState(UI.PAUSED);
        this._showObs(obj.observation);
        this._createEpisode();
      }).catch((e) => {
        this._handleError(e);
      });
    }
  };

  UI.prototype._onPlay = function() {
    if (this._state === UI.PAUSED) {
      this._setState(UI.PLAYING);
      this._episode.play();
      this._registerEvents();
    }
  };

  UI.prototype._onPause = function() {
    if (this._state === UI.PLAYING) {
      this._setState(UI.PAUSED);
      this._episode.pause();
      this._unregisterEvents();
    }
  };

  UI.prototype._setState = function(s) {
    this._state = s;
    document.body.className = s;
  };

  UI.prototype._handleError = function(e) {
    this._error.textContent = e;
    this._setState(UI.ERROR);
  };

  UI.prototype._showObs = function(obs) {
    if (this._pendingImage) {
      this._pendingImage.onload = () => false;
    }
    this._pendingImage = new Image();
    this._pendingImage.onload = () => {
      var ctx = this._canvas.getContext('2d');
      ctx.drawImage(this._pendingImage, 0, 0);
      this._pendingImage = null;
    };
    this._pendingImage.src = 'data:image/png;base64,' + obs;
  };

  UI.prototype._createEpisode = function() {
    this._episode = new window.Episode(this._env);
    this._episode.ondone = () => {
      this._episode = null;
      this._unregisterEvents();
      this._setState(UI.DONE);
    };
    this._episode.onreward = () => {
      this._score.textContent = this._episode.totalReward();
    };
    this._episode.onobs = (obs) => this._showObs(obs);
    this._episode.onerror = (e) => this._handleError(e);
  };

  UI.prototype._gotEvent = function(e) {
    if (this._episode) {
      this._episode.pushEvent(e);
    }
  };

  UI.prototype._registerEvents = function() {
    MOUSE_EVENTS.forEach((evt) => {
      this._canvas.addEventListener(evt, this._eventHandler);
    });
    KEY_EVENTS.forEach((evt) => {
      window.addEventListener(evt, this._eventHandler);
    });
  };

  UI.prototype._unregisterEvents = function() {
    MOUSE_EVENTS.forEach((evt) => {
      this._canvas.removeEventListener(evt, this._eventHandler);
    });
    KEY_EVENTS.forEach((evt) => {
      window.removeEventListener(evt, this._eventHandler);
    });
  };

  window.addEventListener('load', function() {
    new UI();
  });

})();
