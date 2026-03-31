import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';

// Fix default marker icons for Leaflet + bundlers
delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
});

const bikeIcon = new L.Icon({
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41],
});

export default function BikeMap({
  bikes = [],
  center = [4.711, -74.0721],
  zoom = 13,
  height = '500px',
  onMarkerClick,
}) {
  return (
    <MapContainer
      center={center}
      zoom={zoom}
      style={{ height, width: '100%', borderRadius: '12px' }}
    >
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      {bikes.map((bike) => {
        if (!bike.location?.latitude || !bike.location?.longitude) return null;
        return (
          <Marker
            key={bike.id}
            position={[bike.location.latitude, bike.location.longitude]}
            icon={bikeIcon}
            eventHandlers={{
              click: () => onMarkerClick?.(bike),
            }}
          >
            <Popup>
              <strong>Bike {bike.id?.slice(0, 8)}</strong>
              <br />
              Lat: {bike.location.latitude.toFixed(5)}
              <br />
              Lng: {bike.location.longitude.toFixed(5)}
            </Popup>
          </Marker>
        );
      })}
    </MapContainer>
  );
}
