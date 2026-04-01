import createClient from './client';
const client = createClient(import.meta.env.VITE_API_GATEWAY_URL_RENTALS || 'http://localhost:8080');

const PREFIX = import.meta.env.VITE_RENTALS_PREFIX || '/rentals';

export const rentalsApi = {
  create(bicycleId, userId) {
    return client.post(PREFIX, { bicycle_id: bicycleId, user_id: userId });
  },

  finalize(rentalId) {
    return client.put(`${PREFIX}/${rentalId}/finalize`);
  },

  getActive() {
    return client.get(`${PREFIX}/active`);
  },
};
