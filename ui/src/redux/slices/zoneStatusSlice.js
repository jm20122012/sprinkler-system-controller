import { createSlice } from "@reduxjs/toolkit";

const initialState = {
    zoneList: [],
    zoneStatus: [
        {
            "zoneLabel": "Zone 1",
            "zoneStatus": "Off",
            "nextEvent": "None",
        },
        {
            "zoneLabel": "Zone 2",
            "zoneStatus": "Off",
            "nextEvent": "None",
        },
        {
            "zoneLabel": "Zone 3",
            "zoneStatus": "Off",
            "nextEvent": "None",
        },
        {
            "zoneLabel": "Zone 4",
            "zoneStatus": "Off",
            "nextEvent": "None",
        }
    ]
};

const zoneStatusSlice = createSlice({
    name: "zoneStatus",
    initialState,
    reducers: {
        addZone(state, action) {
            state.zoneList.push(action.payload);
        },
        removeZone(state, action) {
            state.zoneList.filter(zone => zone !== action.payload);
        },
        updateZoneStatus(state, action) {
            state.zoneStatus = action.payload;
        },
    },
});

export const { 
    addZone,
    removeZone,
    updateZoneStatus 
} = zoneStatusSlice.actions;
export default zoneStatusSlice.reducer;
