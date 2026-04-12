import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Document, Collaborator } from '../../types';

interface EditorState {
  document: Document | null;
  content: string;
  isCompiling: boolean;
  pdfUrl: string | null;
  pdfData: Uint8Array | null;
  collaborators: Collaborator[];
  isConnected: boolean;
  compileError: string | null;
}

const initialState: EditorState = {
  document: null,
  content: '',
  isCompiling: false,
  pdfUrl: null,
  pdfData: null,
  collaborators: [],
  isConnected: false,
  compileError: null,
};

const editorSlice = createSlice({
  name: 'editor',
  initialState,
  reducers: {
    setDocument: (state, action: PayloadAction<Document>) => {
      state.document = action.payload;
      state.content = action.payload.content;
    },
    setContent: (state, action: PayloadAction<string>) => {
      state.content = action.payload;
    },
    setCompiling: (state, action: PayloadAction<boolean>) => {
      state.isCompiling = action.payload;
    },
    setPdfData: (state, action: PayloadAction<{ pdfUrl: string; pdfData: Uint8Array } | null>) => {
      if (action.payload) {
        state.pdfUrl = action.payload.pdfUrl;
        state.pdfData = action.payload.pdfData;
      } else {
        state.pdfUrl = null;
        state.pdfData = null;
      }
    },
    setCollaborators: (state, action: PayloadAction<Collaborator[]>) => {
      state.collaborators = action.payload;
    },
    addCollaborator: (state, action: PayloadAction<Collaborator>) => {
      if (!state.collaborators.find(c => c.id === action.payload.id)) {
        state.collaborators.push(action.payload);
      }
    },
    removeCollaborator: (state, action: PayloadAction<string>) => {
      state.collaborators = state.collaborators.filter(c => c.id !== action.payload);
    },
    setConnected: (state, action: PayloadAction<boolean>) => {
      state.isConnected = action.payload;
    },
    setCompileError: (state, action: PayloadAction<string | null>) => {
      state.compileError = action.payload;
    },
    clearEditor: (state) => {
      state.document = null;
      state.content = '';
      state.pdfUrl = null;
      state.pdfData = null;
      state.collaborators = [];
      state.isConnected = false;
    },
  },
});

export const {
  setDocument,
  setContent,
  setCompiling,
  setPdfData,
  setCollaborators,
  addCollaborator,
  removeCollaborator,
  setConnected,
  setCompileError,
  clearEditor,
} = editorSlice.actions;

export default editorSlice.reducer;
