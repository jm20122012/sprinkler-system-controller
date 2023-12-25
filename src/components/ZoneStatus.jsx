import React from 'react'

import { useSelector } from 'react-redux';

import "../stylesheets/ZoneStatus.css";

const ZoneStatus = (props) => {
    // const zoneStatus = useSelector(state => state.zoneStatus.map(zone => zone));
    return (
        <>
            <div className="zone-status-item-container">
                <p>Zone Label: {props.zone}</p>
                <p>Zone Status: </p>
                <p>Next Event: </p>
                <button>Manual On/Off</button>
            </div>
        </>
    )
}

export default ZoneStatus