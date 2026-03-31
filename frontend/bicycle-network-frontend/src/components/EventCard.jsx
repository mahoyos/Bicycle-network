import { Calendar, MapPin } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const TYPE_COLORS = {
  Route: '#6366f1',
  Tour: '#10b981',
  Competition: '#ef4444',
};

export default function EventCard({ event }) {
  const navigate = useNavigate();
  const eventDate = new Date(event.date);

  return (
    <div
      className="event-card"
      onClick={() => navigate(`/events/${event.id}`)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => e.key === 'Enter' && navigate(`/events/${event.id}`)}
    >
      <div className="event-card-header">
        <div
          className="event-type-badge"
          style={{ backgroundColor: TYPE_COLORS[event.type] || '#6366f1' }}
        >
          {event.type}
        </div>
      </div>

      <div className="event-card-body">
        <h3 className="event-name">{event.name}</h3>
        <p className="event-description">{event.description}</p>

        <div className="event-meta">
          <div className="event-meta-item">
            <Calendar size={16} />
            <span>
              {eventDate.toLocaleDateString('en-US', {
                weekday: 'short',
                year: 'numeric',
                month: 'short',
                day: 'numeric',
              })}
            </span>
          </div>
          <div className="event-meta-item">
            <MapPin size={16} />
            <span>
              {event.start_location_lat.toFixed(4)},{' '}
              {event.start_location_lng.toFixed(4)}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
