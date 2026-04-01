import createClient from './client';
const client = createClient(import.meta.env.VITE_API_GATEWAY_URL_BIKES || 'http://localhost:8080');

const PREFIX = import.meta.env.VITE_BIKES_PREFIX || '/bikes';

export const bicyclesApi = {
  list(params = {}) {
    const query = new URLSearchParams();
    if (params.type) query.set('type', params.type);
    if (params.page) query.set('page', String(params.page));
    if (params.limit) query.set('limit', String(params.limit));
    const qs = query.toString();
    return client.get(`${PREFIX}${qs ? `?${qs}` : ''}`);
  },

  getById(bikeId) {
    return client.get(`${PREFIX}/${bikeId}`);
  },

  create(data) {
    return client.post(PREFIX, data);
  },

  update(bikeId, data) {
    return client.put(`${PREFIX}/${bikeId}`, data);
  },

  delete(bikeId) {
    return client.delete(`${PREFIX}/${bikeId}`);
  },
};
