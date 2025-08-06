# Docker Route CLI

A powerful CLI tool for deploying and managing multiple application types (Next.js, React.js, Node.js) using Docker containers with automatic subdomain routing.

## Features

- 🚀 **Multiple Framework Support**: Deploy Next.js, React.js, and Node.js applications
- 🌐 **Automatic Subdomain Routing**: Each deployment gets `preview-{container-name}.dock-route.local`
- 🐳 **Docker Integration**: Built-in Docker container management
- 🔄 **Reverse Proxy**: Dynamic routing with built-in reverse proxy
- 📦 **Template System**: Extensible Docker templates for different app types
- ⚡ **Hot Reloading**: Live file mounting for development


## Quick Start

### 1. Setup DNS
Add to `/etc/hosts`:
```text
127.0.0.1 *.dock-route.localhost
```

### 2. Deploy an Application
Deploy a Next.js app
dock-route deploy nextjs my-next-app ./my-nextjs-project
Deploy a React.js app
dock-route deploy reactjs my-react-app ./my-react-project
Deploy a Node.js app
dock-route deploy nodejs my-api ./my-nodejs-api



### 3. Access Your Application
Your app will be available at:
- `http://preview-my-next-app.dock-route.local:8080`

## Commands

### Deploy Applications

Basic deployment
dock-route deploy [app-type] [container-name] [source-path]
With custom options
dock-route deploy nextjs my-app ./src --image custom-image:v1 --host-port 8082


### List Resources
dock-routes list templates

List running containers
dock-route list containers

### Remove Deployments
Remove container
dock-route remove my-app

Force remove with image cleanup
dock-route remove my-app --force --remove-image



## Supported Application Types

### Next.js (`nextjs`)
- **Port**: 3000
- **Features**: SSR, Static Generation, API Routes
- **Requirements**: `package.json`, `next.config.js`

### React.js (`reactjs`)
- **Port**: 80 (via Nginx)
- **Features**: SPA, Production build
- **Requirements**: `package.json`, build script

### Node.js (`nodejs`)
- **Port**: 3000
- **Features**: Express, APIs, Microservices
- **Requirements**: `package.json`, main entry point

## Configuration

### Global Configuration
Create `~/.dock-route.yaml`:

port: "8080"
domain: "dock-route.local"
docker:
registry: "localhost:5000"


### Template Customization
Templates are located in `internal/templates/data/` directory:

data/
├── nextjs/
│ ├── Dockerfile
│ └── template.yaml
├── reactjs/
│ ├── Dockerfile
│ └── template.yaml
└── nodejs/
├── Dockerfile
└── template.yaml



## Advanced Usage
```bash
dock-route deploy nextjs my-app ./src --image my-registry/nextjs:custom
```

### Port Configuration
Use different proxy port

```bash
dock-route deploy nextjs my-app ./src --port 9000
```

Use different container host port
```bash
dock-route deploy nextjs my-app ./src --host-port 8083
```

### Custom Docker Images
```bash
dock-route deploy nextjs my-app ./src --image my-registry/nextjs:custom
```


### Environment Variables
Set in template.yaml:

environment:
NODE_ENV: "development"
API_URL: "http://localhost:3001"
CUSTOM_VAR: "value"


### Project Structure



### Adding New Templates
1. Create directory in `templates/`
2. Add `Dockerfile` and `template.yaml`
3. Test with `dock-route deploy [new-type] test-app <absolute path>/test-project`

### Building

Development build
go build -o dock-route .
Production build
go build -ldflags "-s -w" -o dock-route .


## Troubleshooting

### Common Issues

**Container fails to start:**

Check container logs
docker logs preview-[container-name]
Check if port is available
lsof -i :8081


**Subdomain not accessible:**
- Verify `/etc/hosts` entry
- Check proxy server is running
- Confirm container port binding

**Build failures:**
- Ensure Dockerfile exists in template
- Check source path is correct
- Verify Docker daemon is running

### Debug Mode
Enable verbose logging
dock-route deploy nextjs my-app ./src --verbose

Check container status
dock-route list containers

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/new-template`)
3. Add tests for new functionality
4. Submit pull request

## License

MIT License - see LICENSE file for details.
