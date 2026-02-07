import { useContext, useEffect, useState } from 'react';
import { useNavigate } from 'react-router';
import { AxiosContext } from '../../context/axios';
import { AuthContext } from '../../context/auth';
import { VerificationCodeForm } from '../../components/shared';
import { verifyEmail } from '../../api/profile';

export const VerifyEmail = () => {
  const navigate = useNavigate();
  const axiosInstance = useContext(AxiosContext);
  const { invalidateProfile } = useContext(AuthContext);

  const [code, setCode] = useState('');
  const [error, setError] = useState<string | undefined>(undefined);
  const [success, setSuccess] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  // Extract code from URL fragment on mount
  useEffect(() => {
    const hash = window.location.hash;
    if (hash.startsWith('#')) {
      const fragmentCode = hash.substring(1);
      if (fragmentCode.length === 8) {
        setCode(fragmentCode);
      }
    }
  }, []);

  // Auto-submit when code is populated from fragment
  useEffect(() => {
    if (code.length === 8 && !submitting && !success && window.location.hash.startsWith('#')) {
      handleSubmit();
    }
  }, [code]);

  const handleSubmit = async () => {
    if (axiosInstance === null) return;
    if (code.length !== 8) {
      setError('Verification code must be 8 characters');
      return;
    }

    setSubmitting(true);
    setError(undefined);
    setSuccess(false);

    try {
      await verifyEmail(axiosInstance, code);
      setSuccess(true);
      invalidateProfile();

      setTimeout(() => {
        navigate('/user/profile');
      }, 2000);
    } catch {
      setError('Invalid or expired verification code. Please request a new one.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div>
      <h1 className="text-2xl text-secondary">Verify E-Mail</h1>
      <div className="mt-4">
        <div className="max-w-md">
          {success ? (
            <p className="text-success mt-2">
              E-Mail verified successfully! Redirecting to your profile...
            </p>
          ) : (
            <VerificationCodeForm
              code={code}
              onCodeChange={setCode}
              onSubmit={handleSubmit}
              submitting={submitting}
              error={error}
            />
          )}
        </div>
      </div>
    </div>
  );
};
