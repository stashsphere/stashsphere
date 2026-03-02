import { FormEvent, useContext, useEffect, useState } from 'react';
import { AxiosContext } from '../context/axios';
import { useNavigate } from 'react-router';
import { AuthContext } from '../context/auth';
import { PrimaryButton, SecondaryButton } from '../components/shared';
import { login } from '../api/auth';
import { getInstanceInfo } from '../api/info';
import { InstanceInfo } from '../api/resources';
import { ConfigContext } from '../context/config';

export const Login = () => {
  const navigate = useNavigate();
  const axiosInstance = useContext(AxiosContext);
  const authContext = useContext(AuthContext);
  const config = useContext(ConfigContext);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [instanceInfo, setInstanceInfo] = useState<InstanceInfo | null>(null);

  const [error, setError] = useState<string | undefined>(undefined);

  useEffect(() => {
    if (authContext.loggedIn) {
      navigate('/');
    }
  }, [navigate, authContext.loggedIn]);

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    getInstanceInfo(axiosInstance)
      .then(setInstanceInfo)
      .catch((e) => console.error(e));
  }, [axiosInstance]);

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (axiosInstance === null) {
      return;
    }
    try {
      await login(axiosInstance, email, password);
      setError(undefined);
    } catch {
      setError('Wrong username or password.');
    }
  };

  const oidcProviders = instanceInfo?.oidcProviders ?? [];

  return (
    <div className="flex items-center justify-center">
      <div className="flex-none bg-white p-8 rounded-sm shadow-md w-96">
        <h2 className="text-primary text-2xl font-semibold mb-4">Login</h2>
        <form onSubmit={onSubmit}>
          <div className="mb-4">
            <label htmlFor="email" className="block text-primary text-sm font-medium">
              E-Mail
            </label>
            <input
              type="text"
              id="email"
              name="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="mt-1 p-2 w-full border border-secondary rounded-sm text-display"
            />
          </div>
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
            />
          </div>
          <PrimaryButton type="submit">Login</PrimaryButton>
          {error && <p className="text-danger-400">{error}</p>}
        </form>
        {oidcProviders.length > 0 && (
          <div className="mt-4">
            <div className="flex items-center mb-3">
              <hr className="flex-grow border-secondary" />
              <span className="px-2 text-sm text-secondary">or</span>
              <hr className="flex-grow border-secondary" />
            </div>
            {oidcProviders.map((provider) => {
              const redirectTo = window.location.origin + '/auth/callback';
              const authorizeUrl =
                config.apiHost +
                '/api/auth/oidc/' +
                provider.name +
                '/authorize' +
                '?redirect_to=' +
                encodeURIComponent(redirectTo);
              return (
                <a key={provider.name} href={authorizeUrl} className="block mb-2">
                  <SecondaryButton type="button" className="w-full">
                    Login with {provider.displayName}
                  </SecondaryButton>
                </a>
              );
            })}
          </div>
        )}
        <a href="/user/register" className="underline text-secondary mt-2 inline-block">
          Register an account
        </a>
      </div>
    </div>
  );
};
