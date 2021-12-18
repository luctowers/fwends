ARG FRONTEND
FROM $FRONTEND as frontend

FROM nginx:stable-alpine
COPY --from=frontend /usr/src/fwends-frontend/dist/ /usr/share/nginx/html/
COPY ./nginx.conf /etc/nginx/nginx.conf
