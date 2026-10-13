import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import App from "./App";
import { I18nProvider } from "./i18n";
import { applyTheme, useThemeStore } from "./stores/theme";
import "./index.css";

// 初始化主题（在 React 树挂载前应用，避免闪烁）
applyTheme(useThemeStore.getState().mode);

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, refetchOnWindowFocus: false, staleTime: 30_000 },
  },
});

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <I18nProvider>
          <App />
        </I18nProvider>
      </BrowserRouter>
    </QueryClientProvider>
  </React.StrictMode>,
);
