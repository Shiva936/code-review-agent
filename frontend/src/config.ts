export type FrontendConfig = {
  apiBaseUrl: string;
  auth: {
    username: string;
    password: string;
  };
};

const STORAGE_KEY = "code-review-agent:config:v1";

function envString(key: string): string {
  const v = (import.meta as any).env?.[key];
  return typeof v === "string" ? v : "";
}

export function loadConfig(): FrontendConfig {
  const apiBaseUrl = envString("VITE_API_URL") || "http://localhost:8080";

  const envUser = envString("VITE_AUTH_USERNAME");
  const envPass = envString("VITE_AUTH_PASSWORD");

  const fallback: FrontendConfig = {
    apiBaseUrl,
    auth: { username: envUser, password: envPass },
  };

  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return fallback;
    const parsed = JSON.parse(raw) as Partial<FrontendConfig>;
    return {
      apiBaseUrl: typeof parsed.apiBaseUrl === "string" && parsed.apiBaseUrl.trim() ? parsed.apiBaseUrl.trim() : fallback.apiBaseUrl,
      auth: {
        username:
          typeof parsed.auth?.username === "string" ? parsed.auth.username : fallback.auth.username,
        password:
          typeof parsed.auth?.password === "string" ? parsed.auth.password : fallback.auth.password,
      },
    };
  } catch {
    return fallback;
  }
}

export function saveConfig(cfg: FrontendConfig) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(cfg));
}

export function clearConfig() {
  localStorage.removeItem(STORAGE_KEY);
}

export function basicAuthHeader(username: string, password: string): string | null {
  if (!username && !password) return null;
  const token = btoa(`${username}:${password}`);
  return `Basic ${token}`;
}

