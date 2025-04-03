###
# Stage 1: Builder
###
FROM node:23-alpine AS builder
WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

###
# Stage 2: Runner (production)
###
FROM node:23-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production

# Copy only the package files again
COPY package*.json ./

RUN npm install --only=production

COPY --from=builder /app/build ./build

EXPOSE 3000
CMD ["node", "build/index.js"]
