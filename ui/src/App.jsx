import MainContent from './components/MainContent.jsx';
import Navbar from "./components/Navbar.jsx";
import Form from 'react-bootstrap/Form';
import InputGroup from 'react-bootstrap/InputGroup';
import Button from 'react-bootstrap/Button';

import './App.css'
import 'bootstrap/dist/css/bootstrap.min.css';


function App() {

  return (
    <>
      <div className="app-container">
        <Navbar />
        <MainContent />
      </div>
    </>
  )
}

export default App
