import React, {useState, useEffect} from 'react'
import Table from 'react-bootstrap/Table';

import "../stylesheets/UpcomingEventsContainer.css";

const UpcomingEventsContainer = () => {
  const [upcomingEvents, setUpcomingEvents] = useState([]);

  const fetchUpcomingEvents = async () => {
    const resp = await fetch("http://localhost:3000/schedule");
    const d = await resp.json();
    console.log("Schedule: ", d);
    return d;
  };

  useEffect(() => {
    console.log("Upcoming events rendered");
    const fetchData = async () => {
      const data = await fetchUpcomingEvents();
      setUpcomingEvents(data);
    };
    fetchData();
  }, []);

  useEffect(() => {
    console.log("Upcoming events updated: ", upcomingEvents);
  },[upcomingEvents]);

  return (
    <>
      <div className="upcoming-events-container d-flex flex-column">
        <h4>Upcoming Events</h4>
        {upcomingEvents && upcomingEvents.length > 0 ? (
          <UpcomingEventsTableComponent upcomingEvents={upcomingEvents}/>
        ) : (
          <p className="text-white">No upcoming events</p>
        )}
      </div>
    </>
  )
}

const UpcomingEventsTableComponent = (props) => {
  return (
    <>
      <div className="d-flex w-100">
        <Table striped bordered hover variant='dark' className="m-3">
          <thead>
            <tr>
              <th>Event ID</th>
              <th>Zone ID</th>
              <th>Start Time</th>
              <th>Stop Time</th>
              <th>Duration (minutes)</th>
              <th>Active</th>
            </tr>
          </thead>
          <tbody>
            {
              props.upcomingEvents.map((ue, idx) => {
                return (
                  <tr key={`${ue.zoneID}-${idx}`}>
                    <td>{ue.eventID}</td>
                    <td>{ue.zoneID}</td>
                    <td>{ue.startTime}</td>
                    <td>{ue.stopTime}</td>
                    <td>{ue.duration}</td>
                    <td>{ue.active ? "TRUE" : "FALSE"}</td>
                  </tr>
                )
              })
            }
          </tbody>
        </Table>
      </div>
    </>
  )
};

export default UpcomingEventsContainer;