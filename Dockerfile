FROM node:20-alpine AS frontend-build
WORKDIR /app/front
COPY front/package.json front/postcss.config.mjs front/vite.config.ts ./
COPY front/src ./src
COPY front/index.html ./index.html
COPY front/src/styles ./src/styles
RUN npm install && npm run build

FROM golang:1.22-alpine AS backend-build
WORKDIR /app
RUN apk add --no-cache build-base
COPY backend ./backend
COPY --from=frontend-build /app/front/dist ./frontend
WORKDIR /app/backend/cmd/server
RUN go build -o /app/server .

FROM alpine:3.20
WORKDIR /app
COPY --from=backend-build /app/server ./server
COPY --from=backend-build /app/frontend ./frontend
EXPOSE 8080
CMD ["./server"]

