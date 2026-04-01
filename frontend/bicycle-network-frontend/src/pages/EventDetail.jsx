import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { eventsApi } from '../api/events';
import { useAuth } from '../contexts/AuthContext';
import BikeMap from '../components/BikeMap';
import EventFormModal from '../components/EventFormModal';
import {
  ArrowLeft,
  Calendar,
  MapPin,
  Tag,
  FileText,
  UserPlus,
  UserMinus,
  Pencil,
  Trash2,
} from 'lucide-react';
import toast from 'react-hot-toast';

export default function EventDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [event, setEvent] = useState(null);
  const [loading, setLoading] = useState(true);
  const [registered, setRegistered] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [showEdit, setShowEdit] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const { data } = await eventsApi.getById(id);
        setEvent(data);

        // Check if user is registered
        if (user?.id) {
          try {
            const regRes = await eventsApi.getUserRegistrations(user.id);
            const regs = regRes.data || [];
            setRegistered(regs.some((r) => r.event_id === Number(id)));
          } catch {
            // Ignore - user may not have registrations
          }
        }
      } catch {
        toast.error('Event not found');
        navigate('/events');
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [id, navigate, user?.id]);

  const handleRegister = async () => {
    setSubmitting(true);
    try {
      await eventsApi.register(id, user.id);
      setRegistered(true);
      toast.success('Registered for event!');
    } catch (err) {
      const msg = err.response?.data?.detail || 'Failed to register';
      toast.error(msg);
    } finally {
      setSubmitting(false);
    }
  };

  const handleUnregister = async () => {
    setSubmitting(true);
    try {
      await eventsApi.unregister(id, user.id);
      setRegistered(false);
      toast.success('Unregistered from event');
    } catch {
      toast.error('Failed to unregister');
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="loading-screen">
        <div className="spinner" />
      </div>
    );
  }

  if (!event) return null;

  const handleDelete = async () => {
    if (!window.confirm('Are you sure you want to delete this event?')) return;
    setDeleting(true);
    try {
      await eventsApi.delete(id);
      toast.success('Event deleted');
      navigate('/events');
    } catch {
      toast.error('Failed to delete event');
    } finally {
      setDeleting(false);
    }
  };

  const eventDate = new Date(event.date);
  const startMarker = {
    id: 'start',
    location: {
      latitude: event.start_location_lat,
      longitude: event.start_location_lng,
    },
  };
  const endMarker = {
    id: 'end',
    location: {
      latitude: event.end_location_lat,
      longitude: event.end_location_lng,
    },
  };

  return (
    <div className="page event-detail-page">
      <button className="back-btn" onClick={() => navigate('/events')}>
        <ArrowLeft size={20} />
        Back to Events
      </button>

      <div className="admin-actions">
        <button className="btn btn-outline" onClick={() => setShowEdit(true)}>
          <Pencil size={16} /> Edit
        </button>
        <button className="btn btn-danger" onClick={handleDelete} disabled={deleting}>
          {deleting ? <span className="spinner-sm" /> : <><Trash2 size={16} /> Delete</>}
        </button>
      </div>

      {showEdit && (
        <EventFormModal
          event={event}
          onClose={() => setShowEdit(false)}
          onSubmit={async (data) => {
            const { data: updated } = await eventsApi.update(id, data);
            setEvent(updated);
            toast.success('Event updated');
            setShowEdit(false);
          }}
        />
      )}

      <div className="event-detail-grid">
        <div className="event-detail-info">
          <h2>{event.name}</h2>

          <div className="detail-fields">
            <div className="detail-field">
              <Tag size={18} />
              <div>
                <span className="field-label">Type</span>
                <span className="field-value">{event.type}</span>
              </div>
            </div>

            <div className="detail-field">
              <Calendar size={18} />
              <div>
                <span className="field-label">Date</span>
                <span className="field-value">
                  {eventDate.toLocaleDateString('en-US', {
                    weekday: 'long',
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric',
                  })}
                  {' at '}
                  {eventDate.toLocaleTimeString('en-US', {
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </span>
              </div>
            </div>

            <div className="detail-field">
              <MapPin size={18} />
              <div>
                <span className="field-label">Start Location</span>
                <span className="field-value">
                  {event.start_location_lat.toFixed(5)},{' '}
                  {event.start_location_lng.toFixed(5)}
                </span>
              </div>
            </div>

            <div className="detail-field">
              <MapPin size={18} />
              <div>
                <span className="field-label">End Location</span>
                <span className="field-value">
                  {event.end_location_lat.toFixed(5)},{' '}
                  {event.end_location_lng.toFixed(5)}
                </span>
              </div>
            </div>

            <div className="detail-field">
              <FileText size={18} />
              <div>
                <span className="field-label">Description</span>
                <span className="field-value">{event.description}</span>
              </div>
            </div>
          </div>

          {/* Registration Actions */}
          <div className="event-actions">
            {registered ? (
              <button
                className="btn btn-warning btn-full"
                onClick={handleUnregister}
                disabled={submitting}
              >
                {submitting ? (
                  <span className="spinner-sm" />
                ) : (
                  <>
                    <UserMinus size={18} />
                    Unregister from Event
                  </>
                )}
              </button>
            ) : (
              <button
                className="btn btn-primary btn-full"
                onClick={handleRegister}
                disabled={submitting}
              >
                {submitting ? (
                  <span className="spinner-sm" />
                ) : (
                  <>
                    <UserPlus size={18} />
                    Register for Event
                  </>
                )}
              </button>
            )}
          </div>
        </div>

        {/* Map Section */}
        <div className="event-detail-map">
          <h3>
            <MapPin size={20} /> Route Map
          </h3>
          <BikeMap
            bikes={[startMarker, endMarker]}
            center={[event.start_location_lat, event.start_location_lng]}
            zoom={13}
            height="400px"
          />
          <div className="map-legend">
            <span>Markers show start and end locations</span>
          </div>
        </div>
      </div>
    </div>
  );
}
