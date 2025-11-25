import { FormEvent, useContext, useState } from 'react';
import { PrimaryButton } from '../components/shared';
import { AxiosContext } from '../context/axios';
import { useNavigate } from 'react-router';

export const Register = () => {
  const axiosInstance = useContext(AxiosContext);
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [passwordConfirm, setPasswordConfirm] = useState('');
  const [inviteCode, setInviteCode] = useState('');
  const navigate = useNavigate();

  const [error, setError] = useState<string | undefined>(undefined);
  const submitable = password === passwordConfirm && password.length > 0;

  const register = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (axiosInstance === null) {
      return;
    }
    try {
      await axiosInstance.post(
        '/user/register',
        {
          email,
          password,
          name,
          inviteCode,
        },
        {
          headers: {
            'Content-Type': 'application/json',
          },
        }
      );
      setError(undefined);
      navigate(`/user/login`);
    } catch (error) {
      console.error(error);
      setError('Could not register user.');
    }
  };

  return (
    <div className="flex items-center justify-center">
      <div className="flex-none bg-white p-8 rounded-sm shadow-md w-96">
        <h2 className="text-primary text-2xl font-semibold mb-4">Register an account</h2>
        <form onSubmit={register}>
          <div className="mb-4">
            <label htmlFor="name" className="block text-primary text-sm font-medium">
              Name
            </label>
            <input
              type="text"
              id="name"
              name="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="mt-1 p-2 w-full border border-secondary rounded-sm text-display"
            />
          </div>
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
          <div className="mb-4">
            <label htmlFor="password_confirm" className="block text-primary text-sm font-medium">
              Password (confirm)
            </label>
            <input
              type="password"
              id="password_confirm"
              name="password_confirm"
              value={passwordConfirm}
              onChange={(e) => setPasswordConfirm(e.target.value)}
              className="mt-1 p-2 w-full border border-secondary rounded-sm text-display"
            />
          </div>
          <div className="mb-4">
            <label htmlFor="invite_code" className="block text-primary text-sm font-medium">
              Invite Code
            </label>
            <input
              type="text"
              id="invite_code"
              name="invite_code"
              value={inviteCode}
              onChange={(e) => setInviteCode(e.target.value)}
              className="mt-1 p-2 w-full border border-secondary rounded-sm text-display"
            />
          </div>

          <PrimaryButton type="submit" disabled={!submitable}>
            Register User
          </PrimaryButton>
          {error && <p className="text-warning">{error}</p>}
        </form>
        <a href="/user/login" className="underline text-primary">
          Login instead
        </a>
      </div>
    </div>
  );
};
