FROM node:16-alpine as build
WORKDIR /usr/src/fwends-frontend
COPY package.json .
RUN npm install
COPY . .
ARG WEBPACK_MODE=production
RUN npm run build -- --mode=$WEBPACK_MODE
CMD ["npm", "run", "start"]
