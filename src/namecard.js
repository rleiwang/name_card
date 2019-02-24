import React from 'react';
import PropTypes from 'prop-types';

import postal from 'postal';

import { withStyles } from '@material-ui/core/styles';
import Button from '@material-ui/core/Button';

import Paper from '@material-ui/core/Paper';

const styles = theme => ({
    root: {
        width: '20ch',
        display: 'flex',
        flexDirection: 'column',
        justify: 'center',
        alignItems: 'center',
    },
    smallButton: {
        margin: -10,
        fontSize: '1ch',
        justify: 'center',
        alignItems: 'center',
    },
    button: {
        margin: -10,
        fontSize: '2.4ch',
        justify: 'center',
        alignItems: 'center',
    }
});

class NameCard extends React.Component {
    render() {
        const { data, classes, disabled } = this.props;
        const bc = data.absent ? { 'backgroundColor': '#ffd699' } : { backgroundColor: '#99c2ff' };
        return (<Paper className={classes.root} style={bc}>
            <Button className={classes.button} onClick={this.onSelect.bind(this, data)} disabled={disabled}>
                {data.cname}
            </Button>
            <Button size='small' className={classes.smallButton} onClick={this.onSelect.bind(this, data)} disabled={disabled}>
                {data.last + "," + data.first}
            </Button>
        </Paper>);
    }

    onSelect(n) {
        n.absent = !n.absent;
        this.setState({ state: this.state });
        postal.publish({ channel: "event", topic: "send", data: [n] });
    }
}

NameCard.propTypes = {
    classes: PropTypes.object.isRequired,
    data: PropTypes.array.isRequired,
    disabled: PropTypes.bool.isRequired,
};

export default withStyles(styles)(NameCard);