
import { KeyboardEvent, ReactNode, useContext, useState } from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import stashsphereLogo from "../assets/stashsphere.svg";
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
        let name = "text-onprimary py-4 px-2 font-semibold hover:border-b-4 hover:border-secondary-hover transition duration-300"
        name += isActive ? " border-b-4 border-secondary" : "";
        return name;
      }}

    >
      {children}
    </NavLink>
  );
};

const NameAndLogo = () => {
  return <div className="flex">
    <div className="flex items-center">
      <img src={stashsphereLogo} alt="StashSphere" className="h-10"/> 
    </div>
    <span className="py-4 px-2 text-onprimary font-semibold">
      stashsphere
    </span>
  </div>
}

const HeaderLoggedIn = ({ userName }: { userName: string }) => {
  const [query, setQuery] = useState("");
  const { setSearchTerm } = useContext(SearchContext);
  const navigate = useNavigate();

  const handleSearch = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      navigate(`/search`);
    }
    setSearchTerm(query);
  }

  return (
    <nav className="bg-primary shadow-lg mb-4">
      <div className="max-w-6xl mx-auto px-4">
        <div className="flex justify-between">
          <a href="/"><NameAndLogo /></a>
          <div className="flex flex-grow mx-5">
            <input className="w-full my-2 px-4 py-2 text-display rounded border
            focus:outline-none
            bg-neutral-secondary
            border-secondary"
              placeholder="Search"
              onChange={(e) => setQuery(e.target.value)}
              value={query}
              onKeyDown={(e) => handleSearch(e)}
            />
          </div>
          <div className="md:flex items-center space-x-1">
            <NavItem to="/things">Things</NavItem>
            <NavItem to="/lists">Lists</NavItem>
            <NavItem to="/images">Images</NavItem>
            <NavItem to="/friends">Friends</NavItem>
          </div>
          <div className='border-2 mx-2 border-highlight'></div>
          <div className="md:flex items-center space-x-1">
            <div className="py-4 px-2 text-highlight font-semibold">
              <Icon icon="mdi--user" />
              {userName}
            </div>
            <NavItem to="/user/profile">Profile</NavItem>
            <NavItem to="/user/logout">
              Logout
            </NavItem>
          </div>
        </div>
      </div>
    </nav>
  );
};

const HeaderLoggedOut = () => {
  return (
    <nav className="bg-primary shadow-lg mb-4">
      <div className="max-w-6xl mx-auto px-4">
        <div className="flex justify-between">
          <a href="/"><NameAndLogo /></a>
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
