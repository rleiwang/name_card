import React from 'react';
import PropTypes from 'prop-types';

import NameCard from './namecard';

class SingleLineGridList extends React.Component {
  constructor(props, context) {
    super(props, context)
    this.state = { width: 0, height: 0, rows: 0 };
    this.updateWindowDimensions = this.updateWindowDimensions.bind(this);
  }

  componentDidMount() {
    this.updateWindowDimensions();
    window.addEventListener('resize', this.updateWindowDimensions);
  }

  componentWillUnmount() {
    window.removeEventListener('resize', this.updateWindowDimensions);
  }

  updateWindowDimensions() {
    this.setState({
      width: window.innerWidth,
      height: window.innerHeight,
      rows: Math.floor((window.innerHeight - this.props.cellHeight - 1) / this.props.cellHeight)
    });
  }

  render() {
    if (this.state.width === 0 || this.state.height === 0) {
      return <div />;
    }

    return (
      <div className="wrapper">
        <style>{`:root { --af: ${this.state.rows}; }`}</style>
        {this.props.data.map(nc =>
          <NameCard key={nc.id} data={nc} rowsPerPage={this.state.rows} disabled={this.props.showAll && !nc.absent} />
        )}
      </div>
    );
  }
}

SingleLineGridList.propTypes = {
  classes: PropTypes.object.isRequired,
  cellHeight: PropTypes.number.isRequired,
  showAll: PropTypes.bool.isRequired,
};

export default SingleLineGridList;