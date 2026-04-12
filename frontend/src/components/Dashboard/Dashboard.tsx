import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { FileText, Clock, Users, Trash2, MoreVertical, Copy } from 'lucide-react';
import { useAppSelector } from '../../hooks/useRedux';
import { documentService } from '../../services/api';
import { Document } from '../../types';
import toast from 'react-hot-toast';
import { formatDistanceToNow } from '../../lib/utils';

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useAppSelector(state => state.auth);
  const [documents, setDocuments] = useState<Document[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [menuOpen, setMenuOpen] = useState<string | null>(null);

  useEffect(() => {
    fetchDocuments();
  }, []);

  const fetchDocuments = async () => {
    try {
      const docs = await documentService.list();
      setDocuments(docs);
    } catch (error) {
      toast.error('Failed to load documents');
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this document?')) return;
    try {
      await documentService.delete(id);
      setDocuments(prev => prev.filter(doc => doc.id !== id));
      toast.success('Document deleted');
    } catch {
      toast.error('Failed to delete document');
    }
    setMenuOpen(null);
  };

  const handleDuplicate = async (doc: Document) => {
    try {
      const newDoc = await documentService.create(`${doc.title} (Copy)`, doc.content);
      toast.success('Document duplicated');
      navigate(`/documents/${newDoc.id}`);
    } catch (error) {
      toast.error('Failed to duplicate document');
    }
    setMenuOpen(null);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500" />
      </div>
    );
  }

  return (
    <div className="p-8 h-full overflow-auto">
      <div className="max-w-6xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Welcome back, {user?.name}</h1>
          <p className="text-slate-400">Create and manage your professional resumes</p>
        </div>
      
        {(documents?.length ?? 0) === 0 ? (
          <div className="text-center py-20 bg-dark-100 rounded-2xl border border-slate-700/50">
            <div className="w-20 h-20 mx-auto mb-6 rounded-2xl bg-slate-700/50 flex items-center justify-center">
              <FileText className="w-10 h-10 text-slate-400" />
            </div>
            <h2 className="text-xl font-semibold text-white mb-2">No documents yet</h2>
            <p className="text-slate-400 mb-6">Create your first LaTeX resume to get started</p>
            <button
              onClick={async () => {
                const doc = await documentService.create('Untitled Resume');
                navigate(`/documents/${doc.id}`);
              }}
              className="px-6 py-3 bg-primary-500 hover:bg-primary-600 text-white rounded-xl font-medium transition-all"
            >
              Create Document
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {documents.map((doc) => (
              <div
                key={doc.id}
                className="group bg-dark-100 rounded-2xl border border-slate-700/50 hover:border-primary-500/30 transition-all duration-300 overflow-hidden cursor-pointer"
                onClick={() => navigate(`/documents/${doc.id}`)}
              >
                <div className="h-32 bg-gradient-to-br from-primary-500/10 to-cyan-500/10 flex items-center justify-center">
                  <div className="w-16 h-20 bg-white rounded shadow-lg transform -rotate-3" />
                </div>
                <div className="p-4">
                  <div className="flex items-start justify-between mb-2">
                    <h3 className="font-semibold text-white truncate flex-1">{doc.title}</h3>
                    <div className="relative">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setMenuOpen(menuOpen === doc.id ? null : doc.id);
                        }}
                        className="p-1 text-slate-400 hover:text-white rounded-lg hover:bg-slate-700/50 transition-all"
                      >
                        <MoreVertical className="w-4 h-4" />
                      </button>
                      {menuOpen === doc.id && (
                        <div className="absolute right-0 top-8 w-40 bg-dark-200 border border-slate-700 rounded-xl shadow-xl z-10 overflow-hidden">
                          <button
                            onClick={(e) => { e.stopPropagation(); handleDuplicate(doc); }}
                            className="flex items-center gap-2 w-full px-4 py-2 text-sm text-slate-300 hover:bg-slate-700/50 hover:text-white"
                          >
                            <Copy className="w-4 h-4" />
                            Duplicate
                          </button>
                          <button
                            onClick={(e) => { e.stopPropagation(); handleDelete(doc.id); }}
                            className="flex items-center gap-2 w-full px-4 py-2 text-sm text-red-400 hover:bg-red-500/10"
                          >
                            <Trash2 className="w-4 h-4" />
                            Delete
                          </button>
                        </div>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-4 text-xs text-slate-400">
                    <span className="flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      {formatDistanceToNow(new Date(doc.updated_at))}
                    </span>
                    {(doc.collaborators?.length ?? 0) > 0 && (
                      <span className="flex items-center gap-1">
                        <Users className="w-3 h-3" />
                        {(doc.collaborators?.length ?? 0) + 1}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default Dashboard;
