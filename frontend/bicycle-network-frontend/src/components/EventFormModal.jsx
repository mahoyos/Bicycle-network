import { useState, useEffect } from 'react';
import { X } from 'lucide-react';

const EVENT_TYPES = ['Route', 'Tour', 'Competition'];

export default function EventFormModal({ event, onClose, onSubmit }) {
  const isEdit = !!event;
  const [form, setForm] = useState({
    name: '',
    type: 'Route',
    date: '',
    description: '',
    start_location_lat: '',
    start_location_lng: '',
    end_location_lat: '',
    end_location_lng: '',
  });
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (event) {
      setForm({
        name: event.name || '',
        type: event.type || 'Route',
        date: event.date ? event.date.slice(0, 16) : '',
        description: event.description || '',
        start_location_lat: event.start_location_lat ?? '',
        start_location_lng: event.start_location_lng ?? '',
        end_location_lat: event.end_location_lat ?? '',
        end_location_lng: event.end_location_lng ?? '',
      });
    }
  }, [event]);

  const handleChange = (e) => {
    setForm((f) => ({ ...f, [e.target.name]: e.target.value }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await onSubmit({
        ...form,
        start_location_lat: parseFloat(form.start_location_lat),
        start_location_lng: parseFloat(form.start_location_lng),
        end_location_lat: parseFloat(form.end_location_lat),
        end_location_lng: parseFloat(form.end_location_lng),
      });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal modal-lg" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>{isEdit ? 'Edit Event' : 'New Event'}</h3>
          <button className="modal-close" onClick={onClose}>
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="modal-body">
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="name">Name</label>
              <div className="input-wrapper">
                <input
                  id="name"
                  name="name"
                  type="text"
                  value={form.name}
                  onChange={handleChange}
                  placeholder="Event name"
                  required
                  maxLength={200}
                />
              </div>
            </div>

            <div className="form-group">
              <label htmlFor="type">Type</label>
              <div className="input-wrapper">
                <select id="type" name="type" value={form.type} onChange={handleChange} required>
                  {EVENT_TYPES.map((t) => (
                    <option key={t} value={t}>{t}</option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          <div className="form-group">
            <label htmlFor="date">Date &amp; Time</label>
            <div className="input-wrapper">
              <input
                id="date"
                name="date"
                type="datetime-local"
                value={form.date}
                onChange={handleChange}
                required
              />
            </div>
          </div>

          <div className="form-group">
            <label htmlFor="description">Description</label>
            <div className="input-wrapper">
              <textarea
                id="description"
                name="description"
                value={form.description}
                onChange={handleChange}
                placeholder="Describe the event..."
                rows={3}
                required
              />
            </div>
          </div>

          <p className="form-section-title">Start Location</p>
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="start_location_lat">Latitude</label>
              <div className="input-wrapper">
                <input
                  id="start_location_lat"
                  name="start_location_lat"
                  type="number"
                  step="any"
                  value={form.start_location_lat}
                  onChange={handleChange}
                  placeholder="e.g. 6.2442"
                  required
                />
              </div>
            </div>
            <div className="form-group">
              <label htmlFor="start_location_lng">Longitude</label>
              <div className="input-wrapper">
                <input
                  id="start_location_lng"
                  name="start_location_lng"
                  type="number"
                  step="any"
                  value={form.start_location_lng}
                  onChange={handleChange}
                  placeholder="e.g. -75.5812"
                  required
                />
              </div>
            </div>
          </div>

          <p className="form-section-title">End Location</p>
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="end_location_lat">Latitude</label>
              <div className="input-wrapper">
                <input
                  id="end_location_lat"
                  name="end_location_lat"
                  type="number"
                  step="any"
                  value={form.end_location_lat}
                  onChange={handleChange}
                  placeholder="e.g. 6.2518"
                  required
                />
              </div>
            </div>
            <div className="form-group">
              <label htmlFor="end_location_lng">Longitude</label>
              <div className="input-wrapper">
                <input
                  id="end_location_lng"
                  name="end_location_lng"
                  type="number"
                  step="any"
                  value={form.end_location_lng}
                  onChange={handleChange}
                  placeholder="e.g. -75.5734"
                  required
                />
              </div>
            </div>
          </div>

          <div className="modal-actions">
            <button type="button" className="btn btn-outline" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={submitting}>
              {submitting ? <span className="spinner-sm" /> : isEdit ? 'Save Changes' : 'Create Event'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
