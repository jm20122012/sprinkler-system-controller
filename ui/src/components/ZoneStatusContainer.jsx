import React, {useEffect} from 'react'
import { useSelector, useDispatch } from 'react-redux';

import ZoneStatusItem from './ZoneStatusItem.jsx';

import { addZone, removeZone } from "../redux/slices/zoneStatusSlice.js";

import "../stylesheets/ZoneStatusContainer.css";

const ZoneStatusContainer = () => {
    const dispatch = useDispatch();

    const zoneList = useSelector(state => state.zoneStatus.zoneList);

    useEffect(() => {
        console.log("Fetching zone list...");
        fetch('http://localhost:3000/zoneList')
        .then(res => res.json())
        .then(data => {
            console.log(data)
            data.map(zone => {
                console.log("Adding zone: ", zone);
                dispatch(addZone(zone));
            })
        })
        .catch(err => console.err(err));
    }, [])

    useEffect(() => {console.log("Zone list updated: ", zoneList)}, [zoneList]);
    
    return (
        <>
            <div className="zone-status-container">
                {
                    zoneList && 
                    zoneList.map(zone => {
                        console.log("Rendering zone: ", zone);
                        return <ZoneStatusItem key={zone} zoneLabel={zone} />
                    })
                }
            </div>
        </>
    )
}

export default ZoneStatusContainer