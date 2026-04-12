export interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface Document {
  id: string;
  title: string;
  content: string;
  user_id: string;
  is_public: boolean;
  collaborators: string[];
  created_at: string;
  updated_at: string;
}

export interface CompileRequest {
  document_id: string;
  content: string;
}

export interface CompileResponse {
  success: boolean;
  pdf?: string;
  error?: string;
  log?: string;
}

export interface WebSocketMessage {
  type: 'content_change' | 'cursor_move' | 'user_joined' | 'user_left' | 'user_list' | 'document_update' | 'sync';
  document_id: string;
  user_id: string;
  content?: string | CursorPosition | string[];
  timestamp: string;
}

export interface CursorPosition {
  line: number;
  column: number;
}

export interface Collaborator {
  id: string;
  name: string;
  color: string;
  cursor: CursorPosition;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
}

export interface EditorState {
  document: Document | null;
  content: string;
  isCompiling: boolean;
  pdfUrl: string | null;
  collaborators: Collaborator[];
  isConnected: boolean;
}
