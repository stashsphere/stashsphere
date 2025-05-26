import { FormEvent, ReactNode, useEffect, useState } from 'react';
import { Profile } from '../../api/resources';

export type EditableProfile = {
  name: string;
};

type Props = {
  profile: Profile;
  onUpdateProfile: (profile: EditableProfile) => void;
  children?: ReactNode;
};

export const ProfileEditor = ({ children, profile, onUpdateProfile }: Props) => {
  const [name, setName] = useState('');

  useEffect(() => {
    setName(profile.name);
  }, [profile]);

  const onSubmit = (event: FormEvent) => {
    event.preventDefault();
    const data = {
      name,
    };
    onUpdateProfile(data);
  };

  return (
    <form onSubmit={onSubmit}>
      <div className="mb-4">
        <label htmlFor="email" className="block text-primary text-sm font-medium">
          Name
        </label>
        <input
          type="text"
          id="name"
          name="name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="mt-1 p-2 border border-gray-300 rounded-sm text-display"
        />
      </div>
      {children}
    </form>
  );
};
