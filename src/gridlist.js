import React from 'react';
import PropTypes from 'prop-types';

import NameCard from './namecard';

class SingleLineGridList extends React.Component {
  constructor(props, context) {
    super(props, context)
    this.state = { width: 0, height: 0, rows: 0 };
    this.divref = React.createRef()
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
    let nrows = 0
    if (this.divref.current) {
      let computed = window.getComputedStyle(this.divref.current, null).getPropertyValue("line-height");
      const height = Number(computed.slice(0, -2)) * 1.8
      nrows = Math.floor((window.innerHeight - height - 1) / height)
    }
    if (nrows === 0) {
      nrows = Math.floor((window.innerHeight - this.props.cellHeight - 1) / this.props.cellHeight)
    }

    this.setState({
      width: window.innerWidth,
      height: window.innerHeight,
      rows: nrows
    });
  }

  render() {
    if (this.state.width === 0 || this.state.height === 0) {
      return <div />;
    }

    return (
      <div ref={this.divref} className="wrapper">
        <style>{`:root { --af: 10; }`}</style>
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