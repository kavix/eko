FROM node:24-alpine AS build

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

RUN npm run build

FROM nginx:1.31.1-alpine AS serve

RUN rm -rf /usr/share/nginx/html

COPY --from=build /app/out/ /usr/share/nginx/html
