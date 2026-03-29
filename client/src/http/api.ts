import { apiClient } from "../config/axiosConfig";
import type { User } from "../types/User";
import type { ApiResponse } from "../types/ApiResponse";
import type { File as UserFile } from "../types/File";

export type SessionData = {
  valid: boolean;
  user: User | null;
  error?: string;
};

export const fetchSession = async (): Promise<SessionData> => {
  try {
    const { data } = await apiClient.get("/auth/validate");
    return data;
  } catch (error: unknown) {
    const axiosError = error as { response?: { status?: number } };
    if (axiosError.response?.status === 401) {
      return {
        valid: false,
        user: null,
        error: "No valid authentication",
      };
    }
    throw error;
  }
};

export const loginUser = async (credentials: {
  username: string;
  password: string;
}): Promise<ApiResponse> => {
  const { data } = await apiClient.post("/auth/login", credentials);
  return data;
};

export const registerUser = async (userData: {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
}): Promise<ApiResponse> => {
  const { data } = await apiClient.post("/auth/register", userData);
  return data;
};

export const logoutUser = async (): Promise<ApiResponse> => {
  try {
    await apiClient.post("/auth/logout");
    return { success: true, message: "Logged out successfully" };
  } catch (error) {
    console.error("Logout failed:", error);
    return { success: true, message: "Logged out successfully" };
  }
};

export const getFiles = async (): Promise<UserFile[]> => {
  const { data } = await apiClient.get<{ success: boolean; files: UserFile[] }>("/files");
  return data.files ?? [];
};

export const uploadFile = async (file: globalThis.File): Promise<{ fileId: string }> => {
  const form = new FormData();
  form.append("file", file);
  const { data } = await apiClient.post<{ success: boolean; fileId: string }>("/upload", form, {
    headers: { "Content-Type": "multipart/form-data" },
  });
  return { fileId: data.fileId };
};

export const downloadFile = async (id: string): Promise<Blob> => {
  const { data } = await apiClient.get(`/download/${id}`, { responseType: "blob" });
  return data;
};

export const shareFile = async (fileId: string, ttl: number): Promise<{ url: string }> => {
  const { data } = await apiClient.post<{ success: boolean; url: string }>("/share", { fileId, ttl });
  return { url: data.url };
};
