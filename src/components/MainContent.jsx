import React from 'react'
import ZoneStatusContainer from '../stylesheets/ZoneStatusContainer';

import "../stylesheets/MainContent.css";

const MainContent = () => {
  return (
    <>
        <div className="main-content-container">
            <p>Main Content</p>
            <ZoneStatusContainer />
            <ZoneStatusContainer />
            <ZoneStatusContainer />
            <ZoneStatusContainer />
        </div>
    </>
  )
}

export default MainContent