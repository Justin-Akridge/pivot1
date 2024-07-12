const JobModal = ({ showJobModal, setShowJobModal }) => {
    return showJobModal ? (
        <div className="fixed inset-0 z-50 overflow-auto bg-black bg-opacity-50 flex items-center justify-center">
            <div className="relative pt-4 bg-white bg-opacity-90 w-full h-1/3 max-w-md m-auto flex-col flex rounded-md">
                <div className="flex justify-between items-center border-b-2 pb-2">
                    <h5 className="text-lg font-bold text-black ml-4">Create New Job</h5>
                </div>
                <div className="p-4">
                    <form action="http://localhost:8080/createJob" method="POST" id="create-job-form">
                        <div className="mb-7">
                            <label htmlFor="job-name" className="block text-lg font-medium text-gray-700">
                                Job Name:
                            </label>
                            <input
                                type="text"
                                id="job-name"
                                name="job-name"
                                className="mt-1 bg-white text-black block w-full border border-black rounded-md p-3 shadow-sm focus:border-blue-500 focus:ring focus:ring-blue-500 focus:ring-opacity-50"
                                placeholder="Enter Job Name"
                                required
                            />
                        </div>
                        <div className="mb-4">
                            <label htmlFor="company-name" className="block text-lg font-medium text-gray-700">
                                Company Name:
                            </label>
                            <input
                                type="text"
                                id="company-name"
                                name="company-name"
                                className="mt-1 bg-white block text-black w-full border border-black rounded-md p-3 shadow-sm focus:border-blue-500 focus:ring focus:ring-blue-500 focus:ring-opacity-50"
                                placeholder="Enter Company Name"
                                required
                            />
                        </div>
                        {/* Modal Footer */}
                        <div className="flex justify-between mt-9">
                            <button
                                onClick={() => setShowJobModal(false)}
                                type="button"
                                className="mr-2 px-4 py-2 bg-gray-400 text-gray-700 rounded-md hover:bg-gray-300 focus:outline-none"
                            >
                                Close
                            </button>
                            <button
                                type="submit"
                                className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none"
                            >
                                Create Job
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    ) : <></>
};

export default JobModal;

