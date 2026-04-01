import { useEffect, useState, useCallback } from 'react';
import { eventsApi } from '../api/events';
import EventCard from '../components/EventCard';
import EventFormModal from '../components/EventFormModal';
import { Calendar, Filter, Search, Plus } from 'lucide-react';
import toast from 'react-hot-toast';

const EVENT_TYPES = ['Route', 'Tour', 'Competition'];

export default function Events() {
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [typeFilter, setTypeFilter] = useState('');
  const [searchName, setSearchName] = useState('');
  const [showCreate, setShowCreate] = useState(false);

  const fetchEvents = useCallback(async () => {
    setLoading(true);
    try {
      const params = { limit: 50 };
      if (typeFilter) params.type = typeFilter;
      if (searchName) params.name = searchName;

      const { data } = await eventsApi.list(params);
      setEvents(Array.isArray(data) ? data : []);
    } catch {
      toast.error('Failed to load events');
    } finally {
      setLoading(false);
    }
  }, [typeFilter, searchName]);

  useEffect(() => {
    fetchEvents();
  }, [fetchEvents]);

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h2>
            <Calendar size={24} /> Events
          </h2>
          <p className="text-muted">{events.length} events found</p>
        </div>
        <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
          <Plus size={18} /> Create Event
        </button>
      </div>

      {showCreate && (
        <EventFormModal
          onClose={() => setShowCreate(false)}
          onSubmit={async (data) => {
            await eventsApi.create(data);
            toast.success('Event created');
            setShowCreate(false);
            fetchEvents();
          }}
        />
      )}

      {/* Filters */}
      <div className="filters-bar">
        <div className="filter-group">
          <Search size={18} />
          <input
            type="text"
            placeholder="Search events..."
            value={searchName}
            onChange={(e) => setSearchName(e.target.value)}
          />
        </div>
        <div className="filter-group">
          <Filter size={18} />
          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
          >
            <option value="">All Types</option>
            {EVENT_TYPES.map((t) => (
              <option key={t} value={t}>
                {t}
              </option>
            ))}
          </select>
        </div>
      </div>

      {loading ? (
        <div className="loading-screen">
          <div className="spinner" />
        </div>
      ) : (
        <div className="events-grid">
          {events.map((event) => (
            <EventCard key={event.id} event={event} />
          ))}
          {events.length === 0 && (
            <div className="empty-state">
              <Calendar size={48} />
              <p>No events found</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
