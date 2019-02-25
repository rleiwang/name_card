import React from 'react';
import PropTypes from 'prop-types';

import { withStyles } from '@material-ui/core/styles';

import postal from 'postal';

import AppBar from '@material-ui/core/AppBar';
import Avatar from '@material-ui/core/Avatar';
import Badge from '@material-ui/core/Badge';
import deepPurple from '@material-ui/core/colors/deepPurple';
import Toolbar from '@material-ui/core/Toolbar';
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';
import Typography from '@material-ui/core/Typography';
import IconButton from '@material-ui/core/IconButton';
import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import MoreVertIcon from '@material-ui/icons/MoreVert';
import CircularProgress from '@material-ui/core/CircularProgress';
import green from '@material-ui/core/colors/green';

import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogTitle from '@material-ui/core/DialogTitle';
import Slide from '@material-ui/core/Slide';

import Disconnected from '@material-ui/icons/PortableWifiOff';
import Connected from '@material-ui/icons/Wifi';
import Keyboard from '@material-ui/icons/Keyboard';

import ScrollPaper from './gridlist';
import Admin from './admin';

function TabContainer(props) {
  return (
    <Typography component="div" style={{ height: '100vh', width: '100vw' }}>
      {props.children}
    </Typography>
  );
}

TabContainer.propTypes = {
  children: PropTypes.node.isRequired,
};

const styles = theme => ({
  root: {
    flexGrow: 1,
    height: '100%',
  },
  row: {
    display: 'flex',
    justifyContent: 'center',
    fontSize: '1ch',
  },
  purpleAvatar: {
    margin: 2,
    width: 20,
    height: 20,
    color: deepPurple[500],
  },
  buttonConnecting: {
    color: green[500],
    position: 'absolute',
    top: '50%',
    left: '50%',
    marginTop: -12,
    marginLeft: -12,
  },
  padding: {
    padding: `0 ${theme.spacing.unit * 1}px`,
  },
});

class SimpleTabs extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      data: [],
      quicklist: [],
      value: 0,
      connected: false,
      connecting: false,
      filter: n => true,
      anchorEl: null,
      expand: true,
      focused: false,
      cellHeight: 48,
    };
  }

  handleChange = (event, value) => {
    this.setState({ value });
  };

  handleConnectionClick = (event) => {
    this.setState({ connecting: true })
    if (!this.state.connected) {
      postal.publish({
        channel: "event",
        topic: "conn",
        data: { connecting: true, connected: false }
      });
    }
  }

  componentWillUnmount() {
    if (this._subscriptions) {
      this._subscriptions.array.forEach(e => e.unsubscribe());
    }
  }

  componentDidMount() {
    this._ql = [];
    this._subscriptions = [postal.subscribe({
      channel: "event",
      topic: "conn",
      callback: (data, envelope) => this.setState({ connected: data.connected, connecting: false })
    }), postal.subscribe({
      channel: "event",
      topic: "recv",
      callback: (data, envelope) => {
        if (data.type === 'Load') {
          if (data.list) {
            this._ql = Object.keys(data.list.reduce((m, l) => {
              if (l.family && l.family.length > 0) {
                if (l.family.charAt(0) >= 'A' && l.family.charAt(0) <= 'Z')
                  m[l.family.charAt(0)] = true
              }
              return m;
            }, {})).sort();
          }
          this.setState({ data: data.list, quicklist: this._ql, expand: false });
        } else if (data.type === 'Update') {
          data.list.forEach(e => this.state.data[e.idx].absent = e.absent);
          this.setState(this.state)
        }
      }
    }),
    postal.subscribe({
      channel: "event",
      topic: "key",
      callback: e => {
        if (e.key) {
          if (e.key === '^a') {
            this.setState({ value: 2 });
          } else if (e.key === 'Escape') {
            this.setState({ filter: d => true, expand: !this.state.expand, quicklist: this.state.expand ? this._ql : [] });
          } else {
            this.setState({ filter: d => d.family && d.family.startsWith(e.key.toUpperCase()) });
          }
        } else {
          this.setState({ focused: e.focused })
        }
      }
    })];
  }

  render() {
    const { classes } = this.props;
    const { value, anchorEl } = this.state;

    let cnt = this.state.data.filter(d => !d.absent).length

    return (
      <div className={classes.root}>
        <AppBar position="static">
          <Toolbar>
            <Tabs value={value} onChange={this.handleChange} style={{ flex: 1 }}>
              <Tab label="聚會會眾" />
              <Tab label={
                <Badge className={classes.padding} badgeContent={cnt} color="secondary">
                  簽到
              </Badge>} />
            </Tabs>
            <IconButton color={this.state.focused ? "inherit" : "black"}>
              <Keyboard />
            </IconButton>
            <IconButton onClick={this.handleConnectionClick} color="inherit">
              {this.state.connected ? <Connected /> : <Disconnected />}
            </IconButton>
          </Toolbar>
        </AppBar>
        {value === 0 &&
          <TabContainer>
            <ScrollPaper data={this.state.data.filter(d => this.state.filter(d))}
              cellHeight={this.state.cellHeight} showAll={true} />
          </TabContainer>
        }
        {value === 1 &&
          <TabContainer>
            <ScrollPaper data={this.state.data.filter(d => !d.absent && this.state.filter(d))}
              cellHeight={this.state.cellHeight} showAll={false} />
          </TabContainer>
        }
        {value === 2 && <TabContainer><Admin data={this.state.data} cellHeight={this.state.cellHeight} /></TabContainer>}
        {this.state.connected ? null :
          <Dialog open={true}
            TransitionComponent={props => <Slide direction="up" {...props} />}
            keepMounted={false}
            aria-labelledby="alert-dialog-slide-title"
            aria-describedby="alert-dialog-slide-description"
          >
            <DialogTitle id="alert-dialog-slide-title">
              {"Connection lost, reconnect?"}
            </DialogTitle>
            <DialogActions>
              <Button onClick={this.handleConnectionClick} color="primary" disabled={this.state.connecting}>Connect</Button>
              {this.state.connecting && <CircularProgress size={24} className={classes.buttonConnecting} />}
            </DialogActions>
          </Dialog>}
      </div>
    );
  }
}

SimpleTabs.propTypes = {
  classes: PropTypes.object.isRequired,
};

export default withStyles(styles)(SimpleTabs);