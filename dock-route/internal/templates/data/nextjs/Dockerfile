FROM node:18-alpine

WORKDIR /app

# Install dependencies for development
RUN apk add --no-cache libc6-compat

# Copy package files
COPY package*.json ./

# Install ALL dependencies (including devDependencies)
RUN npm ci

# Copy source code (this will be overridden by bind mount in dev mode)
COPY . .

# Expose port
EXPOSE 3000

# Default command (can be overridden)
CMD ["npm", "run", "dev"]
