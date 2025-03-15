# build Angular frontend
FROM node:22 AS nodebuilder
COPY frontend /frontend
WORKDIR /frontend
RUN npm ci
RUN npx ng build

# build GO binary
FROM golang:1.24.1 AS gobuilder
COPY backend /backend
WORKDIR /backend
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o backend-server ./server.go

# create final image
FROM debian:stable-slim
ENV SUGGESTION_DB_PATH=/data/suggestions.db
ENV SUGGESTION_FRONTEND_DIST_PATH=/frontend/dist

COPY --from=gobuilder /backend/backend-server /bin/backend-server
COPY --from=nodebuilder /frontend/dist/frontend/browser /frontend/dist

EXPOSE 8080
CMD ["/bin/backend-server"]
