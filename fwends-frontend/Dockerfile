FROM node:16-alpine as build
WORKDIR /usr/src/fwends-frontend
COPY package.json .
RUN npm install
COPY . .
RUN npm run build
CMD [ "npm", "run", "start" ]

FROM nginx:stable-alpine
COPY --from=build /usr/src/fwends-frontend/dist/ /usr/share/nginx/html/
COPY ./nginx.conf /etc/nginx/nginx.conf
