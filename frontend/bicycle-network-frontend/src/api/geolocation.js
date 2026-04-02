import createClient from './client';
const client = createClient(import.meta.env.VITE_API_GATEWAY_URL_GEOLOCATION || 'http://localhost:8080');

const PREFIX = import.meta.env.VITE_GEOLOCATION_PREFIX || '/api/v1';

export const geolocationApi = {
  getActiveBikes() {
    return client.get(`${PREFIX}/locations/active`);
  },

  getBikeLocation(bikeId) {
    return client.get(`${PREFIX}/locations/${bikeId}`);
  },
};
