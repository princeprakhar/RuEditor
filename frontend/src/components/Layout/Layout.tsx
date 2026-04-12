import React from 'react';
import { Outlet } from 'react-router-dom';
import { useAppSelector } from '../../hooks/useRedux';
import Sidebar from '../Sidebar/Sidebar';
import Navbar from '../Navbar/Navbar';

const Layout: React.FC = () => {
  const { user } = useAppSelector(state => state.auth);

  return (
    <div className="flex h-screen bg-dark-200">
      <Sidebar />
      <div className="flex-1 flex flex-col overflow-hidden">
        <Navbar user={user} />
        <main className="flex-1 overflow-hidden">
          <Outlet />
        </main>
      </div>
    </div>
  );
};

export default Layout;
