import { Bike } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const TYPE_COLORS = {
  Cross: '#10b981',
  'Mountain Bike': '#f59e0b',
  Route: '#6366f1',
};

export default function BikeCard({ bike, rentalStatus }) {
  const navigate = useNavigate();

  const isRented = rentalStatus === 'rented';

  return (
    <div
      className="bike-card"
      onClick={() => navigate(`/bicycles/${bike.id}`)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => e.key === 'Enter' && navigate(`/bicycles/${bike.id}`)}
    >
      <div className="bike-card-header">
        <div
          className="bike-type-badge"
          style={{ backgroundColor: TYPE_COLORS[bike.type] || '#6366f1' }}
        >
          {bike.type}
        </div>
        <div className={`availability-dot ${isRented ? 'rented' : 'available'}`} />
      </div>

      <div className="bike-card-icon">
        <Bike size={48} strokeWidth={1.5} />
      </div>

      <div className="bike-card-body">
        <h3 className="bike-brand">{bike.brand}</h3>
        <div className="bike-color">
          <span
            className="color-swatch"
            style={{ backgroundColor: bike.color?.toLowerCase() }}
          />
          <span>{bike.color}</span>
        </div>
      </div>

      <div className="bike-card-footer">
        <span className={`status-label ${isRented ? 'rented' : 'available'}`}>
          {isRented ? 'Currently Rented' : 'Available'}
        </span>
      </div>
    </div>
  );
}
