import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const apiUrl = env.VITE_API_URL || "http://localhost:8080";

  return {
    plugins: [react()],
    server: {
      proxy: {
        "/health": { target: apiUrl, changeOrigin: true },
        "/run": { target: apiUrl, changeOrigin: true },
        "/runs": { target: apiUrl, changeOrigin: true },
        "/run-groups": { target: apiUrl, changeOrigin: true },
      },
    },
  };
});

