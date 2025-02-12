# Stage 1: Build the application
FROM node:23-alpine AS builder
WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

# Stage 2: Run the application on ARM64
FROM node:23-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production

COPY --from=builder /app .

EXPOSE 3000

CMD ["node", "build/index.js"]
