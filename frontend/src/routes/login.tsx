import { FormEvent, useContext, useEffect, useState } from 'react';
import { AxiosContext } from '../context/axios';
import { useNavigate } from 'react-router';
import { AuthContext } from '../context/auth';
import { PrimaryButton } from '../components/button';

export const Login = () => {
  const navigate = useNavigate();
  const axiosInstance = useContext(AxiosContext);
  const authContext = useContext(AuthContext);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const [error, setError] = useState<string | undefined>(undefined);

  useEffect(() => {
    if (authContext.loggedIn) {
      navigate('/');
    }
  }, [navigate, authContext.loggedIn]);

  const login = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (axiosInstance === null) {
      return;
    }
    try {
      await axiosInstance.post(
        '/user/login',
        {
          email: email,
          password: password,
        },
        {
          headers: {
            'Content-Type': 'application/json',
          },
        }
      );
      setError(undefined);
    } catch {
      setError('Wrong username or password.');
    }
  };

  return (
    <div className="flex items-center justify-center">
      <div className="flex-none bg-white p-8 rounded-sm shadow-md w-96">
        <h2 className="text-primary text-2xl font-semibold mb-4">Login</h2>
        <form onSubmit={login}>
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
          {error && <p className="text-warning-primary">{error}</p>}
        </form>
        <a href="/user/register" className="underline text-secondary">
          Register an account
        </a>
      </div>
    </div>
  );
};
