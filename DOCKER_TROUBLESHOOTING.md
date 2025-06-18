# BlackHole Blockchain - Docker Troubleshooting Guide

## 🔧 Common Issues and Permanent Solutions

### **Issue 1: Docker Desktop Not Running**
```
ERROR: error during connect: Head "http://%2F%2F.%2Fpipe%2FdockerDesktopLinuxEngine/_ping"
```

**Root Cause:** Docker Desktop is not running or not properly started.

**Permanent Solution:**
1. **Start Docker Desktop:**
   - Check system tray for Docker Desktop icon
   - If not running, start Docker Desktop from Start menu
   - Wait 2-3 minutes for full startup (Docker Desktop takes time to initialize)

2. **Verify Docker is Running:**
   ```bash
   docker info
   ```
   This should return Docker system information without errors.

3. **If Docker Desktop won't start:**
   - Restart Docker Desktop
   - Restart your computer
   - Check Windows Services for "Docker Desktop Service"

### **Issue 2: Go Module Resolution Errors**
```
module github.com/Shivam-Patel-G/blackhole-blockchain/core/relay-chain@latest found, 
but does not contain package github.com/Shivam-Patel-G/blackhole-blockchain/core/relay-chain/governance
```

**Root Cause:** Enhanced modules (governance, monitoring, validation) exist locally but not in GitHub main branch.

**Permanent Solution:**
1. **Merge Enhanced Features to Main Branch:**
   ```bash
   git checkout main
   git merge your-feature-branch
   git push origin main
   ```

2. **Wait for GitHub to Update (5-10 minutes):**
   - GitHub needs time to process the new modules
   - Go module proxy needs time to cache the changes

3. **Alternative - Use Local Deployment:**
   ```bash
   .\quick-start.bat
   ```
   This bypasses Docker and uses your local Go environment.

### **Issue 3: Port Conflicts**
```
Error starting userland proxy: listen tcp4 0.0.0.0:8080: bind: address already in use
```

**Root Cause:** Required ports (8080, 3000, 80) are already in use.

**Permanent Solution:**
1. **Find processes using ports:**
   ```bash
   netstat -ano | findstr :8080
   netstat -ano | findstr :3000
   netstat -ano | findstr :80
   ```

2. **Stop conflicting processes:**
   ```bash
   taskkill /PID <process_id> /F
   ```

3. **Or use alternative ports in docker-compose.yml**

### **Issue 4: Docker Build Failures**
```
failed to solve: process "/bin/sh -c go mod tidy" did not complete successfully
```

**Root Cause:** Go version conflicts or missing dependencies in Docker.

**Permanent Solution:**
1. **Ensure Enhanced Modules are in Main Branch** (see Issue 2)

2. **Use Local Build + Docker Deploy:**
   - Build locally first: `go build -o blockchain-node.exe .`
   - Use Dockerfile.local for deployment

3. **Alternative - Skip Docker:**
   ```bash
   .\quick-start.bat
   ```

## 🎯 **Recommended Deployment Strategy**

### **For Development and Testing:**
```bash
.\quick-start.bat
```
**Why:** 
- Uses your local Go 1.24.3 environment
- No Docker version conflicts
- All enhanced features work immediately
- Faster startup and debugging

### **For Production Docker Deployment:**
1. **First, merge to main branch:**
   ```bash
   git checkout main
   git merge feature-branch
   git push origin main
   ```

2. **Wait 10 minutes for GitHub/Go proxy to update**

3. **Then run Docker deployment:**
   ```bash
   .\deploy.bat
   ```

### **For Quick Docker Testing:**
```bash
.\deploy-simple.bat
```

## 🔍 **Diagnostic Commands**

### **Check Docker Status:**
```bash
docker info
docker version
docker-compose --version
```

### **Check Container Status:**
```bash
docker ps -a
docker-compose ps
docker-compose logs -f
```

### **Check Port Usage:**
```bash
netstat -ano | findstr :8080
netstat -ano | findstr :3000
netstat -ano | findstr :80
```

### **Check Go Module Status:**
```bash
go mod tidy
go mod download
```

## 🚀 **Quick Recovery Steps**

If Docker deployment fails:

1. **Stop all containers:**
   ```bash
   docker-compose down
   docker stop $(docker ps -aq)
   ```

2. **Use local deployment:**
   ```bash
   .\quick-start.bat
   ```

3. **Access your blockchain:**
   - Dashboard: http://localhost:8080
   - All enhanced features available

## 📋 **Pre-Deployment Checklist**

Before running `deploy.bat`:

- [ ] Docker Desktop is running (`docker info` works)
- [ ] Enhanced modules are in main branch on GitHub
- [ ] Ports 8080, 3000, 80 are available
- [ ] No other blockchain instances running
- [ ] At least 4GB RAM available for Docker

## 🎉 **Success Indicators**

Deployment is successful when:
- `docker ps` shows containers with "Up" status
- http://localhost:8080 shows blockchain dashboard
- No error messages in `docker-compose logs`
- Container restart count is 0

## 🆘 **When All Else Fails**

The local deployment always works:
```bash
.\quick-start.bat
```

This provides:
- ✅ All enhanced features (monitoring, governance, validation, load testing)
- ✅ Full blockchain functionality
- ✅ Web dashboard at http://localhost:8080
- ✅ CLI access to advanced features
- ✅ No Docker complications
