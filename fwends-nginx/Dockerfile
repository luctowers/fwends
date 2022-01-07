ARG FRONTEND
FROM $FRONTEND as frontend

FROM nginx:stable-alpine
COPY --from=frontend /usr/src/fwends-frontend/dist/ /usr/share/nginx/html/
COPY ./nginx.conf.template /etc/nginx/templates/
ENV NGINX_ENVSUBST_OUTPUT_DIR=/etc/nginx/
