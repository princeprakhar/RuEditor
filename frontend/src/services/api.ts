import axios, { AxiosError } from "axios";
import { User, Document, CompileResponse } from "../types";

const API_URL = process.env.REACT_APP_API_URL || "http://localhost:8080/api/v1";

const api = axios.create({
  baseURL: API_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem("token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error),
);

api.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      localStorage.removeItem("token");
      window.location.href = "/login";
    }
    return Promise.reject(error);
  },
);

export const authService = {
  register: async (
    email: string,
    password: string,
    name: string,
  ): Promise<{ user: User; token: string }> => {
    const response = await api.post("/auth/register", {
      email,
      password,
      name,
    });
    return response.data;
  },

  login: async (
    email: string,
    password: string,
  ): Promise<{ user: User; token: string }> => {
    const response = await api.post("/auth/login", { email, password });
    return response.data;
  },
};

export const documentService = {
  list: async (): Promise<Document[]> => {
    const response = await api.get("/documents");
    return response.data.documents;
  },

  get: async (id: string): Promise<Document> => {
    const response = await api.get(`/documents/${id}`);
    return response.data;
  },

  create: async (title: string, content?: string): Promise<Document> => {
    const response = await api.post("/documents", { title, content });
    return response.data;
  },

  update: async (
    id: string,
    data: { title?: string; content?: string },
  ): Promise<Document> => {
    const response = await api.put(`/documents/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/documents/${id}`);
  },

  compile: async (id: string, content: string): Promise<CompileResponse> => {
    const response = await api.post(`/documents/${id}/compile`, { content });
    return response.data;
  },

  inviteCollaborator: async (id: string, email: string): Promise<Document> => {
    const response = await api.post(`/documents/${id}/collaborate`, { email });
    return response.data;
  },
};

export const healthService = {
  check: async (): Promise<{ status: string }> => {
    const response = await api.get("/health");
    return response.data;
  },
};

export default api;
