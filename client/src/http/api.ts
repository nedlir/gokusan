import { apiClient } from "../config/axiosConfig";
import type { User } from "../types/User";
import type { ApiResponse } from "../types/ApiResponse";

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

export const uploadFile = async (): Promise<unknown> => {
  const { data } = await apiClient.get("/upload");
  return data;
};

export const downloadFile = async (): Promise<unknown> => {
  const { data } = await apiClient.get("/download");
  return data;
};
