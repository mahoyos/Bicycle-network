import createClient from './client';
const client = createClient(import.meta.env.VITE_API_GATEWAY_URL_AUTH || 'http://localhost:8080');

const PREFIX = import.meta.env.VITE_AUTH_PREFIX || '/auth';

export const authApi = {
  register(email, password) {
    return client.post(`${PREFIX}/register`, { email, password });
  },

  login(email, password) {
    return client.post(`${PREFIX}/login`, { email, password });
  },

  refresh(refreshToken) {
    return client.post(`${PREFIX}/refresh`, { refresh_token: refreshToken });
  },

  logout(refreshToken) {
    return client.post(`${PREFIX}/logout`, { refresh_token: refreshToken });
  },

  getMe() {
    return client.get(`${PREFIX}/me`);
  },

  passwordRecovery(email) {
    return client.post(`${PREFIX}/password-recovery`, { email });
  },

  passwordReset(token, newPassword) {
    return client.post(`${PREFIX}/password-reset`, { token, new_password: newPassword });
  },

  getGoogleLoginUrl() {
    const apiUrl = import.meta.env.VITE_API_GATEWAY_URL || 'http://localhost:8080';
    return `${apiUrl}${PREFIX}/google/login`;
  },

  openGoogleLogin() {
    return new Promise((resolve, reject) => {
      const url = this.getGoogleLoginUrl();
      const popup = window.open(url, 'google-login', 'width=500,height=600,left=200,top=100');
      if (!popup) {
        reject(new Error('Popup blocked. Please allow popups for this site.'));
        return;
      }

      const interval = setInterval(() => {
        try {
          if (popup.closed) {
            clearInterval(interval);
            reject(new Error('Login cancelled'));
            return;
          }
          // Once popup is back on our origin after Google redirect
          if (popup.location.origin === window.location.origin) {
            const text = popup.document.body.innerText;
            if (text) {
              const data = JSON.parse(text);
              popup.close();
              clearInterval(interval);
              resolve(data);
            }
          }
        } catch {
          // Cross-origin while on Google's domain — ignore
        }
      }, 500);
    });
  },
};
