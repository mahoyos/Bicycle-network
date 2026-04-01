import client from './client';

const PREFIX = import.meta.env.VITE_EVENTS_PREFIX || '/events';

export const eventsApi = {
  list(params = {}) {
    const query = new URLSearchParams();
    if (params.skip != null) query.set('skip', String(params.skip));
    if (params.limit != null) query.set('limit', String(params.limit));
    if (params.name) query.set('name', params.name);
    if (params.type) query.set('type', params.type);
    if (params.date) query.set('date', params.date);
    const qs = query.toString();
    return client.get(`${PREFIX}/${qs ? `?${qs}` : ''}`);
  },

  getById(eventId) {
    return client.get(`${PREFIX}/${eventId}`);
  },

  create(data) {
    return client.post(`${PREFIX}/`, data);
  },

  update(eventId, data) {
    return client.put(`${PREFIX}/${eventId}`, data);
  },

  delete(eventId) {
    return client.delete(`${PREFIX}/${eventId}`);
  },

  register(eventId, userId) {
    return client.post(`${PREFIX}/${eventId}/registrations`, { user_id: userId });
  },

  unregister(eventId, userId) {
    return client.delete(`${PREFIX}/${eventId}/registrations/${userId}`);
  },

  getUserRegistrations(userId) {
    return client.get(`/users/${userId}/registrations`);
  },
};
