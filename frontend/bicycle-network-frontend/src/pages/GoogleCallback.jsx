import { useEffect, useRef } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

export default function GoogleCallback() {
  const [searchParams] = useSearchParams();
  const { loginWithTokens } = useAuth();
  const navigate = useNavigate();
  const processed = useRef(false);

  const accessToken = searchParams.get('access_token');
  const refreshToken = searchParams.get('refresh_token');

  useEffect(() => {
    if (processed.current) return;

    if (!accessToken || !refreshToken) {
      navigate('/login', { replace: true });
      return;
    }

    processed.current = true;

    if (!window.opener) {
      loginWithTokens({ access_token: accessToken, refresh_token: refreshToken })
        .then(() => navigate('/dashboard', { replace: true }))
        .catch(() => navigate('/login', { replace: true }));
    }
  }, [accessToken, refreshToken, loginWithTokens, navigate]);

  if (accessToken && refreshToken) {
    return <>{JSON.stringify({ access_token: accessToken, refresh_token: refreshToken })}</>;
  }

  return (
    <div className="loading-screen">
      <div className="spinner" />
    </div>
  );
}
