import React, { Component } from 'react';
import postal from 'postal';

import './App.css';

import Tabs from './tabs';

function connect() {
  const ws = new WebSocket('ws://' + window.location.hostname + ':8080/ws');
  // subscribe to '' topics
  let subscriptions = [postal.subscribe({
    channel: "event",
    topic: "reconn",
    callback: () => { ws.close(); connect(); }
  })];

  ws.onopen = function () {
    subscriptions.push(postal.subscribe({
      channel: "event",
      topic: "send",
      callback: (data, envelope) => ws.send(JSON.stringify(data))
    }));

    postal.publish({
      channel: "event",
      topic: "conn",
      data: { connected: true }
    });
  };

  ws.onmessage = function (evt) {
    postal.publish({
      channel: "event",
      topic: "recv",
      data: JSON.parse(evt.data)
    });
    if (evt.data.type === 'Load') {
      sessionStorage.setItem('list', JSON.stringify(evt.data.list))
    }
  };

  ws.onclose = function () {
    postal.publish({
      channel: "event",
      topic: "conn",
      data: { connected: false }
    });

    subscriptions.forEach(s => s.unsubscribe());
  };
}

class App extends Component {
  constructor(props) {
    super(props);
    this.divRef = React.createRef();
    this._keyMap = {};
    this.__handleKeyPress = this.__handleKeyPress.bind(this);
  }

  componentDidMount() {
    this._subcription = postal.subscribe({
      channel: "event",
      topic: "conn",
      callback: function (data, envelope) {
        if (data.connecting && !data.connected) {
          connect();
        }
      }
    });
    connect();
    this.divRef.current.focus();
    // register as document level. At component level, like 
    // <div onKeydown=.. onKeyup...= /> key events won't bubbling up if child
    // componet received the event
    document.addEventListener("keydown", this.__handleKeyPress, false);
    document.addEventListener("keyup", this.__handleKeyPress, false);
  }

  componentWillUnmount() {
    document.removeEventListener("keydown", this.__handleKeyPress, false);
    document.removeEventListener("keyup", this.__handleKeyPress, false);
  }

  __handleKeyPress(e) {
    console.log(e)
    this._keyMap[e.keyCode] = e.type === 'keydown';
    if (this._keyMap[e.keyCode]) {
      if (e.keyCode === 65 && this._keyMap[17]) {
        postal.publish({
          channel: "event",
          topic: "key",
          data: { key: '^a', focused: true }
        });
      } else if (e.keyCode === 27 || (64 < e.keyCode && e.keyCode < 91)) {
        postal.publish({
          channel: "event",
          topic: "key",
          data: { key: e.key, focused: true }
        });
      }
    }
  }

  __onBlur(e) {
    postal.publish({
      channel: "event",
      topic: "key",
      data: { focused: false }
    });
  }

  __onFocus(e) {
    postal.publish({
      channel: "event",
      topic: "key",
      data: { focused: true }
    });
  }

  render() {
    return (
      <div className="App" tabIndex="0" ref={this.divRef} onFocus={this.__onFocus} onBlur={this.__onBlur}>
        <Tabs />
      </div>
    );
  }
}

export default App;
