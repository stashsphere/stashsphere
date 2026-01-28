import { ReactNode } from 'react';
import { NavLink, Outlet } from 'react-router';

const UserNavItem = ({ to, children }: { to: string; children: ReactNode }) => {
  return (
    <NavLink
      to={to}
      className={({ isActive }) => {
        let name = 'text-onprimary py-2 px-4 font-semibold border-b-4 transition duration-300';

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

export const UserLayout = () => {
  return (
    <div>
      <nav className="bg-primary shadow-md mb-4">
        <div className="max-w-6xl mx-auto px-4">
          <div className="flex items-center space-x-1">
            <UserNavItem to="/user/profile">Profile</UserNavItem>
            <UserNavItem to="/user/account">Account</UserNavItem>
          </div>
        </div>
      </nav>
      <Outlet />
    </div>
  );
};
