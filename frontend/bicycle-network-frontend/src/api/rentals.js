import client from './client';

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
