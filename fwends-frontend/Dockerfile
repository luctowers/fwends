FROM node:16-alpine3.14
WORKDIR /usr/src/fwends-frontend
COPY package.json .
RUN npm install
COPY . .
CMD [ "npm", "run", "start-dev" ]
