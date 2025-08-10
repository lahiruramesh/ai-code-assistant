# Use Node.js 22 or latest Node.js 20 that supports crypto.hash
FROM node:22-alpine

WORKDIR /app

# Install dependencies for development
RUN apk add --no-cache libc6-compat

# Set pnpm store directory
ENV PNPM_STORE_DIR=/app/.pnpm-store

# Install pnpm
RUN npm install -g pnpm

# Set working directory
WORKDIR /app

# Configure pnpm store
RUN pnpm config set store-dir $PNPM_STORE_DIR

# Copy package files
COPY package.json pnpm-lock.yaml ./

# Install dependencies
RUN pnpm install

# Copy rest of application
COPY . .
# Expose port
EXPOSE 3000

# Set environment variables for Vite
ENV VITE_HOST=0.0.0.0
ENV VITE_PORT=3000

# Default command with host binding
CMD ["pnpm", "run", "dev", "--host", "0.0.0.0", "--port", "3000"]
