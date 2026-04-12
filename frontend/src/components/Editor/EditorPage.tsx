import React, { useEffect, useRef, useCallback, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import Editor from '@monaco-editor/react';
import { useAppSelector, useAppDispatch } from '../../hooks/useRedux';
import {
  setDocument,
  setContent,
  setCompiling,
  setPdfData,
  setConnected,
  setCompileError,
} from '../../store/slices/editorSlice';
import { documentService } from '../../services/api';
import { websocketService } from '../../services/websocket';
import { debounce, getCollaboratorColor } from '../../lib/utils';
import { WebSocketMessage, Collaborator } from '../../types';
import toast from 'react-hot-toast';
import Split from 'react-split';
import {
  Play,
  Save,
  Users,
  Wifi,
  WifiOff,
  ChevronLeft,
  Download,
  Loader2,
} from 'lucide-react';

const EditorPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const { user } = useAppSelector(state => state.auth);
  const { content, isCompiling, pdfData, isConnected, pdfUrl } = useAppSelector(
    state => state.editor
  );
  const docTitle = useAppSelector(state => state.editor.document?.title);

  const [collaboratorList, setCollaboratorList] = useState<Collaborator[]>([]);
  const editorRef = useRef<unknown>(null);
  const monacoRef = useRef<unknown>(null);

  useEffect(() => {
    if (!id) return;
    loadDocument();
    return () => {
      websocketService.disconnect();
    };
  }, [id]);

  useEffect(() => {
    if (!id || !user) return;
    websocketService.connect(id, user.id);

    websocketService.onConnect(() => {
      dispatch(setConnected(true));
    });

    websocketService.onDisconnect(() => {
      dispatch(setConnected(false));
    });

    const unsubscribe = websocketService.onMessage((message: WebSocketMessage) => {
      handleWebSocketMessage(message);
    });

    return () => {
      unsubscribe();
    };
  }, [id, user?.id]);

  const loadDocument = async () => {
    if (!id) return;
    try {
      const doc = await documentService.get(id);
      dispatch(setDocument(doc));
    } catch {
      toast.error('Failed to load document');
      navigate('/');
    }
  };

  const handleWebSocketMessage = (message: WebSocketMessage) => {
    switch (message.type) {
      case 'content_change':
        if (message.user_id !== user?.id && typeof message.content === 'string') {
          dispatch(setContent(message.content));
          if (editorRef.current) {
            const editorInstance = editorRef.current as { getValue: () => string; setValue: (v: string) => void };
            if (editorInstance.getValue() !== message.content) {
              editorInstance.setValue(message.content);
            }
          }
        }
        break;
      case 'user_list':
        if (Array.isArray(message.content)) {
          const newCollaborators: Collaborator[] = message.content
            .filter((uid) => uid !== user?.id)
            .map((uid, index) => ({
              id: uid,
              name: `User ${index + 1}`,
              color: getCollaboratorColor(index),
              cursor: { line: 1, column: 1 },
            }));
          setCollaboratorList(newCollaborators);
        }
        break;
      case 'cursor_move':
        if (message.user_id !== user?.id && typeof message.content === 'object') {
          setCollaboratorList((prev) =>
            prev.map((c) =>
              c.id === message.user_id ? { ...c, cursor: message.content as { line: number; column: number } } : c
            )
          );
        }
        break;
    }
  };

  const debouncedSave = useCallback(
    debounce((newContent: string) => {
      if (!id) return;
      websocketService.sendContentChange(newContent);
      documentService.update(id, { content: newContent }).catch(() => {
        toast.error('Failed to save');
      });
    }, 500),
    [id]
  );

  const handleEditorChange = (value: string | undefined) => {
    const newContent = value || '';
    dispatch(setContent(newContent));
    debouncedSave(newContent);
  };

  const handleCompile = async () => {
    if (!id || !content) return;
    dispatch(setCompiling(true));
    dispatch(setCompileError(null));

    try {
      const result = await documentService.compile(id, content);
      if (result.success && result.pdf) {
        const binaryString = atob(result.pdf);
        const bytes = new Uint8Array(binaryString.length);
        for (let i = 0; i < binaryString.length; i++) {
          bytes[i] = binaryString.charCodeAt(i);
        }
        const blob = new Blob([bytes], { type: 'application/pdf' });
        const url = URL.createObjectURL(blob);
        dispatch(setPdfData({ pdfUrl: url, pdfData: bytes }));
        toast.success('Compiled successfully!');
      } else {
        dispatch(setCompileError(result.error || 'Compilation failed'));
        toast.error(result.error || 'Compilation failed');
      }
    } catch {
      dispatch(setCompileError('Compilation error'));
      toast.error('Compilation error');
    } finally {
      dispatch(setCompiling(false));
    }
  };

  const handleManualSave = async () => {
    if (!id) return;
    try {
      await documentService.update(id, { content });
      toast.success('Saved!');
    } catch {
      toast.error('Failed to save');
    }
  };

  const handleDownloadPDF = () => {
    if (!pdfData) return;
    const blob = new Blob([pdfData], { type: 'application/pdf' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'resume.pdf';
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleEditorDidMount = (editor: unknown, monaco: unknown) => {
    editorRef.current = editor;
    monacoRef.current = monaco;

    const monacoEditor = monaco as { editor: { defineTheme: (name: string, theme: object) => void; setTheme: (name: string) => void } };
    const editorInstance = editor as { addCommand: (keybinding: number, handler: () => void) => void; getValue: () => string; setValue: (v: string) => void };
    const KeyMod = (monaco as { KeyMod: { CtrlCmd: number } }).KeyMod;
    const KeyCode = (monaco as { KeyCode: { KeyS: number; KeyP: number } }).KeyCode;

    monacoEditor.editor.defineTheme('rueditor-dark', {
      base: 'vs-dark',
      inherit: true,
      rules: [],
      colors: {
        'editor.background': '#0f172a',
        'editor.foreground': '#f1f5f9',
        'editor.lineHighlightBackground': '#1e293b',
        'editor.selectionBackground': '#0ea5e944',
        'editorCursor.foreground': '#0ea5e9',
        'editorLineNumber.foreground': '#64748b',
        'editorLineNumber.activeForeground': '#94a3b8',
      },
    });

    monacoEditor.editor.setTheme('rueditor-dark');

    editorInstance.addCommand(KeyMod.CtrlCmd | KeyCode.KeyS, () => {
      handleManualSave();
    });

    editorInstance.addCommand(KeyMod.CtrlCmd | KeyCode.KeyP, () => {
      handleCompile();
    });
  };

  return (
    <div className="h-full flex flex-col bg-dark-200">
      <header className="h-14 bg-dark-100 border-b border-slate-700/50 px-4 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/')}
            className="p-2 text-slate-400 hover:text-white hover:bg-slate-700/50 rounded-lg transition-all"
          >
            <ChevronLeft className="w-5 h-5" />
          </button>
          <div>
            <h1 className="text-sm font-medium text-white">
              {docTitle || 'Loading...'}
            </h1>
            <div className="flex items-center gap-2 text-xs text-slate-400">
              {isConnected ? (
                <>
                  <Wifi className="w-3 h-3 text-emerald-400" />
                  <span>Connected</span>
                </>
              ) : (
                <>
                  <WifiOff className="w-3 h-3 text-red-400" />
                  <span>Disconnected</span>
                </>
              )}
              {collaboratorList.length > 0 && (
                <span className="flex items-center gap-1 ml-2">
                  <Users className="w-3 h-3" />
                  {collaboratorList.length + 1} users
                </span>
              )}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {collaboratorList.map((collab) => (
            <div
              key={collab.id}
              className="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-medium"
              style={{ backgroundColor: collab.color }}
              title={collab.name}
            >
              {collab.name.charAt(0).toUpperCase()}
            </div>
          ))}
          <button
            onClick={handleManualSave}
            className="flex items-center gap-2 px-3 py-1.5 text-sm text-slate-300 hover:text-white hover:bg-slate-700/50 rounded-lg transition-all"
          >
            <Save className="w-4 h-4" />
            Save
          </button>
          <button
            onClick={handleCompile}
            disabled={isCompiling}
            className="flex items-center gap-2 px-4 py-1.5 bg-primary-500 hover:bg-primary-600 text-white text-sm rounded-lg font-medium transition-all disabled:opacity-50"
          >
            {isCompiling ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Compiling...
              </>
            ) : (
              <>
                <Play className="w-4 h-4" />
                Compile
              </>
            )}
          </button>
          {pdfData && (
            <button
              onClick={handleDownloadPDF}
              className="flex items-center gap-2 px-3 py-1.5 text-sm text-slate-300 hover:text-white hover:bg-slate-700/50 rounded-lg transition-all"
            >
              <Download className="w-4 h-4" />
              PDF
            </button>
          )}
        </div>
      </header>

      <div className="flex-1 overflow-hidden">
        <Split
          className="flex h-full"
          sizes={[50, 50]}
          minSize={300}
          gutterSize={8}
          gutterStyle={() => ({
            backgroundColor: '#1e293b',
          })}
        >
          <div className="h-full overflow-hidden">
            <Editor
              height="100%"
              defaultLanguage="latex"
              value={content}
              onChange={handleEditorChange}
              onMount={handleEditorDidMount}
              options={{
                fontSize: 14,
                fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
                minimap: { enabled: false },
                lineNumbers: 'on',
                renderLineHighlight: 'line',
                scrollBeyondLastLine: false,
                wordWrap: 'on',
                automaticLayout: true,
                tabSize: 2,
                padding: { top: 16 },
              }}
            />
          </div>

          <div className="h-full bg-slate-900 flex flex-col">
            <div className="h-8 bg-dark-100 border-b border-slate-700/50 px-3 flex items-center">
              <span className="text-xs text-slate-400 font-medium">PDF Preview</span>
            </div>
            <div className="flex-1 overflow-auto p-4">
              {pdfData ? (
                <iframe
                  src={pdfUrl || ''}
                  className="w-full h-full min-h-[600px] bg-white rounded-lg shadow-xl"
                  title="PDF Preview"
                />
              ) : (
                <div className="h-full flex flex-col items-center justify-center text-slate-400">
                  <div className="w-16 h-16 mb-4 rounded-xl bg-slate-800 flex items-center justify-center">
                    <Play className="w-8 h-8" />
                  </div>
                  <p className="text-sm">Click Compile to generate PDF</p>
                  <p className="text-xs mt-1">or press Ctrl+P</p>
                </div>
              )}
            </div>
          </div>
        </Split>
      </div>
    </div>
  );
};

export default EditorPage;
