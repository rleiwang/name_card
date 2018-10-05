import React from 'react';
import PropTypes from 'prop-types';

import postal from 'postal';

class Admin extends React.Component {
    constructor(props, context) {
        super(props, context);
        this.state = {
            file: null,
        };
    }

    render() {
        const { data, classes } = this.props;

        return (<div className="previewComponent">
            <form onSubmit={(e) => this._handleSubmit(e)}>
                <input className="fileInput"
                    type="file"
                    onChange={(e) => this._handleFileSelected(e)} />
                <button className="submitButton"
                    type="submit"
                    onClick={(e) => this._handleSubmit(e)}>Upload</button>
            </form>

            <a href={'http://' + window.location.hostname + ":8080/download"} download="report.csv">download report</a>
        </div>);
    }

    _handleSubmit(e) {
        e.preventDefault();

        var formData = new FormData();
        formData.append('file', this.state.file);

        fetch('http://' + window.location.hostname + ':8080/upload', {
            method: 'POST',
            body: formData,
        }).then(response => {
            if (response.status == 200) {
                postal.publish({
                    channel: "event",
                    topic: "reconn",
                    data: {}
                });
            }
        });
    }

    _handleFileSelected(e) {
        e.preventDefault();
        this.setState({ file: e.target.files[0] });
    }
}

Admin.propTypes = {
    classes: PropTypes.object.isRequired,
    data: PropTypes.array.isRequired,
    onSelectAllClick: PropTypes.func.isRequired,
};

export default (Admin);