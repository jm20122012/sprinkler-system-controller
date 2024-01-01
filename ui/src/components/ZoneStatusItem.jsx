import React, {useState} from 'react'
import {Row, Col} from 'react-bootstrap';
import Button from 'react-bootstrap/Button';
import Form from 'react-bootstrap/Form';
import InputGroup from 'react-bootstrap/InputGroup';

import { useSelector } from 'react-redux';

import "../stylesheets/ZoneStatusItem.css";

const ZoneStatusItem = (props) => {
    const [manualControlSwitch, setManualControlSwitch] = useState(false);
    const [manualControlMinutes, setManualControlMinutes] = useState(1);
    
    const handleManualControlToggle = async (checked) => {
        console.log("Manual control toggled to: ", checked);
        console.log("Manual control minutes: ", manualControlMinutes);
        if (checked) {
            console.log("Checking if zone is ready for manual control...")
            const zoneReady = await checkZoneReady();
            console.log("Zone ready: ", zoneReady);
            if (zoneReady) {
                setManualControlSwitch(true);
            } else {
                console.log("Zone not ready for manual control");
                return;
            }
        } else {
            setManualControlSwitch(false);
        }
    };

    // TODO: Change this to hit the actual endpoint with
    // query params for zone label
    const checkZoneReady = async () => {
        const response = await fetch(`http://localhost:3000/zoneReady`);
        const data = await response.json();
        return data[props.zoneLabel];
    }

    return (
        <>
            <div className="zone-status-item-container">
                <div className="zone-status-item-info-container">
                    <ZoneStatusInfoRow label="Zone Label" value={props.zoneLabel} />
                    <ZoneStatusInfoRow label="Zone Status" value="On" />
                    <ZoneStatusInfoRow label="Next Event" value="13:00" />
                    <p>Manual Control</p>
                    <Row className="align-items-center" style={{"width": "100%"}}> {/* Align items center */}
                        <Col>
                            <InputGroup>
                                <Form.Check
                                    type="switch"
                                    id={`${props.zoneLabel}-manual-control-switch`}
                                    label="Off/On"
                                    checked={manualControlSwitch}
                                    onChange={(e) => handleManualControlToggle(e.target.checked)}
                                />
                            </InputGroup>
                        </Col>
                        <Col>
                            <InputGroup>
                                <Form.Control
                                    type="number"
                                    min="1"
                                    max="60"
                                    placeholder="Duration (minutes)"
                                    aria-label="Duration"
                                    aria-describedby="basic-addon2"
                                    onChange={(e) => setManualControlMinutes(e.target.value)}
                                />
                            </InputGroup>
                        </Col>
                    </Row>
                </div>
            </div>
        </>
    )
}

const ZoneStatusInfoRow = (props) => {
    return (
        <>
            <Row style={{"width": "100%"}}>
                <Col>
                    <p>{props.label}</p>
                </Col>
                <Col>
                    <p>{props.value}</p>
                </Col>
            </Row>
        </>
    )
};

export default ZoneStatusItem