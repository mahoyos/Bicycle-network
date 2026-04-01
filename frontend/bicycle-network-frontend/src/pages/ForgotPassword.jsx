import { useState } from 'react';
import { Link } from 'react-router-dom';
import { authApi } from '../api/auth';
import { Bike, Mail, ArrowLeft } from 'lucide-react';
import toast from 'react-hot-toast';

export default function ForgotPassword() {
  const [email, setEmail] = useState('');
  const [submitted, setSubmitted] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await authApi.passwordRecovery(email);
      setSubmitted(true);
    } catch {
      toast.error('Something went wrong. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="auth-page">
      <div className="auth-card">
        <div className="auth-logo">
          <Bike size={48} />
          <h1>Reset Password</h1>
          <p>
            {submitted
              ? 'Check your email for a reset link'
              : 'Enter your email to receive a recovery link'}
          </p>
        </div>

        {!submitted ? (
          <form onSubmit={handleSubmit} className="auth-form">
            <div className="form-group">
              <label htmlFor="email">Email</label>
              <div className="input-wrapper">
                <Mail size={18} />
                <input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="you@example.com"
                  required
                />
              </div>
            </div>

            <button type="submit" className="btn btn-primary btn-full" disabled={submitting}>
              {submitting ? <span className="spinner-sm" /> : 'Send Recovery Link'}
            </button>
          </form>
        ) : (
          <div className="success-message">
            <p>If an account exists with that email, you will receive a password reset link.</p>
          </div>
        )}

        <Link to="/login" className="back-link">
          <ArrowLeft size={16} />
          Back to login
        </Link>
      </div>
    </div>
  );
}
