import { useMemo } from 'react';

type PasswordComplexity = 'weak' | 'medium' | 'strong';

interface PasswordInputProps {
  password: string;
  confirmPassword: string;
  onPasswordChange: (password: string) => void;
  onConfirmPasswordChange: (confirmPassword: string) => void;
  passwordLabel?: string;
  confirmLabel?: string;
  minLength?: number;
}

const getPasswordComplexity = (password: string, minLength: number): PasswordComplexity => {
  if (password.length < minLength) {
    return 'weak';
  }

  const hasLowercase = /[a-z]/.test(password);
  const hasUppercase = /[A-Z]/.test(password);
  const hasNumbers = /[0-9]/.test(password);
  const hasSpecial = /[^a-zA-Z0-9]/.test(password);

  const typesCount = [hasLowercase, hasUppercase, hasNumbers, hasSpecial].filter(Boolean).length;

  // Length bonuses: 12+ chars = +1, 16+ chars = +2
  let lengthBonus = 0;
  if (password.length >= 16) {
    lengthBonus = 2;
  } else if (password.length >= 12) {
    lengthBonus = 1;
  }

  const score = typesCount + lengthBonus;

  if (score >= 4) {
    return 'strong';
  }
  if (score >= 2) {
    return 'medium';
  }
  return 'weak';
};

const complexityColors: Record<PasswordComplexity, string> = {
  weak: 'bg-danger',
  medium: 'bg-warning',
  strong: 'bg-success',
};

const complexityLabels: Record<PasswordComplexity, string> = {
  weak: 'Weak',
  medium: 'Medium',
  strong: 'Strong',
};

export const PasswordInput = ({
  password,
  confirmPassword,
  onPasswordChange,
  onConfirmPasswordChange,
  passwordLabel = 'Password',
  confirmLabel = 'Confirm Password',
  minLength = 8,
}: PasswordInputProps) => {
  const complexity = useMemo(
    () => getPasswordComplexity(password, minLength),
    [password, minLength]
  );
  const passwordsMatch = password === confirmPassword;
  const isLongEnough = password.length >= minLength;

  return (
    <>
      <div className="mb-4">
        <label htmlFor="password" className="block text-primary text-sm font-medium">
          {passwordLabel}
        </label>
        <input
          type="password"
          id="password"
          name="password"
          value={password}
          onChange={(e) => onPasswordChange(e.target.value)}
          className="mt-1 p-2 w-full border border-secondary rounded-sm text-display"
        />
        {password.length > 0 && (
          <div className="mt-2">
            <div className="flex items-center gap-2">
              <div className="flex-1 h-2 bg-gray-200 rounded-sm overflow-hidden">
                <div
                  className={`h-full transition-all ${complexityColors[complexity]}`}
                  style={{
                    width: complexity === 'weak' ? '33%' : complexity === 'medium' ? '66%' : '100%',
                  }}
                />
              </div>
              <span className="text-xs text-secondary">{complexityLabels[complexity]}</span>
            </div>
            {!isLongEnough && (
              <p className="text-warning text-xs mt-1">
                Password must be at least {minLength} characters
              </p>
            )}
          </div>
        )}
      </div>
      <div className="mb-4">
        <label htmlFor="confirmPassword" className="block text-primary text-sm font-medium">
          {confirmLabel}
        </label>
        <input
          type="password"
          id="confirmPassword"
          name="confirmPassword"
          value={confirmPassword}
          onChange={(e) => onConfirmPasswordChange(e.target.value)}
          className="mt-1 p-2 w-full border border-secondary rounded-sm text-display"
        />
        {confirmPassword.length > 0 && !passwordsMatch && (
          <p className="text-warning text-xs mt-1">Passwords do not match</p>
        )}
      </div>
    </>
  );
};
