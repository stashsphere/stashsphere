import { KeyboardEvent, ReactNode, useContext, useState } from 'react';
import { NavLink, useNavigate } from 'react-router';
import stashsphereLogo from '../assets/stashsphere.svg';
import { Icon } from './icon';
import { SearchContext } from '../context/search';

type HeaderProps = {
  userName: string | null;
};

const NavItem = ({ to, children }: { to: string; children: ReactNode }) => {
  return (
    <NavLink
      to={to}
      className={({ isActive }) => {
        let name =
          'text-onprimary py-4 px-2 font-semibold hover:border-b-4 hover:border-secondary-hover transition duration-300';
        name += isActive ? ' border-b-4 border-secondary' : '';
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

const HeaderLoggedIn = ({ userName }: { userName: string }) => {
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
              className="w-full my-2 px-4 py-2 text-display rounded border focus:outline-hidden bg-neutral-secondary border-secondary"
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
          <div className="hidden md:block border-2 mx-2 border-highlight"></div>
          <div className="hidden md:flex items-center space-x-1">
            <div className="py-4 px-2 text-highlight font-semibold">
              <Icon icon="mdi--user" />
              {userName}
            </div>
            <NavItem to="/user/profile">Profile</NavItem>
            <NavItem to="/user/logout">Logout</NavItem>
          </div>
        </div>

        {mobileMenuOpen && (
          <div className="md:hidden pt-2 pb-4 bg-primary">
            <div className="flex flex-col space-y-2">
              <NavItem to="/things">Things</NavItem>
              <NavItem to="/lists">Lists</NavItem>
              <NavItem to="/images">Images</NavItem>
              <NavItem to="/friends">Friends</NavItem>
              <div className="border-t-2 my-2 border-highlight"></div>
              <div className="py-2 px-2 text-highlight font-semibold flex items-center">
                <Icon icon="mdi--user" className="mr-2" />
                {userName}
              </div>
              <NavItem to="/user/profile">Profile</NavItem>
              <NavItem to="/user/logout">Logout</NavItem>
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

export const Header = ({ userName }: HeaderProps) => {
  return userName !== null ? <HeaderLoggedIn userName={userName} /> : <HeaderLoggedOut />;
};
