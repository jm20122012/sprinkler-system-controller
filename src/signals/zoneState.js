import { signal } from "@preact/signals-react";

const zoneList = signal([]);
const zoneStatus = signal({});

export { zoneList, zoneStatus };