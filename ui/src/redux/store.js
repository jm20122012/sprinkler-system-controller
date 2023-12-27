// src/app/store.js
import { configureStore } from '@reduxjs/toolkit';

import zoneStatusReducer from './slices/zoneStatusSlice.js';

export const store = configureStore({
  reducer: {
    zoneStatus: zoneStatusReducer,
  },
});

export default store;
