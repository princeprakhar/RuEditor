import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus, Search, Bell } from 'lucide-react';
import { } from '../../hooks/useRedux';
import { documentService } from '../../services/api';
import toast from 'react-hot-toast';

interface NavbarProps {
  user: {
    name: string;
    email: string;
  } | null;
}

const Navbar: React.FC<NavbarProps> = ({ user }) => {
  const navigate = useNavigate();
  const [isCreating, setIsCreating] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  const handleCreateDocument = async () => {
    setIsCreating(true);
    try {
      const doc = await documentService.create('Untitled Resume');
      toast.success('Document created!');
      navigate(`/documents/${doc.id}`);
    } catch (error) {
      toast.error('Failed to create document');
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <header className="h-16 bg-dark-100 border-b border-slate-700/50 px-6 flex items-center justify-between">
      <div className="flex items-center gap-4 flex-1">
        <div className="relative max-w-md flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
          <input
            type="text"
            placeholder="Search documents..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-dark-200 border border-slate-700/50 rounded-xl text-white placeholder-slate-400 focus:outline-none focus:border-primary-500/50 focus:ring-2 focus:ring-primary-500/20 transition-all"
          />
        </div>
      </div>

      <div className="flex items-center gap-4">
        <button
          onClick={handleCreateDocument}
          disabled={isCreating}
          className="flex items-center gap-2 px-4 py-2 bg-primary-500 hover:bg-primary-600 text-white rounded-xl font-medium transition-all duration-200 shadow-lg shadow-primary-500/20 hover:shadow-primary-500/30 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <Plus className="w-5 h-5" />
          {isCreating ? 'Creating...' : 'New Document'}
        </button>

        <button className="relative p-2 text-slate-400 hover:text-white hover:bg-slate-700/50 rounded-xl transition-all">
          <Bell className="w-5 h-5" />
          <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full" />
        </button>
      </div>
    </header>
  );
};

export default Navbar;
