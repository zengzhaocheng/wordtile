FROM node:20-alpine AS build
# 国内网络访问 npmjs 官方源不稳定，使用 npmmirror
RUN npm config set registry https://registry.npmmirror.com
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY . .
# 前后端分离部署时可传入后端地址；留空则使用同源的 /api/v1（由 nginx 反代）
ARG VITE_API_URL=""
RUN if [ -n "$VITE_API_URL" ]; then \
      VITE_API_URL="$VITE_API_URL" npm run build; \
    else \
      npm run build; \
    fi

FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
