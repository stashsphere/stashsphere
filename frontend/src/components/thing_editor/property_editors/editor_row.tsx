import { PrimaryButton } from '../../shared';

export const EditorRow = ({
  children,
  error,
  disabled,
  onSubmit,
  isEditing,
}: {
  children: React.ReactNode;
  error: string | null;
  disabled: boolean;
  onSubmit: () => void;
  isEditing?: boolean;
}) => {
  const buttonText = isEditing ? 'Save' : 'Add';
  return (
    <div>
      <div className="flex flex-wrap items-end gap-3">
        {children}
        <PrimaryButton onClick={onSubmit} disabled={disabled}>
          {buttonText}
        </PrimaryButton>
      </div>
      {error && <p className="mt-2 text-sm text-warning">{error}</p>}
    </div>
  );
};
