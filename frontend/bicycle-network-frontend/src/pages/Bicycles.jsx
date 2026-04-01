import { useEffect, useState, useCallback } from 'react';
import { bicyclesApi } from '../api/bicycles';
import { rentalsApi } from '../api/rentals';
import { useAuth } from '../contexts/AuthContext';
import BikeCard from '../components/BikeCard';
import BikeFormModal from '../components/BikeFormModal';
import { Search, Filter, ChevronLeft, ChevronRight, Plus } from 'lucide-react';
import toast from 'react-hot-toast';

const BIKE_TYPES = ['Cross', 'Mountain Bike', 'Route'];

export default function Bicycles() {
  const { isAdmin } = useAuth();
  const [bikes, setBikes] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [typeFilter, setTypeFilter] = useState('');
  const [loading, setLoading] = useState(true);
  const [rentalMap, setRentalMap] = useState({});
  const [showCreate, setShowCreate] = useState(false);

  const fetchBikes = useCallback(async () => {
    setLoading(true);
    try {
      const params = { page, limit: 12 };
      if (typeFilter) params.type = typeFilter;

      const [bikesRes, rentalRes] = await Promise.allSettled([
        bicyclesApi.list(params),
        rentalsApi.getActive(),
      ]);

      if (bikesRes.status === 'fulfilled') {
        const d = bikesRes.value.data;
        setBikes(d.data || []);
        setTotal(d.total || 0);
        setTotalPages(d.total_pages || 1);
      }

      if (rentalRes.status === 'fulfilled' && rentalRes.value.data) {
        const rental = rentalRes.value.data;
        setRentalMap({ [rental.bicycle_id]: 'rented' });
      }
    } catch {
      toast.error('Failed to load bicycles');
    } finally {
      setLoading(false);
    }
  }, [page, typeFilter]);

  useEffect(() => {
    fetchBikes();
  }, [fetchBikes]);

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h2>Bicycles</h2>
          <p className="text-muted">{total} bicycles found</p>
        </div>
        {isAdmin && (
          <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
            <Plus size={18} /> Add Bicycle
          </button>
        )}
      </div>

      {showCreate && (
        <BikeFormModal
          onClose={() => setShowCreate(false)}
          onSubmit={async (data) => {
            await bicyclesApi.create(data);
            toast.success('Bicycle created');
            setShowCreate(false);
            fetchBikes();
          }}
        />
      )}

      {/* Filters */}
      <div className="filters-bar">
        <div className="filter-group">
          <Filter size={18} />
          <select
            value={typeFilter}
            onChange={(e) => {
              setTypeFilter(e.target.value);
              setPage(1);
            }}
          >
            <option value="">All Types</option>
            {BIKE_TYPES.map((t) => (
              <option key={t} value={t}>
                {t}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Bike Grid */}
      {loading ? (
        <div className="loading-screen">
          <div className="spinner" />
        </div>
      ) : (
        <>
          <div className="bike-grid">
            {bikes.map((bike) => (
              <BikeCard
                key={bike.id}
                bike={bike}
                rentalStatus={rentalMap[bike.id]}
              />
            ))}
            {bikes.length === 0 && (
              <div className="empty-state">
                <Search size={48} />
                <p>No bicycles match your criteria</p>
              </div>
            )}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="pagination">
              <button
                className="btn btn-outline"
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
              >
                <ChevronLeft size={18} />
                Previous
              </button>
              <span className="page-info">
                Page {page} of {totalPages}
              </span>
              <button
                className="btn btn-outline"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
                <ChevronRight size={18} />
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
