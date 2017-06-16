(function() {

  function UI() {
    this._state = UI.INITIALIZING;
    this._episode = null;

    this._canvas = document.getElementById('canvas');
    this._error = document.getElementById('error');
    this._score = document.getElementById('score');

    var butToEvt = {
      'play-button': () => this._onPlay(),
      'pause-button': () => this._onPause(),
      'reset-button': () => this._onReset(),
    };
    Object.keys(butToEvt).forEach((id) => {
      document.getElementById(id).addEventListener('click', butToEvt[id]);
    });

    window.connectEnv(window.spec.name).then((env) => {
      this._env = env;
      this._setState(UI.NEEDS_RESET);
    }).catch((e) => {
      this._handleError(e);
    });
  }

  UI.INITIALIZING = 'state-initializing';
  UI.NEEDS_RESET = 'state-needs-reset';
  UI.RESETTING = 'state-resetting';
  UI.PLAYING = 'state-playing';
  UI.PAUSED = 'state-paused';
  UI.ERROR = 'state-error';

  UI.prototype._onReset = function() {
    if (this._state === UI.PLAYING ||
        this._state === UI.PAUSED) {
      this._episode.close();
      this._episode = null;
      this._setState(UI.NEEDS_RESET);
    }
    if (this._state === UI.NEEDS_RESET) {
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
    }
  };

  UI.prototype._onPause = function() {
    if (this._state === UI.PLAYING) {
      this._setState(UI.PAUSED);
      this._episode.pause();
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
    var ctx = this._canvas.getContext('2d');
    var image = new Image();
    image.onload = function() {
      ctx.drawImage(image, 0, 0);
    };
    image.src = 'data:image/png;base64,' + obs;
  };

  UI.prototype._createEpisode = function() {
    this._episode = new window.Episode(this._env);
    this._episode.ondone = () => {
      this._episode = null;
      this._setState(UI.NEEDS_RESET);
    };
    this._episode.onreward = () => {
      this._score.textContent = this._episode.totalReward();
    };
    this._episode.onobs = (obs) => this._showObs(obs);
    this._episode.onerror = (e) => this._handleError(e);
  };

  window.addEventListener('load', function() {
    new UI();
  });

})();
