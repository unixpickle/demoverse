(function() {

  function Episode(env) {
    this._env = env;
    this._playing = false;
    this._timeout = null;
    this._promise = null;
    this._totalReward = 0;
    this._lastUpdate = 0;
    this._events = [];
    this._clearHandlers();
  }

  Episode.prototype.totalReward = function() {
    return this._totalReward;
  };

  Episode.prototype.pushEvent = function(e) {
    this._events.push(e);
  };

  Episode.prototype.play = function() {
    if (this._playing) {
      throw new Error('already playing');
    }
    this._playing = true;
    if (this._promise) {
      return;
    }
    this._lastUpdate = new Date().getTime();
    this._scheduleTimeout();
  };

  Episode.prototype.pause = function() {
    if (!this._playing) {
      throw new Error('not playing');
    }
    this._playing = false;
    this._cancelTimeout();
  };

  Episode.prototype.close = function() {
    this._clearHandlers();
    this._cancelTimeout();
    this._playing = false;
  };

  Episode.prototype._scheduleTimeout = function() {
    var sinceLast = new Date().getTime() - this._lastUpdate;
    var remaining = Math.max(1, window.spec.interval-sinceLast);
    this._timeout = setTimeout(() => {
      this._timeout = null;
      this._promise = this._env.step(this._events).then((msg) => {
        this._promise = null;
        if (this._playing) {
          this._scheduleTimeout();
        }
        this.onobs(msg.observation);
        this.onreward(msg.reward);
        if (msg.done) {
          this.ondone();
          this.close();
        }
      }).catch((e) => {
        this.onerror(e);
        this.close();
      });
    }, remaining);
  };

  Episode.prototype._cancelTimeout = function() {
    if (this._timeout !== null) {
      clearTimeout(this._timeout);
      this._timeout = null;
    }
  };

  Episode.prototype._clearHandlers = function() {
    this.onobs = () => false;
    this.ondone = () => false;
    this.onreward = () => false;
    this.onerror = () => false;
  };

  window.Episode = Episode;

})();
