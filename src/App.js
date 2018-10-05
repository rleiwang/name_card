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
  }

  render() {
    return (
      <div className="App">
        <Tabs />
      </div>
    );
  }
}

export default App;
