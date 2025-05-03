import { KeyboardEvent, ReactNode, useContext, useState } from 'react';
import { NavLink, useNavigate } from 'react-router';
import stashsphereLogo from '../assets/stashsphere.svg';
import { Icon } from './shared';
import { SearchContext } from '../context/search';

type HeaderProps = {
  userName: string | null;
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
        <img src={stashsphereLogo} alt="StashSphere" className="h-10" />
      </div>
      <span className="py-4 px-2 text-onprimary font-semibold">stashsphere</span>
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

const HeaderLoggedIn = ({
  userName,
  hasUnacknowledgedNotifications,
}: {
  userName: string;
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
              <span className="text-highlight">
                <Icon icon="mdi--user" className="mr-2" />
                {userName}
              </span>
            </NavItem>
            <NavItem to="/notifications">
              <NotificationItem hasUnacknowledged={hasUnacknowledgedNotifications} />
            </NavItem>
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
                {userName}
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

export const Header = ({ userName, hasUnacknowledgedNotifications }: HeaderProps) => {
  return userName !== null ? (
    <HeaderLoggedIn
      userName={userName}
      hasUnacknowledgedNotifications={hasUnacknowledgedNotifications}
    />
  ) : (
    <HeaderLoggedOut />
  );
};
