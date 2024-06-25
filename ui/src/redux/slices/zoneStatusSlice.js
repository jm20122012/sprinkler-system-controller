import { createSlice } from "@reduxjs/toolkit";

const initialState = {
    zoneList: {
        "zone1": {
            "zoneLabel": "Zone 1",
            "zoneStatus": "Off",
            "nextEvent": "None",
        },
        "zone2": {
            "zoneLabel": "Zone 2",
            "zoneStatus": "Off",
            "nextEvent": "None",
        },
        "zone3": {
            "zoneLabel": "Zone 3",
            "zoneStatus": "Off",
            "nextEvent": "None",
        },
        "zone4": {
            "zoneLabel": "Zone 4",
            "zoneStatus": "Off",
            "nextEvent": "None",
        }
    },
};

const zoneStatusSlice = createSlice({
    name: "zoneStatus",
    initialState,
    reducers: {
        updateZoneStatus(state, action) {
            state.zoneList[action.payload.zone].zoneStatus = action.payload.status;
        },
        updateZoneNextEvent(state, action) {
            state.zoneList[action.payload.zone].nextEvent = action.payload.nextEvent;
        }
    },
});

export const { 
    updateZoneStatus ,
    updateZoneNextEvent
} = zoneStatusSlice.actions;
export default zoneStatusSlice.reducer;
