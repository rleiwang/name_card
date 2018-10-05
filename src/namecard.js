import React from 'react';
import PropTypes from 'prop-types';

import { withStyles } from '@material-ui/core/styles';
import Button from '@material-ui/core/Button';

import Paper from '@material-ui/core/Paper';

const styles = theme => ({
    root: {
        width: '35ch',
        justify: 'center',
    },
    button: {
        margin: -5,
    }
});

class NameCard extends React.Component {
    constructor(props, context) {
        super(props, context);
    }

    render() {
        const { data, classes, disabled } = this.props;
        const bc = data.absent ? { 'backgroundColor': '#ffd699' } : { backgroundColor: '#99c2ff' };
        return (<Paper className={classes.root} style={bc}>
            <Button size='small' onClick={this.props.onSelectAllClick} disabled={disabled}>
                {data.cname}
            </Button>
            <Button size='small' className={classes.button} onClick={this.props.onSelectAllClick} disabled={disabled}>
                {data.first}
            </Button>
            <Button size='small' className={classes.button} disabled={disabled}>
                {data.family}
            </Button>
        </Paper>);
    }
}

NameCard.propTypes = {
    classes: PropTypes.object.isRequired,
    data: PropTypes.array.isRequired,
    onSelectAllClick: PropTypes.func.isRequired,
    disabled: PropTypes.bool.isRequired,
};

export default withStyles(styles)(NameCard);