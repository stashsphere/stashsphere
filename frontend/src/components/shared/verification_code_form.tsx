import { FormEvent } from 'react';
import { PrimaryButton } from './button';

type VerificationCodeFormProps = {
  code: string;
  onCodeChange: (code: string) => void;
  onSubmit: () => void;
  submitting: boolean;
  error?: string;
};

export const VerificationCodeForm = ({
  code,
  onCodeChange,
  onSubmit,
  submitting,
  error,
}: VerificationCodeFormProps) => {
  const isCodeValid = code.length === 8 && /^[0-9]{8}$/.test(code);

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    onSubmit();
  };

  return (
    <form onSubmit={handleSubmit}>
      <div className="mb-4">
        <label htmlFor="code" className="block text-primary text-sm font-medium">
          Verification Code
        </label>
        <input
          type="text"
          id="code"
          name="code"
          value={code}
          onChange={(e) => onCodeChange(e.target.value)}
          className="mt-1 p-2 w-full border border-secondary rounded-sm text-display"
          placeholder="12345678"
          maxLength={8}
          autoFocus
        />
        <p className="text-xs text-secondary mt-1">Enter the 8-digit code from your e-mail</p>
      </div>

      <PrimaryButton type="submit" disabled={!isCodeValid || submitting}>
        {submitting ? 'Verifying...' : 'Verify E-Mail'}
      </PrimaryButton>

      {error && <p className="text-warning mt-2">{error}</p>}
    </form>
  );
};
