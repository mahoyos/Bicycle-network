import { useEffect, useState } from 'react';
import { geolocationApi } from '../api/geolocation';
import { useNavigate } from 'react-router-dom';
import BikeMap from '../components/BikeMap';
import { MapPin, RefreshCw } from 'lucide-react';
import toast from 'react-hot-toast';

export default function MapPage() {
  const [bikes, setBikes] = useState([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  const fetchLocations = async () => {
    setLoading(true);
    try {
      const { data } = await geolocationApi.getActiveBikes();
      setBikes(Array.isArray(data) ? data : []);
    } catch {
      toast.error('Failed to load bike locations');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLocations();
  }, []);

  const handleMarkerClick = (bike) => {
    navigate(`/bicycles/${bike.id}`);
  };

  return (
    <div className="page map-page">
      <div className="page-header">
        <div>
          <h2>
            <MapPin size={24} /> Bike Map
          </h2>
          <p className="text-muted">
            {bikes.length} bikes with location data
          </p>
        </div>
        <button className="btn btn-outline" onClick={fetchLocations} disabled={loading}>
          <RefreshCw size={18} className={loading ? 'spin' : ''} />
          Refresh
        </button>
      </div>

      <div className="map-container">
        {loading ? (
          <div className="loading-screen">
            <div className="spinner" />
          </div>
        ) : (
          <BikeMap
            bikes={bikes}
            height="calc(100vh - 200px)"
            onMarkerClick={handleMarkerClick}
          />
        )}
      </div>
    </div>
  );
}
