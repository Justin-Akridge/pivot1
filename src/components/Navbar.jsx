import React, { useState, useRef, useEffect } from 'react';
import image from '../assets/image.png';
import { FaSearch } from "react-icons/fa";
import { IoCreateOutline } from "react-icons/io5";
import { FaCaretDown } from "react-icons/fa";
import { IoMdClose } from "react-icons/io";
import { FaGoogle } from "react-icons/fa";
import { IoSettingsOutline } from "react-icons/io5";
import JobModal from './JobModal';



export const Navbar = ({ map }) => {
  const [showTools, setShowTools] = useState(false);
  const [showJobModal, setShowJobModal] = useState(false);
  const [showJobSearch, setShowJobSearch] = useState(false);
  const [jobs, setJobs] = useState(false)

  const toggleJobSearch = () => {
    setShowJobSearch(!showJobSearch); // Toggle dropdown visibility
  };
  // const mapInstanceRef = useRef(null);
  // const searchBoxRef = useRef(null);
  // useEffect(() => {
  //   const mapInstance = mapInstanceRef.current;

  //   if (!mapInstance) return;

  //   // Initialize Autocomplete
  //   const input = document.getElementById('google-search-input');
  //   const autocomplete = new window.google.maps.places.Autocomplete(input);

  //   autocomplete.bindTo('bounds', mapInstance);

  //   // Handle place selection
  //   autocomplete.addListener('place_changed', () => {
  //     const place = autocomplete.getPlace();
  //     if (!place.geometry) {
  //       console.log("No details available for input: '" + place.name + "'");
  //       return;
  //     }

  //     if (place.geometry.viewport) {
  //       mapInstance.fitBounds(place.geometry.viewport);
  //     } else {
  //       mapInstance.setCenter(place.geometry.location);
  //       mapInstance.setZoom(17);
  //     }
  //   });

  //   // Cleanup listeners on component unmount
  //   return () => {
  //     window.google.maps.event.clearInstanceListeners(input);
  //   };
  // }, [mapInstanceRef.current]);

  useEffect(() => {
    const fetchJobs = async () => {
      try {
        const response = await fetch('http://localhost:8080/jobs'); // Replace with your API endpoint
        if (!response.ok) {
          throw new Error('Failed to fetch jobs');
        }
        const data = await response.json();
        console.log(data)
        setJobs(data);
      } catch (error) {
        console.error('Error fetching jobs:', error);
      }
    };

    if (showJobSearch) {
      fetchJobs();
    }
  }, [showJobSearch]);
  return (
    <>
      <div className="fixed flex items-center top-0 bottom-0 w-full h-16 bg-black bg-opacity-80 z-50 p-2">
          <img src={image} className='w-14 h-auto' alt='pivot logo' />
          <div className="flex h-full items-center justify-between" id="job-search-container">
            <div 
              onClick={toggleJobSearch} // Toggle dropdown visibility on click
              className="flex items-center ml-2 mr-2 pt-3 pb-3 cursor-pointer border border-transparent hover:border-blue-600 transition duration-300 rounded-md px-4 py-2">
              <FaSearch className='mr-3 text-2xl'/>
              <input className="w-full border-none outline-none placeholder-white-400 text-xl bg-transparent" id="google-search-input" placeholder="Search Jobs" autoComplete="off"/>
              <FaCaretDown className='color-white text-2xl cursor-pointer'/>

            </div>
          </div>
        <div className="flex items-center justify-center  border-opacity-25  cursor-pointer">
          <div onClick={() => setShowJobModal(true)} className="w-100 h-100 flex items-center justify-center overflow-hidden">
            <div className="flex items-center justify-center w-full h-full hover:border-blue-600 border border-transparent transition duration-300 rounded-md p-2">
              <IoCreateOutline className='text-white text-3xl' />
            </div>
          </div>
        </div>
          <div className="flex fixed left-1/2 transform -translate-x-1/2 h-full items-center justify-between" id="job-search-container">
            <div className="flex items-center ml-2 mr-2 pt-3 pb-3 cursor-pointer border border-transparent hover:border-blue-600 transition duration-300 rounded-md px-4 py-2">
              <FaSearch className='mr-3 text-2xl'/>
              <input className="w-full border-none outline-none placeholder-white-400 text-xl bg-transparent" id="google-search-input" placeholder="Search Map" autoComplete="off"/>
            </div>
          </div>

          <div className="w-100 h-100 fixed right-4 flex items-center justify-center overflow-hidden cursor-pointer">
            <div className="flex items-center justify-center w-full h-full hover:border-blue-600 border border-transparent transition duration-300 rounded-md p-2">
              <IoSettingsOutline className='text-white text-3xl' />
            </div>
          </div>
      </div>

      {showJobSearch && (
        <div className='fixed left-20 top-16 h-14 w-96 z-50 p-2 flex justify-start items-center bg-black bg-opacity-75 rounded-md border'>
          <div className="mt-2 text-white">List of jobs goes here...</div>
        </div>
      )}
      {/* {showJobSearch && (
        <div className='fixed left-20 top-0 h-20 pt-4 w-fit z-50 p-2 flex justify-center items-start bg-black bg-opacity-75 rounded-md'>
          <div className="flex items-center justify-between" id="job-search-container">
            <div className="flex items-center pb-4 ml-6 cursor-pointer border border-transparent hover:border-blue-600 transition duration-300 rounded-md px-4 py-2">
              <FaSearch className='ml-2 mr-3 text-xl'/>
              <input className="w-full border-none outline-none placeholder-white-400 text-lg bg-transparent" id="job-search-input" placeholder="Search Jobs" autoComplete="off"/>
              <FaCaretDown className='color-white text-2xl cursor-pointer'/>
            </div>
            <button className='ml-10 text-lg'>Create Job</button>
          </div>
        </div>
      )} */}

      <JobModal showJobModal={showJobModal} setShowJobModal={setShowJobModal} />
    
    </>
  );
}

//  /* <div id="nav-bar">
//<div id="header">
//   <img src="/assets/image.png" id="logo" alt="logo"/>
//   <span id="logo-name">Pivot</span>
// </div>
// <div id="header-upload-container">
//   <h2 id="jobs-header">Jobs</h2>
//   {showTools && (
//     <>
//       <i id="upload-button" className="fa-solid fa-upload"></i>
//       <input
//         type="file"
//         id="file-input"
//         name="file"
//         accept=".las"
//         style={{ display: 'none' }}
//       />
//     </>
//   )}
// </div>
// <div id="left-container">
//   <div className="search-container" id="job-search-container">
//     <i className="fa-solid fa-magnifying-glass"></i>
//     <input className="search" id="job-search-input" placeholder="Search Jobs" autoComplete="off"/>
//     <i id="joblist-dropdown" className="fa-solid fa-caret-down"></i>
//   </div>
//   <button id="create-job-button">create job</button>
// </div>
// <div id="middle-container">
//   <i className="fa-solid fa-magnifying-glass"></i>
//   <div className="search-container">
//     <input className="search" id="google-maps-search" placeholder="Search location..."/>
//   </div>
// </div>
// <div id="right-container">
//   <div className="avatar">
//     <span className="avatar-letter">JA</span>
//   </div>
// </div>
// </div>

// <div id="jobModal" className="modal" tabIndex="-1">
// <div className="modal-dialog">
//   <div className="modal-content">
//     <div className="modal-header">
//       <h5 className="modal-title">New Job</h5>
//       <button type="button" className="btn-close" aria-label="Close" id="close-modal"></button>
//     </div>
//     <div className="modal-body">
//       <form action="http://localhost:8080/createJob" method="POST" id="create-job-form">
//         <div className="mb-3">
//           <label htmlFor="job-name" className="form-label">Job Name:</label>
//           <input type="text" className="form-control" id="job-name" name="job-name" placeholder="Enter Job Name" required />
//         </div>
//         <div className="mb-3">
//           <label htmlFor="company-name" className="form-label">Company Name:</label>
//           <input type="text" className="form-control" id="company-name" name="company-name" placeholder="Enter Company Name" required />
//         </div>
//         <div className="modal-footer">
//           <button type="button" className="btn btn-secondary" id="cancel-modal-button">Close</button>
//           <button type="submit" className="btn btn-primary">Create Job</button>
//         </div>
//       </form>
//     </div>
//   </div>
// </div>
// </div>

// <ul id="dropdown-menu" className="hidden"></ul>

// <ul id="settings-modal">
// <a href="http://localhost:8080/logout" id="logout"><li>logout</li></a>
// </ul>

// <div id="spinner-container" className="modal-container">
// <div id="modal-header">
//   <div id="modal-title">Waiting to upload las file</div>
//   <button type="button" className="btn-close" aria-label="Close" id="loading-close-button"></button>
// </div>
// <div className="loading-content">
//   <div className="spinner"></div><span id="loading">Loading</span>
// </div>
// </div>

// <div id="replace-file-container" className="modal-container">
// <div id="replace-content">
//   <div>File already exists. Do you want to replace it?</div>
// </div>
// <div id="replace-file-button-container">
//   <button id="cancel-replace-button">cancel</button>
//   <button id="replace-button">yes</button>
// </div>
// </div> */}

