import React from 'react';
import PropTypes from 'prop-types';

import postal from 'postal';

import Button from '@material-ui/core/Button';

class Admin extends React.Component {
    constructor(props, context) {
        super(props, context);
        this.state = {
            file: null,
        };
    }

    render() {
        return <div className="previewComponent">
            <Button variant="contained" href={'http://' + window.location.hostname + ":8080/download"} download>Download Report</Button>
            <br />
            <input className={{ display: 'none' }} id="flat-button-file" multiple type="file" onChange={this._handleFileSelected.bind(this)} />
            <label htmlFor="flat-button-file">
                <Button variant="contained" component="span" onClick={this._handleSubmit.bind(this)}>Upload</Button>
            </label>
            <br />
            <Button variant="outlined" color="secondary" onClick={this._handleReset.bind(this)}>Reset</Button>
        </div>;
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

    _handleReset(e) {
        e.preventDefault();

        fetch('http://' + window.location.hostname + ':8080/reset', { method: 'GET', })
            .then(response => {
                if (response.status == 200) {
                    postal.publish({
                        channel: "event",
                        topic: "reconn",
                        data: {}
                    });
                }
            });
    }
}

Admin.propTypes = {
    classes: PropTypes.object.isRequired,
    data: PropTypes.array.isRequired,
    onSelectAllClick: PropTypes.func.isRequired,
};

export default (Admin);