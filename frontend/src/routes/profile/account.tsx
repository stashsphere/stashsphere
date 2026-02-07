import { FormEvent, useContext, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { AuthContext } from '../../context/auth';
import {
  PrimaryButton,
  DangerButton,
  SecondaryButton,
  WarningButton,
  Modal,
  PasswordInput,
  usePasswordValidation,
  VerificationCodeForm,
} from '../../components/shared';
import {
  updatePassword,
  scheduleDeletion,
  cancelDeletion,
  requestEmailVerification,
  verifyEmail,
} from '../../api/profile';

export const Account = () => {
  const axiosInstance = useContext(AxiosContext);
  const { profile, invalidateProfile } = useContext(AuthContext);
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState<string | undefined>(undefined);
  const [success, setSuccess] = useState(false);

  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [deletePassword, setDeletePassword] = useState('');
  const [confirmText, setConfirmText] = useState('');
  const [deleteError, setDeleteError] = useState<string | undefined>(undefined);
  const [deleteLoading, setDeleteLoading] = useState(false);

  const [requesting, setRequesting] = useState(false);
  const [requestError, setRequestError] = useState<string | undefined>(undefined);
  const [requestSuccess, setRequestSuccess] = useState(false);
  const [code, setCode] = useState('');
  const [verifying, setVerifying] = useState(false);
  const [verifyError, setVerifyError] = useState<string | undefined>(undefined);

  const { isValid: isPasswordValid } = usePasswordValidation(newPassword, confirmPassword, 8);

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(undefined);
    setSuccess(false);

    if (axiosInstance === null) {
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

  const handleScheduleDeletion = async () => {
    if (axiosInstance === null || confirmText !== 'YES' || !deletePassword) {
      return;
    }

    setDeleteLoading(true);
    setDeleteError(undefined);

    try {
      await scheduleDeletion(axiosInstance, deletePassword);
      invalidateProfile();
      setDeleteModalOpen(false);
      setConfirmText('');
      setDeletePassword('');
    } catch {
      setDeleteError('Failed to schedule account deletion. Please check your password.');
    } finally {
      setDeleteLoading(false);
    }
  };

  const handleCancelDeletion = async () => {
    if (axiosInstance === null) {
      return;
    }

    setDeleteLoading(true);
    setDeleteError(undefined);

    try {
      await cancelDeletion(axiosInstance);
      invalidateProfile();
    } catch {
      setDeleteError('Failed to cancel account deletion.');
    } finally {
      setDeleteLoading(false);
    }
  };

  const handleRequestVerification = async () => {
    if (axiosInstance === null) return;

    setRequesting(true);
    setRequestError(undefined);
    setRequestSuccess(false);

    try {
      await requestEmailVerification(axiosInstance);
      setRequestSuccess(true);
    } catch {
      setRequestError('Failed to send verification e-mail. Please try again.');
    } finally {
      setRequesting(false);
    }
  };

  const handleVerify = async () => {
    if (axiosInstance === null) return;
    if (code.length !== 8) return;

    setVerifying(true);
    setVerifyError(undefined);

    try {
      await verifyEmail(axiosInstance, code);
      invalidateProfile();
    } catch {
      setVerifyError('Invalid or expired verification code. Please request a new one.');
    } finally {
      setVerifying(false);
    }
  };

  const formatDate = (date: Date) => {
    return new Date(date).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
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
            <PasswordInput
              password={newPassword}
              confirmPassword={confirmPassword}
              onPasswordChange={setNewPassword}
              onConfirmPasswordChange={setConfirmPassword}
              passwordLabel="New Password"
              confirmLabel="Confirm New Password"
              minLength={8}
            />
            <PrimaryButton type="submit" disabled={!isPasswordValid}>
              Update Password
            </PrimaryButton>
            {error && <p className="text-warning mt-2">{error}</p>}
            {success && <p className="text-success mt-2">Password updated successfully.</p>}
          </form>
        </div>

        <div className="max-w-md mt-8 pt-8 border-t border-gray-200">
          <h2 className="text-primary text-xl font-semibold mb-4">E-Mail Verification</h2>
          {profile?.emailVerified === true ? (
            <div className="bg-success-900 border border-success-700 rounded-sm p-4">
              <p className="text-success-200">Your E-Mail address has been verified.</p>
            </div>
          ) : profile?.emailVerified === false ? (
            <div>
              <div className="bg-warning-900 border border-warning-700 rounded-sm p-4 mb-4">
                <p className="text-warning-200">
                  Please verify your E-Mail address. This is required if you ever need to reset your
                  password.
                </p>
              </div>
              {requestSuccess ? (
                <div>
                  <p className="text-display mb-4">
                    Verification E-Mail sent! Check your inbox and enter the code below.
                  </p>
                  <VerificationCodeForm
                    code={code}
                    onCodeChange={setCode}
                    onSubmit={handleVerify}
                    submitting={verifying}
                    error={verifyError}
                  />
                </div>
              ) : (
                <div>
                  <WarningButton onClick={handleRequestVerification} disabled={requesting}>
                    {requesting ? 'Sending...' : 'Send Verification E-Mail'}
                  </WarningButton>
                  {requestError && <p className="text-warning mt-2">{requestError}</p>}
                </div>
              )}
            </div>
          ) : null}
        </div>

        <div className="max-w-md mt-8 pt-8 border-t border-gray-200">
          <h2 className="text-danger text-xl font-semibold mb-4">Delete Account</h2>
          {profile?.purgeAt ? (
            <div>
              <div className="bg-yellow-50 border border-yellow-200 rounded-sm p-4 mb-4">
                <p className="text-yellow-800">
                  Your account is scheduled for deletion on{' '}
                  <strong>{formatDate(profile.purgeAt)}</strong>.
                </p>
                <p className="text-yellow-700 text-sm mt-2">
                  All your data will be permanently deleted at this time.
                </p>
              </div>
              <SecondaryButton onClick={handleCancelDeletion} disabled={deleteLoading}>
                {deleteLoading ? 'Canceling...' : 'Cancel Deletion'}
              </SecondaryButton>
              {deleteError && <p className="text-warning mt-2">{deleteError}</p>}
            </div>
          ) : (
            <div>
              <div className="bg-danger-900 border border-danger-400 rounded-sm p-4 mb-4">
                <p className="text-danger-700">
                  Deletion is scheduled with a grace period during which you can cancel. After that,
                  all your data will be permanently removed.
                </p>
              </div>
              <DangerButton onClick={() => setDeleteModalOpen(true)}>Delete Account</DangerButton>
            </div>
          )}
        </div>
      </div>

      <Modal
        isOpen={deleteModalOpen}
        onClose={() => {
          setDeleteModalOpen(false);
          setDeletePassword('');
          setConfirmText('');
          setDeleteError(undefined);
        }}
        title="Delete Account"
        size="sm"
      >
        <div className="space-y-4">
          <div className="bg-danger-900 border border-danger-400 rounded-sm p-4">
            <p className="text-danger-500 font-medium">
              Are you sure you want to delete your account?
            </p>
            <p className="text-danger-500 text-sm mt-2">
              Your account will be scheduled for deletion. You can cancel the deletion until the
              scheduled date, but once your account is purged, all your things, lists, images, and
              other data will be permanently removed.
            </p>
          </div>

          <div>
            <label htmlFor="deletePassword" className="block text-primary text-sm font-medium mb-2">
              Enter your password
            </label>
            <input
              type="password"
              id="deletePassword"
              value={deletePassword}
              onChange={(e) => setDeletePassword(e.target.value)}
              className="p-2 w-full border border-secondary rounded-sm text-display"
            />
          </div>

          <div>
            <label htmlFor="confirmDelete" className="block text-primary text-sm font-medium mb-2">
              Type <strong>YES</strong> to confirm
            </label>
            <input
              type="text"
              id="confirmDelete"
              value={confirmText}
              onChange={(e) => setConfirmText(e.target.value)}
              className="p-2 w-full border border-secondary rounded-sm text-display"
              placeholder="YES"
            />
          </div>

          {deleteError && <p className="text-warning">{deleteError}</p>}

          <div className="flex gap-2">
            <DangerButton
              onClick={handleScheduleDeletion}
              disabled={confirmText !== 'YES' || !deletePassword || deleteLoading}
            >
              {deleteLoading ? 'Scheduling...' : 'Delete Account'}
            </DangerButton>
            <SecondaryButton
              onClick={() => {
                setDeleteModalOpen(false);
                setDeletePassword('');
                setConfirmText('');
                setDeleteError(undefined);
              }}
            >
              Cancel
            </SecondaryButton>
          </div>
        </div>
      </Modal>
    </div>
  );
};
