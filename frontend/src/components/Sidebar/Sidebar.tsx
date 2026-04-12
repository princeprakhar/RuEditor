import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { FileText, LogOut } from 'lucide-react';
import { useAppDispatch, useAppSelector } from '../../hooks/useRedux';
import { logout } from '../../store/slices/authSlice';

const Sidebar: React.FC = () => {
  const location = useLocation();
  const dispatch = useAppDispatch();
  const { user } = useAppSelector(state => state.auth);

  const handleLogout = () => {
    dispatch(logout());
  };

  return (
    <aside className="w-64 bg-dark-100 border-r border-slate-700/50 flex flex-col">
      <div className="p-6 border-b border-slate-700/50">
        <Link to="/" className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center shadow-lg shadow-primary-500/20">
            <FileText className="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 className="text-lg font-bold text-white">RuEditor</h1>
            <p className="text-xs text-slate-400">LaTeX Resume Editor</p>
          </div>
        </Link>
      </div>

      <div className="flex-1 p-4">
        <Link
          to="/"
          className={`flex items-center gap-3 px-4 py-3 rounded-xl mb-2 transition-all duration-200 ${
            location.pathname === '/'
              ? 'bg-primary-500/10 text-primary-400 border border-primary-500/20'
              : 'text-slate-300 hover:bg-slate-700/50'
          }`}
        >
          <FileText className="w-5 h-5" />
          <span className="font-medium">My Documents</span>
        </Link>
      </div>

      <div className="p-4 border-t border-slate-700/50">
        {user && (
          <div className="flex items-center gap-3 px-4 py-3 mb-2">
            <div className="w-8 h-8 rounded-full bg-gradient-to-br from-emerald-400 to-cyan-500 flex items-center justify-center text-white text-sm font-semibold">
              {user.name.charAt(0).toUpperCase()}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-white truncate">{user.name}</p>
              <p className="text-xs text-slate-400 truncate">{user.email}</p>
            </div>
          </div>
        )}
        <button
          onClick={handleLogout}
          className="flex items-center gap-3 w-full px-4 py-3 rounded-xl text-slate-300 hover:bg-red-500/10 hover:text-red-400 transition-all duration-200"
        >
          <LogOut className="w-5 h-5" />
          <span className="font-medium">Sign Out</span>
        </button>
      </div>
    </aside>
  );
};

export default Sidebar;
