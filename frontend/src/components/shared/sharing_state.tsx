import { SharingState } from '../../api/resources';
import { Icon } from './icon';

export const SharingStateComponent = ({ state }: { state: SharingState }) => {
  const [rendered, icon] = (() => {
    switch (state) {
      case 'friends':
        return ['Friends', 'mdi--account-multiple'];
      case 'friends-of-friends':
        return ['Friends of Friends', 'mdi--account-group'];
      case 'private':
        return ['private', 'mdi--lock-outline'];
    }
  })();

  return (
    <div className="rounded-sm bg-secondary-200 text-onprimary mx-1 px-1 flex flex-row">
      <Icon icon={icon} />
      {rendered}
    </div>
  );
};
