import axios from "axios";

export const baseURL = import.meta.env.VITE_API_URL ?? "/api/v1";

export const api = axios.create({ baseURL });

api.interceptors.request.use((cfg) => {
  const token = localStorage.getItem("wt.access");
  if (token) cfg.headers.Authorization = `Bearer ${token}`;
  return cfg;
});

async function doRefresh(): Promise<string | null> {
  const refresh = localStorage.getItem("wt.refresh");
  if (!refresh) return null;
  try {
    const { data } = await axios.post(`${baseURL}/auth/refresh`, { refresh_token: refresh });
    localStorage.setItem("wt.access", data.access_token);
    if (data.refresh_token) localStorage.setItem("wt.refresh", data.refresh_token);
    return data.access_token as string;
  } catch {
    localStorage.removeItem("wt.access");
    localStorage.removeItem("wt.refresh");
    localStorage.removeItem("wt.user");
    return null;
  }
}

let refreshing: Promise<string | null> | null = null;

api.interceptors.response.use(undefined, async (error) => {
  const original = error.config;
  if (
    error.response?.status === 401 &&
    original &&
    !original._retry &&
    !original.url?.includes("/auth/") &&
    localStorage.getItem("wt.refresh")
  ) {
    original._retry = true;
    refreshing = refreshing ?? doRefresh().finally(() => (refreshing = null));
    const token = await refreshing;
    if (token) {
      original.headers.Authorization = `Bearer ${token}`;
      return api(original);
    }
    window.location.href = "/login";
  }
  return Promise.reject(error);
});

export function errText(e: unknown): string {
  if (axios.isAxiosError(e)) {
    const data = e.response?.data as { error?: string } | undefined;
    return data?.error ?? e.message;
  }
  return e instanceof Error ? e.message : String(e);
}
