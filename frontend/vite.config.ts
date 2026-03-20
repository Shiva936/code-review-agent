import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const apiUrl = process.env.REACT_APP_API_URL || "http://localhost:8080";

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/health": {
        target: apiUrl,
        changeOrigin: true,
      },
      "/run": {
        target: apiUrl,
        changeOrigin: true,
      },
      "/runs": {
        target: apiUrl,
        changeOrigin: true,
      },
    },
  },
});

