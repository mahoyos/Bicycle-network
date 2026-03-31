import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { bicyclesApi } from '../api/bicycles';
import { rentalsApi } from '../api/rentals';
import { geolocationApi } from '../api/geolocation';
import { useAuth } from '../contexts/AuthContext';
import BikeMap from '../components/BikeMap';
import BikeFormModal from '../components/BikeFormModal';
import { Bike, ArrowLeft, Clock, MapPin, Tag, Palette, Pencil, Trash2 } from 'lucide-react';
import toast from 'react-hot-toast';

export default function BikeDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { isAdmin, user } = useAuth();
  const [bike, setBike] = useState(null);
  const [location, setLocation] = useState(null);
  const [activeRental, setActiveRental] = useState(null);
  const [loading, setLoading] = useState(true);
  const [renting, setRenting] = useState(false);
  const [finalizing, setFinalizing] = useState(false);
  const [showEdit, setShowEdit] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [bikeRes, locationRes, rentalRes] = await Promise.allSettled([
          bicyclesApi.getById(id),
          geolocationApi.getBikeLocation(id),
          rentalsApi.getActive(),
        ]);

        if (bikeRes.status === 'fulfilled') {
          setBike(bikeRes.value.data);
        } else {
          toast.error('Bicycle not found');
          navigate('/bicycles');
          return;
        }

        if (locationRes.status === 'fulfilled') {
          setLocation(locationRes.value.data);
        }

        if (rentalRes.status === 'fulfilled') {
          setActiveRental(rentalRes.value.data);
        }
      } catch {
        toast.error('Failed to load bicycle details');
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [id, navigate]);

  const handleRent = async () => {
    setRenting(true);
    try {
      const { data } = await rentalsApi.create(id, user.id);
      setActiveRental(data);
      toast.success('Bicycle rented successfully!');
    } catch (err) {
      const msg = err.response?.data?.error || err.response?.data?.detail || 'Failed to rent bicycle';
      toast.error(msg);
    } finally {
      setRenting(false);
    }
  };

  const handleFinalize = async () => {
    if (!activeRental) return;
    setFinalizing(true);
    try {
      await rentalsApi.finalize(activeRental.id);
      setActiveRental(null);
      toast.success('Rental finalized!');
    } catch (err) {
      const msg = err.response?.data?.error || 'Failed to finalize rental';
      toast.error(msg);
    } finally {
      setFinalizing(false);
    }
  };

  if (loading) {
    return (
      <div className="loading-screen">
        <div className="spinner" />
      </div>
    );
  }

  if (!bike) return null;

  const handleDelete = async () => {
    if (!window.confirm('Are you sure you want to delete this bicycle?')) return;
    setDeleting(true);
    try {
      await bicyclesApi.delete(id);
      toast.success('Bicycle deleted');
      navigate('/bicycles');
    } catch {
      toast.error('Failed to delete bicycle');
    } finally {
      setDeleting(false);
    }
  };

  const isThisBikeRented = activeRental?.bicycle_id === id;
  const hasActiveRental = !!activeRental;

  const mapBikes = location
    ? [{ id: bike.id, location: { latitude: location.latitude, longitude: location.longitude } }]
    : [];

  return (
    <div className="page bike-detail-page">
      <button className="back-btn" onClick={() => navigate('/bicycles')}>
        <ArrowLeft size={20} />
        Back to Bicycles
      </button>

      {isAdmin && (
        <div className="admin-actions">
          <button className="btn btn-outline" onClick={() => setShowEdit(true)}>
            <Pencil size={16} /> Edit
          </button>
          <button className="btn btn-danger" onClick={handleDelete} disabled={deleting}>
            {deleting ? <span className="spinner-sm" /> : <><Trash2 size={16} /> Delete</>}
          </button>
        </div>
      )}

      {showEdit && (
        <BikeFormModal
          bike={bike}
          onClose={() => setShowEdit(false)}
          onSubmit={async (data) => {
            const { data: updated } = await bicyclesApi.update(id, data);
            setBike(updated);
            toast.success('Bicycle updated');
            setShowEdit(false);
          }}
        />
      )}

      <div className="bike-detail-grid">
        <div className="bike-detail-info">
          <div className="bike-detail-header">
            <div className="bike-detail-icon">
              <Bike size={64} strokeWidth={1.2} />
            </div>
            <div>
              <h2>{bike.brand}</h2>
              <span className="bike-detail-id">ID: {bike.id?.slice(0, 12)}...</span>
            </div>
          </div>

          <div className="detail-fields">
            <div className="detail-field">
              <Tag size={18} />
              <div>
                <span className="field-label">Type</span>
                <span className="field-value">{bike.type}</span>
              </div>
            </div>

            <div className="detail-field">
              <Palette size={18} />
              <div>
                <span className="field-label">Color</span>
                <span className="field-value">
                  <span
                    className="color-swatch-lg"
                    style={{ backgroundColor: bike.color?.toLowerCase() }}
                  />
                  {bike.color}
                </span>
              </div>
            </div>

            <div className="detail-field">
              <Clock size={18} />
              <div>
                <span className="field-label">Added</span>
                <span className="field-value">
                  {new Date(bike.created_at).toLocaleDateString('en-US', {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric',
                  })}
                </span>
              </div>
            </div>

            {location && (
              <div className="detail-field">
                <MapPin size={18} />
                <div>
                  <span className="field-label">Location</span>
                  <span className="field-value">
                    {location.latitude.toFixed(5)}, {location.longitude.toFixed(5)}
                  </span>
                </div>
              </div>
            )}
          </div>

          {/* Rental Actions */}
          <div className="rental-actions">
            {isThisBikeRented ? (
              <>
                <div className="rental-active-banner">
                  <Clock size={20} />
                  <div>
                    <strong>You are currently renting this bike</strong>
                    <p>{activeRental.duration_so_far || 'Rental in progress'}</p>
                  </div>
                </div>
                <button
                  className="btn btn-warning btn-full"
                  onClick={handleFinalize}
                  disabled={finalizing}
                >
                  {finalizing ? <span className="spinner-sm" /> : 'Return Bicycle'}
                </button>
              </>
            ) : (
              <button
                className="btn btn-primary btn-full"
                onClick={handleRent}
                disabled={renting || hasActiveRental}
              >
                {renting ? (
                  <span className="spinner-sm" />
                ) : hasActiveRental ? (
                  'You have an active rental'
                ) : (
                  'Rent This Bicycle'
                )}
              </button>
            )}
          </div>
        </div>

        {/* Map Section */}
        <div className="bike-detail-map">
          <h3>
            <MapPin size={20} /> Location
          </h3>
          {mapBikes.length > 0 ? (
            <BikeMap
              bikes={mapBikes}
              center={[location.latitude, location.longitude]}
              zoom={15}
              height="400px"
            />
          ) : (
            <div className="no-location">
              <MapPin size={48} />
              <p>Location data unavailable</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
