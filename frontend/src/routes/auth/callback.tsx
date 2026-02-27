import { FormEvent, useContext, useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router';
import { useCookies } from 'react-cookie';
import { AxiosContext } from '../../context/axios';
import { AuthContext } from '../../context/auth';
import { PrimaryButton } from '../../components/shared';

export const OIDCCallback = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const axiosInstance = useContext(AxiosContext);
  const authContext = useContext(AuthContext);
  const [cookies, , removeCookie] = useCookies(['oidc-link-challenge']);

  const action = searchParams.get('action');
  const status = searchParams.get('status');
  const errorParam = searchParams.get('error');
  const errorDescription = searchParams.get('error_description');
  const email = searchParams.get('email');
  const provider = searchParams.get('provider');
  const challengeToken = cookies['oidc-link-challenge'] as string | undefined;

  const [password, setPassword] = useState('');
  const [linkError, setLinkError] = useState<string | undefined>(undefined);
  const [linking, setLinking] = useState(false);

  // Success case: cookies are already set by the backend redirect, just go home
  useEffect(() => {
    if (status === 'success') {
      window.location.href = '/';
    }
  }, [status]);

  // If already logged in and no linking action, redirect home
  useEffect(() => {
    if (authContext.loggedIn && action !== 'link_required') {
      navigate('/');
    }
  }, [authContext.loggedIn, action, navigate]);

  const onLinkSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (axiosInstance === null || !provider || !challengeToken) {
      return;
    }
    setLinking(true);
    try {
      await axiosInstance.post('/auth/oidc/' + provider + '/link', {
        password,
        challengeToken,
      });
      setLinkError(undefined);
      removeCookie('oidc-link-challenge', { path: '/' });
      window.location.href = '/';
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status;
      if (status === 401) {
        setLinkError('Incorrect password. Please try again.');
      } else if (status === 400) {
        setLinkError('Linking session expired. Please try logging in again.');
      } else {
        setLinkError('An error occurred. Please try again.');
      }
    } finally {
      setLinking(false);
    }
  };

  // Error from OIDC provider
  if (errorParam) {
    return (
      <div className="flex items-center justify-center">
        <div className="flex-none bg-white p-8 rounded-sm shadow-md w-96">
          <h2 className="text-primary text-2xl font-semibold mb-4">Login Failed</h2>
          <p className="text-danger-400 mb-4">
            {errorDescription || errorParam || 'An unknown error occurred during login.'}
          </p>
          <a href="/user/login" className="underline text-secondary">
            Back to login
          </a>
        </div>
      </div>
    );
  }

  // Link required: show password form
  if (action === 'link_required' && email && provider && challengeToken) {
    return (
      <div className="flex items-center justify-center">
        <div className="flex-none bg-white p-8 rounded-sm shadow-md w-96">
          <h2 className="text-primary text-2xl font-semibold mb-4">Link Account</h2>
          <p className="text-primary text-sm mb-4">
            An account with the email <strong>{email}</strong> already exists. Enter your password
            to link this provider to your existing account.
          </p>
          <form onSubmit={onLinkSubmit}>
            <div className="mb-4">
              <label htmlFor="password" className="block text-primary text-sm font-medium">
                Password
              </label>
              <input
                type="password"
                id="password"
                name="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="mt-1 p-2 w-full border border-secondary rounded-sm text-display"
                autoFocus
              />
            </div>
            <PrimaryButton type="submit" disabled={linking || password.length === 0}>
              {linking ? 'Linking...' : 'Link Account'}
            </PrimaryButton>
            {linkError && <p className="text-danger-400 mt-2">{linkError}</p>}
          </form>
          <a href="/user/login" className="underline text-secondary mt-4 inline-block">
            Cancel
          </a>
        </div>
      </div>
    );
  }

  // Loading/success state
  return (
    <div className="flex items-center justify-center">
      <div className="flex-none bg-white p-8 rounded-sm shadow-md w-96">
        <p className="text-primary">Completing login...</p>
      </div>
    </div>
  );
};
