import {APIProvider, Map} from '@vis.gl/react-google-maps';


const MapInstance = () => (
  <APIProvider apiKey={import.meta.env.VITE_GOOGLE_MAPS_API_KEY}>
    <Map
      style={{width: '100vw', height: '100vh'}}
      defaultCenter={{lat: 22.54992, lng: 0}}
      defaultZoom={3}
      mapTypeId= {'hybrid'}
      gestureHandling={'greedy'}
      disableDefaultUI={true}
      streetViewControl= {true}
      minZoom = {3}
      maxZoom = {23}
    />
  </APIProvider>
)


export default MapInstance;
