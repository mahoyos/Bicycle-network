import { useEffect, useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { bicyclesApi } from '../api/bicycles';
import { rentalsApi } from '../api/rentals';
import { eventsApi } from '../api/events';
import { Bike, Clock, Calendar, MapPin, ArrowRight } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';

export default function Dashboard() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [stats, setStats] = useState({ totalBikes: 0, events: 0 });
  const [activeRental, setActiveRental] = useState(null);
  const [recentBikes, setRecentBikes] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [bikesRes, rentalRes, eventsRes] = await Promise.allSettled([
          bicyclesApi.list({ page: 1, limit: 4 }),
          rentalsApi.getActive(),
          eventsApi.list({ limit: 5 }),
        ]);

        if (bikesRes.status === 'fulfilled') {
          setRecentBikes(bikesRes.value.data.data || []);
          setStats((s) => ({ ...s, totalBikes: bikesRes.value.data.total || 0 }));
        }

        if (rentalRes.status === 'fulfilled') {
          setActiveRental(rentalRes.value.data);
        }

        if (eventsRes.status === 'fulfilled') {
          const evts = eventsRes.value.data;
          setStats((s) => ({ ...s, events: Array.isArray(evts) ? evts.length : 0 }));
        }
      } catch {
        toast.error('Failed to load dashboard data');
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  if (loading) {
    return (
      <div className="loading-screen">
        <div className="spinner" />
      </div>
    );
  }

  return (
    <div className="page dashboard-page">
      <div className="page-header">
        <h2>Welcome back, {user?.email?.split('@')[0]}!</h2>
        <p className="text-muted">Here&apos;s what&apos;s happening with BikeNet today.</p>
      </div>

      {/* Stats Cards */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-icon blue">
            <Bike size={24} />
          </div>
          <div className="stat-info">
            <span className="stat-value">{stats.totalBikes}</span>
            <span className="stat-label">Available Bikes</span>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon green">
            <Calendar size={24} />
          </div>
          <div className="stat-info">
            <span className="stat-value">{stats.events}</span>
            <span className="stat-label">Upcoming Events</span>
          </div>
        </div>

        <div className="stat-card clickable" onClick={() => navigate('/map')}>
          <div className="stat-icon purple">
            <MapPin size={24} />
          </div>
          <div className="stat-info">
            <span className="stat-value">Live</span>
            <span className="stat-label">Bike Map</span>
          </div>
        </div>
      </div>

      {/* Active Rental */}
      {activeRental && (
        <div className="section">
          <h3 className="section-title">Your Active Rental</h3>
          <div className="active-rental-card">
            <div className="rental-info">
              <Bike size={32} />
              <div>
                <p className="rental-bike-id">
                  Bike: {activeRental.bicycle_id?.slice(0, 8)}...
                </p>
                <div className="rental-time">
                  <Clock size={16} />
                  <span>
                    {activeRental.duration_so_far || 'Just started'}
                  </span>
                </div>
              </div>
            </div>
            <div className="active-rental-actions">
              <button
                className="btn btn-warning"
                onClick={() => navigate('/bicycles')}
              >
                View Details
              </button>
              <button
                className="btn btn-primary"
                onClick={async () => {
                  try {
                    await rentalsApi.finalize(activeRental.id);
                    toast.success('Rental returned successfully');
                    setActiveRental(null);
                  } catch (err) {
                    console.error('Rental return failed:', err);
                    toast.error('Failed to return the rental');
                  }
                }}
              >
                Return Rental
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Recent Bikes */}
      <div className="section">
        <div className="section-header">
          <h3 className="section-title">Recent Bicycles</h3>
          <button className="link-btn" onClick={() => navigate('/bicycles')}>
            View all <ArrowRight size={16} />
          </button>
        </div>
        <div className="mini-bike-grid">
          {recentBikes.map((bike) => (
            <div
              key={bike.id}
              className="mini-bike-card"
              onClick={() => navigate(`/bicycles/${bike.id}`)}
            >
              <Bike size={28} />
              <div>
                <strong>{bike.brand}</strong>
                <span className="text-muted">{bike.type}</span>
              </div>
            </div>
          ))}
          {recentBikes.length === 0 && (
            <p className="text-muted">No bicycles available yet.</p>
          )}
        </div>
      </div>
    </div>
  );
}
