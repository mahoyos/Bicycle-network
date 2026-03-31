import { useState, useEffect } from 'react';
import { X } from 'lucide-react';

const BIKE_TYPES = ['Cross', 'Mountain Bike', 'Route'];

export default function BikeFormModal({ bike, onClose, onSubmit }) {
  const isEdit = !!bike;
  const [form, setForm] = useState({
    brand: '',
    type: 'Cross',
    color: '',
  });
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (bike) {
      setForm({ brand: bike.brand, type: bike.type, color: bike.color });
    }
  }, [bike]);

  const handleChange = (e) => {
    setForm((f) => ({ ...f, [e.target.name]: e.target.value }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await onSubmit(form);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>{isEdit ? 'Edit Bicycle' : 'New Bicycle'}</h3>
          <button className="modal-close" onClick={onClose}>
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="modal-body">
          <div className="form-group">
            <label htmlFor="brand">Brand</label>
            <div className="input-wrapper">
              <input
                id="brand"
                name="brand"
                type="text"
                value={form.brand}
                onChange={handleChange}
                placeholder="e.g. Trek, Specialized"
                required
                maxLength={100}
              />
            </div>
          </div>

          <div className="form-group">
            <label htmlFor="type">Type</label>
            <div className="input-wrapper">
              <select
                id="type"
                name="type"
                value={form.type}
                onChange={handleChange}
                required
              >
                {BIKE_TYPES.map((t) => (
                  <option key={t} value={t}>{t}</option>
                ))}
              </select>
            </div>
          </div>

          <div className="form-group">
            <label htmlFor="color">Color</label>
            <div className="input-wrapper">
              <input
                id="color"
                name="color"
                type="text"
                value={form.color}
                onChange={handleChange}
                placeholder="e.g. Red, Blue, Black"
                required
                maxLength={100}
              />
            </div>
          </div>

          <div className="modal-actions">
            <button type="button" className="btn btn-outline" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={submitting}>
              {submitting ? <span className="spinner-sm" /> : isEdit ? 'Save Changes' : 'Create Bicycle'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
