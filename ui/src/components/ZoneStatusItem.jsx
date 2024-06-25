import React, {useState} from 'react'
import {Row, Col} from 'react-bootstrap';
import Form from 'react-bootstrap/Form';
import InputGroup from 'react-bootstrap/InputGroup';

import "../stylesheets/ZoneStatusItem.css";

const DURATION_MIN = 1;
const DURATION_MAX = 60;

const ZoneStatusItem = (props) => {

    console.log("Zone status item rendered");
    return (
        <>
            <div className="zone-status-item-container">
                <div className="zone-status-item-info-container">
                    <ZoneStatusInfoRow label="Zone Label" value={props.zoneLabel} />
                    <ZoneStatusInfoRow label="Zone Status" value="On" />
                    <ZoneStatusInfoRow label="Next Event" value="13:00" />
                    <hr style={{"width": "80%"}}></hr>
                    <ManualControlComponent key={props.zoneID} zoneID={props.zoneID}/>
                </div>
            </div>
        </>
    )
}

const ZoneStatusInfoRow = (props) => {
    console.log("Zone status info row rendered")
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

const ManualControlComponent = (props) => {
    const [manualControlSwitch, setManualControlSwitch] = useState(false);
    const [manualControlDuration, setManualControlDuration] = useState(0);
    
    // TODO: Change this to hit the actual endpoint with
    // query params for zone label
    const checkZoneReady = async () => {
        const response = await fetch(`http://localhost:3000/zoneReady`);
        const data = await response.json();
        console.log("Zone ready data: ", data);
        return data[props.zoneID];
    }

    const handleManualControlToggle = async (checked) => {
        console.log("Manual control toggled to: ", checked);
        console.log("Manual control duration: ", manualControlDuration);
        if (checked) {
            console.log("Checking if zone is ready for manual control...")
            const zoneReady = await checkZoneReady();
            console.log(`${props.zoneID} ready state: `, zoneReady);
            if (zoneReady) {
                setManualControlSwitch(true);
            } else {
                console.log(`${props.zoneID} not ready for manual control`);
                return;
            }
        } else {
            setManualControlSwitch(false);
        }
    };

    console.log("Manual control component rendered");
    return (
        <>
            <p>Manual Control</p>
            <InputGroup 
                className="align-items-center">
                <Form.Check
                    type="switch"
                    id={`${props.zoneLabel}-manual-control-switch`}
                    label="Off/On"
                    checked={manualControlSwitch}
                    onChange={(e) => handleManualControlToggle(e.target.checked)}
                    disabled={manualControlDuration === 0 ? true : false}
                    style={{"marginRight": "10px"}}
                />
                <Form.Select
                    value={manualControlDuration}
                    onChange={(e) => setManualControlDuration(Number(e.target.value))}>
                    <option value={0}>Select Duration</option>
                    {
                        Array.from(Array(DURATION_MAX - DURATION_MIN + 1).keys()).map((i) => {
                            return (
                                <option key={i} value={i + DURATION_MIN}>{i + DURATION_MIN}</option>
                            )
                        })
                    }
                </Form.Select>
            </InputGroup>
        </>
    )
};

export default ZoneStatusItem