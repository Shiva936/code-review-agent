export type FrontendConfig = {
  apiBaseUrl: string;
  auth: {
    username: string;
    password: string;
  };
};

function envString(key: string): string {
  const env = import.meta.env as Record<string, string | boolean | undefined>;
  const v = env[key];
  return typeof v === "string" ? v : "";
}

export function loadConfig(): FrontendConfig {
  // Empty apiBaseUrl => same-origin requests (Vite proxy in dev).
  const apiBaseUrl = envString("VITE_API_URL") || "";

  const envUser = envString("VITE_AUTH_USERNAME");
  const envPass = envString("VITE_AUTH_PASSWORD");

  return {
    apiBaseUrl,
    auth: { username: envUser, password: envPass },
  };
}

export function basicAuthHeader(username: string, password: string): string | null {
  if (!username && !password) return null;
  const token = btoa(`${username}:${password}`);
  return `Basic ${token}`;
}

