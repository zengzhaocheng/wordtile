/// <reference types="vitest" />
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { VitePWA } from "vite-plugin-pwa";

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: "autoUpdate",
      injectRegister: "auto",
      manifest: {
        name: "WordTile 单词匹配学习",
        short_name: "WordTile",
        description: "上下分屏点击匹配，主动回忆 + 即时反馈的词汇学习应用",
        theme_color: "#6366f1",
        background_color: "#f8fafc",
        display: "standalone",
        start_url: "/",
        icons: [
          { src: "/pwa-192.png", sizes: "192x192", type: "image/png" },
          { src: "/pwa-512.png", sizes: "512x512", type: "image/png" },
          { src: "/pwa-512.png", sizes: "512x512", type: "image/png", purpose: "maskable" },
        ],
      },
      workbox: {
        globPatterns: ["**/*.{js,css,html,svg,png}"],
        runtimeCaching: [
          {
            urlPattern: /^\/api\/.*/i,
            handler: "NetworkFirst",
            options: {
              cacheName: "api-cache",
              expiration: { maxEntries: 100, maxAgeSeconds: 300 },
            },
          },
        ],
      },
    }),
  ],
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["src/test/setup.ts"],
  },
});
