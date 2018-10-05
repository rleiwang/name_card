import React from 'react';
import PropTypes from 'prop-types';

import postal from 'postal';

import NameCard from './namecard';

class SingleLineGridList extends React.Component {
  constructor(props, context) {
    super(props, context)
    this.state = { width: 0, height: 0 };
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
    this.setState({ width: window.innerWidth, height: window.innerHeight });
  }

  onSelectAllClick = n => {
    n.absent = !n.absent;

    postal.publish({ channel: "event", topic: "send", data: [n] });
  }

  render() {
    if (this.state.width === 0 || this.state.height === 0) {
      return <div />;
    }

    const data = this.props.data.filter(n => this.props.filter(n))
      .sort((a, b) => a.family.localeCompare(b.family));
    const rows = Math.floor((this.state.height - this.props.cellHeight - 1) / this.props.cellHeight);

    return (
      <div className="wrapper">
        <style>{`:root { --af: ${rows}; }`}</style>
        {data.map(nc =>
          <NameCard key={nc.id} data={nc} filter={this.props.filter} rowsPerPage={rows}
            onSelectAllClick={this.onSelectAllClick.bind(this, nc)} disabled={this.props.showAll && !nc.absent} />)
        }
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