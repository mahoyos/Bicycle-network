import client from './client';

const PREFIX = import.meta.env.VITE_GEOLOCATION_PREFIX || '/api/v1';

export const geolocationApi = {
  getActiveBikes() {
    return client.get(`${PREFIX}/locations/active`);
  },

  getBikeLocation(bikeId) {
    return client.get(`${PREFIX}/locations/${bikeId}`);
  },
};
