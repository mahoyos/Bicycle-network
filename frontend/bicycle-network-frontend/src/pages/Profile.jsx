import { useAuth } from '../contexts/AuthContext';
import { User, Mail, Shield, Clock, CheckCircle, XCircle } from 'lucide-react';

export default function Profile() {
  const { user } = useAuth();

  if (!user) return null;

  return (
    <div className="page profile-page">
      <div className="page-header">
        <h2>Profile</h2>
        <p className="text-muted">Your account information</p>
      </div>

      <div className="profile-card">
        <div className="profile-avatar">
          <User size={64} />
        </div>

        <div className="profile-fields">
          <div className="detail-field">
            <Mail size={18} />
            <div>
              <span className="field-label">Email</span>
              <span className="field-value">{user.email}</span>
            </div>
          </div>

          <div className="detail-field">
            <Shield size={18} />
            <div>
              <span className="field-label">Role</span>
              <span className="field-value capitalize">{user.role}</span>
            </div>
          </div>

          <div className="detail-field">
            {user.is_active ? (
              <CheckCircle size={18} className="text-green" />
            ) : (
              <XCircle size={18} className="text-red" />
            )}
            <div>
              <span className="field-label">Status</span>
              <span className="field-value">
                {user.is_active ? 'Active' : 'Inactive'}
              </span>
            </div>
          </div>

          <div className="detail-field">
            <Clock size={18} />
            <div>
              <span className="field-label">Member Since</span>
              <span className="field-value">
                {new Date(user.created_at).toLocaleDateString('en-US', {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric',
                })}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
