import { KeyboardEvent, ReactNode, useContext, useMemo, useState } from 'react';
import { NavLink, useNavigate } from 'react-router';
import stashsphereLogo from '../assets/stashsphere-logo-256.png';
import { Icon } from './shared';
import { SearchContext } from '../context/search';
import { Profile } from '../api/resources';
import { UserNameAndProfile } from './shared/user';
import { CartContext } from '../context/cart';
import { formatHumanUnit } from '../lib/format';

type HeaderProps = {
  profile: Profile | null;
  hasUnacknowledgedNotifications: boolean;
};

const NavItem = ({
  to,
  children,
  onCLick,
}: {
  to: string;
  children: ReactNode;
  onCLick?: () => void;
}) => {
  return (
    <NavLink
      to={to}
      onClick={onCLick}
      className={({ isActive }) => {
        let name = 'text-onprimary py-4 px-2 font-semibold border-b-4 transition duration-300';

        if (isActive) {
          name += ' border-secondary';
        } else {
          name += ' border-transparent hover:border-secondary-hover';
        }

        return name;
      }}
    >
      {children}
    </NavLink>
  );
};

const NameAndLogo = () => {
  return (
    <div className="flex">
      <div className="flex items-center">
        <img src={stashsphereLogo} alt="StashSphere" className="h-12" />
      </div>
      <span className="py-4 px-2 text-onprimary font-goldman font-bold text-xl">stashsphere</span>
    </div>
  );
};

const NotificationItem = ({ hasUnacknowledged }: { hasUnacknowledged: boolean }) => {
  if (hasUnacknowledged) {
    return <Icon className="text-highlight" icon={'mdi--notifications'} />;
  } else {
    return <Icon className="" icon={'mdi--notifications-none'} />;
  }
};

const CartItem = () => {
  const { cart } = useContext(CartContext);

  const formatted = useMemo(() => {
    return formatHumanUnit(cart?.entries.length || 0);
  }, [cart]);

  return (
    <NavItem to="/cart">
      <div className="flex">
        <Icon className="" icon={'mdi--cart-heart'} />
        <div className="w-5 text-sm mt-2">{formatted}</div>
      </div>
    </NavItem>
  );
};

const HeaderLoggedIn = ({
  profile,
  hasUnacknowledgedNotifications,
}: {
  profile: Profile;
  hasUnacknowledgedNotifications: boolean;
}) => {
  const [query, setQuery] = useState('');
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const { setSearchTerm } = useContext(SearchContext);
  const navigate = useNavigate();

  const handleSearch = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      navigate(`/search`);
    }
    setSearchTerm(query);
  };

  const toggleMobileMenu = () => {
    setMobileMenuOpen((prev) => !prev);
  };

  return (
    <nav className="bg-primary shadow-lg mb-4">
      <div className="max-w-6xl mx-auto px-4">
        <div className="flex justify-between items-center">
          <a href="/">
            <NameAndLogo />
          </a>
          <div className="flex grow mx-5 md:mx-5">
            <input
              className="w-full my-2 px-4 py-2 text-display rounded border focus:outline-hidden border-secondary"
              placeholder="Search"
              onChange={(e) => setQuery(e.target.value)}
              value={query}
              onKeyDown={(e) => handleSearch(e)}
            />
          </div>
          <button
            className="md:hidden text-onprimary ml-2"
            onClick={toggleMobileMenu}
            aria-label="Menu"
          >
            <Icon icon="mdi--menu" className="text-2xl" />
          </button>
          <div className="hidden md:flex items-center space-x-1">
            <NavItem to="/things">Things</NavItem>
            <NavItem to="/lists">Lists</NavItem>
            <NavItem to="/images">Images</NavItem>
            <NavItem to="/friends">Friends</NavItem>
          </div>
          <div className="hidden md:block border-2 mx-2 border-highlight self-stretch"></div>
          <div className="hidden md:flex items-center space-x-1">
            <NavItem to="/user/profile" onCLick={toggleMobileMenu}>
              <UserNameAndProfile
                profile={profile}
                imageBorderColor="border-highlight"
                textColor="text-highlight"
              />
            </NavItem>
            <NavItem to="/notifications">
              <NotificationItem hasUnacknowledged={hasUnacknowledgedNotifications} />
            </NavItem>
            <CartItem />
            <NavItem to="/user/logout">Logout</NavItem>
          </div>
        </div>

        {mobileMenuOpen && (
          <div className="md:hidden fixed top-[header-height] left-0 right-0 pt-2 pb-4 bg-primary shadow-lg z-50">
            <div className="flex flex-col space-y-2">
              <NavItem to="/things" onCLick={toggleMobileMenu}>
                Things
              </NavItem>
              <NavItem to="/lists" onCLick={toggleMobileMenu}>
                Lists
              </NavItem>
              <NavItem to="/images" onCLick={toggleMobileMenu}>
                Images
              </NavItem>
              <NavItem to="/friends" onCLick={toggleMobileMenu}>
                Friends
              </NavItem>
              <div className="border-t-2 my-2 border-highlight"></div>
              <NavItem to="/user/profile" onCLick={toggleMobileMenu}>
                <Icon icon="mdi--user" className="mr-2" />
                {profile.name}
              </NavItem>
              <NavItem to="/notifications" onCLick={toggleMobileMenu}>
                Notifications
              </NavItem>
              <NavItem to="/cart" onCLick={toggleMobileMenu}>
                Cart
              </NavItem>
              <NavItem to="/user/logout" onCLick={toggleMobileMenu}>
                Logout
              </NavItem>
            </div>
          </div>
        )}
      </div>
    </nav>
  );
};

const HeaderLoggedOut = () => {
  return (
    <nav className="bg-primary shadow-lg mb-4">
      <div className="max-w-6xl mx-auto px-4">
        <div className="flex justify-between">
          <a href="/">
            <NameAndLogo />
          </a>
          <div className="flex items-center space-x-1">
            <NavItem to="/user/login">Login</NavItem>
          </div>
        </div>
      </div>
    </nav>
  );
};

export const Header = ({ profile, hasUnacknowledgedNotifications }: HeaderProps) => {
  return profile !== null ? (
    <HeaderLoggedIn
      profile={profile}
      hasUnacknowledgedNotifications={hasUnacknowledgedNotifications}
    />
  ) : (
    <HeaderLoggedOut />
  );
};
