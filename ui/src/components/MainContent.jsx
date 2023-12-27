import React from 'react'
import ZoneStatusContainer from './ZoneStatusContainer.jsx';
import UpcomingEventsContainer from './UpcomingEventsContainer.jsx';

// CSS Imports
import "../stylesheets/MainContent.css";

const MainContent = () => {
  return (
    <>
        <div className="main-content-container">
            {/* <p>Main Content</p> */}
            <ZoneStatusContainer />
            <UpcomingEventsContainer />
        </div>
    </>
  )
}

export default MainContent