import { FormEvent, useContext, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { PrimaryButton } from '../../components/shared';
import { updatePassword } from '../../api/profile';

export const Account = () => {
  const axiosInstance = useContext(AxiosContext);
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState<string | undefined>(undefined);
  const [success, setSuccess] = useState(false);

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(undefined);
    setSuccess(false);

    if (axiosInstance === null) {
      return;
    }

    if (newPassword !== confirmPassword) {
      setError('New passwords do not match.');
      return;
    }

    if (newPassword.length < 1) {
      setError('New password cannot be empty.');
      return;
    }

    try {
      await updatePassword(axiosInstance, { oldPassword, newPassword });
      setSuccess(true);
      setOldPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch {
      setError('Failed to update password. Please check your current password.');
    }
  };

  return (
    <div>
      <h1 className="text-2xl text-secondary">Account</h1>
      <div className="mt-4">
        <div className="max-w-md">
          <h2 className="text-primary text-xl font-semibold mb-4">Change Password</h2>
          <form onSubmit={onSubmit}>
            <div className="mb-4">
              <label htmlFor="oldPassword" className="block text-primary text-sm font-medium">
                Current Password
              </label>
              <input
                type="password"
                id="oldPassword"
                name="oldPassword"
                value={oldPassword}
                onChange={(e) => setOldPassword(e.target.value)}
                className="mt-1 p-2 w-full border border-secondary rounded-sm text-display"
              />
            </div>
            <div className="mb-4">
              <label htmlFor="newPassword" className="block text-primary text-sm font-medium">
                New Password
              </label>
              <input
                type="password"
                id="newPassword"
                name="newPassword"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className="mt-1 p-2 w-full border border-secondary rounded-sm text-display"
              />
            </div>
            <div className="mb-4">
              <label htmlFor="confirmPassword" className="block text-primary text-sm font-medium">
                Confirm New Password
              </label>
              <input
                type="password"
                id="confirmPassword"
                name="confirmPassword"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="mt-1 p-2 w-full border border-secondary rounded-sm text-display"
              />
            </div>
            <PrimaryButton type="submit">Update Password</PrimaryButton>
            {error && <p className="text-warning mt-2">{error}</p>}
            {success && <p className="text-success mt-2">Password updated successfully.</p>}
          </form>
        </div>
      </div>
    </div>
  );
};
