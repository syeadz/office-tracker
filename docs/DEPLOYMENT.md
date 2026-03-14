# Deployment Guide

Production deployment guide for Office Tracker.

Navigation:

- [README.md](README.md) — documentation index
- [../README.md](../README.md) — project overview and quick start

## Prerequisites

- Linux server (Ubuntu 22.04+ recommended)
- Go 1.21+ (if building from source)
- SQLite 3.35+ (usually pre-installed)
- Systemd (for service management)
- Optional: Docker and Docker Compose

## Deployment Options

### Option 1: Systemd Service (Recommended)

Best for single-server deployments with direct control.

#### 1. Build the Application

```bash
# On your local machine or build server
git clone <repository-url>
cd office-tracker-v2
go build -o office-tracker cmd/office/main.go
```

#### 2. Create Application User

```bash
sudo useradd -r -s /bin/false office-tracker
```

#### 3. Install Application

```bash
# Create application directory
sudo mkdir -p /opt/office-tracker
sudo mkdir -p /var/lib/office-tracker
sudo mkdir -p /var/log/office-tracker

# Copy binary
sudo cp office-tracker /opt/office-tracker/
sudo chmod +x /opt/office-tracker/office-tracker

# Set ownership
sudo chown -R office-tracker:office-tracker /opt/office-tracker
sudo chown -R office-tracker:office-tracker /var/lib/office-tracker
sudo chown -R office-tracker:office-tracker /var/log/office-tracker
```

#### 4. Create Configuration

```bash
sudo nano /opt/office-tracker/.env
```

Add your configuration:

```bash
# Server
HTTP_PORT=8080

# Database
DB_PATH=/var/lib/office-tracker/office.db

# Discord (optional)
DISCORD_TOKEN=your_token_here
DISCORD_EXEC_GUILD_ID=your_guild_id
DISCORD_COMMUNITY_GUILD_ID=your_guild_id
DISCORD_DASHBOARD_CHANNEL_NAME=office-tracker
DISCORD_REPORTS_CHANNEL_ID=your_channel_id

# Security
API_KEY=your_secure_api_key_here
CORS_ORIGINS=https://yourdomain.com
```

Set proper permissions:

```bash
sudo chown office-tracker:office-tracker /opt/office-tracker/.env
sudo chmod 600 /opt/office-tracker/.env
```

#### 5. Create Systemd Service

```bash
sudo nano /etc/systemd/system/office-tracker.service
```

Add the following:

```ini
[Unit]
Description=Office Tracker
After=network.target

[Service]
Type=simple
User=office-tracker
Group=office-tracker
WorkingDirectory=/opt/office-tracker
ExecStart=/opt/office-tracker/office-tracker
Restart=always
RestartSec=10

# Environment
EnvironmentFile=/opt/office-tracker/.env

# Logging
StandardOutput=append:/var/log/office-tracker/app.log
StandardError=append:/var/log/office-tracker/error.log

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/office-tracker
ReadOnlyPaths=/opt/office-tracker

[Install]
WantedBy=multi-user.target
```

#### 6. Enable and Start Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable office-tracker

# Start service
sudo systemctl start office-tracker

# Check status
sudo systemctl status office-tracker
```

#### 7. View Logs

```bash
# Application logs
sudo tail -f /var/log/office-tracker/app.log

# Error logs
sudo tail -f /var/log/office-tracker/error.log

# Systemd journal
sudo journalctl -u office-tracker -f
```

### Option 2: Docker Deployment

Best for containerized environments or multi-service deployments.

#### 1. Using Existing Dockerfile

```bash
# Build image
docker build -t office-tracker:latest -f build/Dockerfile .

# Run container
docker run -d \
  --name office-tracker \
  -p 8080:8080 \
  -v /var/lib/office-tracker:/data \
  -e DB_PATH=/data/office.db \
  -e DISCORD_TOKEN=your_token \
  --restart unless-stopped \
  office-tracker:latest
```

#### 2. Using Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  office-tracker:
    build:
      context: .
      dockerfile: build/Dockerfile
    container_name: office-tracker
    ports:
      - "8080:8080"
    volumes:
      - office-data:/data
    environment:
      - HTTP_PORT=8080
      - DB_PATH=/data/office.db
      - DISCORD_TOKEN=${DISCORD_TOKEN}
      - DISCORD_EXEC_GUILD_ID=${DISCORD_EXEC_GUILD_ID}
      - DISCORD_COMMUNITY_GUILD_ID=${DISCORD_COMMUNITY_GUILD_ID}
      - API_KEY=${API_KEY}
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  office-data:
```

Deploy:

```bash
docker-compose up -d
```

## Reverse Proxy Setup

### Nginx

```nginx
server {
    listen 80;
    server_name tracker.yourdomain.com;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name tracker.yourdomain.com;

    # SSL certificates (use certbot)
    ssl_certificate /etc/letsencrypt/live/tracker.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/tracker.yourdomain.com/privkey.pem;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    # Proxy settings
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

Apply configuration:

```bash
sudo nginx -t
sudo systemctl reload nginx
```

### Caddy

Simpler alternative with automatic HTTPS:

```text
tracker.yourdomain.com {
    reverse_proxy localhost:8080
}
```

Apply configuration:

```bash
sudo systemctl reload caddy
```

## SSL/TLS Setup

### Using Certbot (Let's Encrypt)

```bash
# Install certbot
sudo apt update
sudo apt install certbot python3-certbot-nginx

# Obtain certificate (Nginx)
sudo certbot --nginx -d tracker.yourdomain.com

# Obtain certificate (standalone, if not using Nginx)
sudo systemctl stop office-tracker
sudo certbot certonly --standalone -d tracker.yourdomain.com
sudo systemctl start office-tracker

# Auto-renewal (certbot sets this up automatically)
sudo certbot renew --dry-run
```

## Database Management

### Initial Setup

The database is created automatically on first run. No manual setup needed.

### Backups

Set up automated daily backups:

```bash
sudo nano /usr/local/bin/backup-office-tracker.sh
```

Add:

```bash
#!/bin/bash
BACKUP_DIR="/var/backups/office-tracker"
DB_PATH="/var/lib/office-tracker/office.db"
DATE=$(date +%Y%m%d)

mkdir -p "$BACKUP_DIR"

# Checkpoint and backup
sqlite3 "$DB_PATH" "PRAGMA wal_checkpoint(TRUNCATE); .backup '$BACKUP_DIR/office-$DATE.db'"

# Keep last 30 days
find "$BACKUP_DIR" -name "office-*.db" -mtime +30 -delete
```

Make executable:

```bash
sudo chmod +x /usr/local/bin/backup-office-tracker.sh
```

Add to crontab:

```bash
sudo crontab -e
```

Add line:

```text
0 3 * * * /usr/local/bin/backup-office-tracker.sh
```

See [DATABASE_MAINTENANCE.md](DATABASE_MAINTENANCE.md) for more database operations.

## Monitoring

### Health check endpoint

The application exposes a health endpoint:

```bash
curl http://localhost:8080/health
```

### Service Monitoring

Check service status:

```bash
# Service status
sudo systemctl status office-tracker

# Recent logs
sudo journalctl -u office-tracker -n 50

# Follow logs
sudo journalctl -u office-tracker -f
```

### Log Rotation

Create `/etc/logrotate.d/office-tracker`:

```text
/var/log/office-tracker/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    missingok
    create 0640 office-tracker office-tracker
    postrotate
        systemctl reload office-tracker > /dev/null 2>&1 || true
    endscript
}
```

## Updates

### Systemd Deployment

```bash
# Build new version
go build -o office-tracker cmd/office/main.go

# Stop service
sudo systemctl stop office-tracker

# Backup current binary
sudo cp /opt/office-tracker/office-tracker /opt/office-tracker/office-tracker.backup

# Copy new binary
sudo cp office-tracker /opt/office-tracker/
sudo chmod +x /opt/office-tracker/office-tracker
sudo chown office-tracker:office-tracker /opt/office-tracker/office-tracker

# Start service
sudo systemctl start office-tracker

# Verify
sudo systemctl status office-tracker
```

### Docker Deployment

```bash
# Pull/build new image
docker-compose pull
# or
docker-compose build

# Restart with new image
docker-compose up -d

# Verify
docker-compose logs -f
```

## Operational Runbook (Quick Response)

Use this when the service is degraded or down and you need a fast, repeatable recovery flow.

### 1) Quick triage (first 5 minutes)

```bash
# Health
curl -i http://localhost:8080/health

# Service state
sudo systemctl status office-tracker

# Recent logs
sudo journalctl -u office-tracker -n 100 --no-pager
```

If running via Docker:

```bash
docker-compose ps
docker-compose logs --tail=100
```

### 2) Safe restart sequence

Systemd:

```bash
sudo systemctl restart office-tracker
sleep 2
sudo systemctl status office-tracker
curl -fsS http://localhost:8080/health
```

Docker:

```bash
docker-compose restart
sleep 2
docker-compose ps
curl -fsS http://localhost:8080/health
```

### 3) If restart fails: common checks

```bash
# Port conflicts
sudo lsof -i :8080

# Database permissions/path
ls -lah /var/lib/office-tracker/

# Env file readability
ls -lah /opt/office-tracker/.env
```

### 4) Roll back quickly (systemd binary deploy)

```bash
sudo systemctl stop office-tracker
sudo cp /opt/office-tracker/office-tracker.backup /opt/office-tracker/office-tracker
sudo chown office-tracker:office-tracker /opt/office-tracker/office-tracker
sudo chmod +x /opt/office-tracker/office-tracker
sudo systemctl start office-tracker
sudo systemctl status office-tracker
curl -fsS http://localhost:8080/health
```

### 5) Incident close checklist

- Service healthy on `/health`
- No repeating errors in the last 50 log lines
- Dashboard/API reachable from expected clients
- Root cause and fix captured in your team issue tracker

## Troubleshooting

### Service Won't Start

Check logs:

```bash
sudo journalctl -u office-tracker -n 100
sudo tail -f /var/log/office-tracker/error.log
```

Common issues:

- Port already in use: Check with `sudo lsof -i :8080`
- Database permissions: Ensure office-tracker user can write to DB_PATH
- Missing environment variables: Verify .env file exists and is readable

### Database Locked

If you see "database is locked" errors:

```bash
# Check for stale connections
sudo lsof /var/lib/office-tracker/office.db

# Checkpoint WAL
sqlite3 /var/lib/office-tracker/office.db "PRAGMA wal_checkpoint(RESTART);"

# Restart service
sudo systemctl restart office-tracker
```

### High Memory Usage

Check database size:

```bash
du -h /var/lib/office-tracker/office.db*
```

Run maintenance:

```bash
sqlite3 /var/lib/office-tracker/office.db "VACUUM; ANALYZE;"
```

## Security Checklist

- [ ] Use strong API_KEY (32+ random characters)
- [ ] Enable firewall (allow only 80, 443, 22)
- [ ] Use HTTPS in production
- [ ] Set restrictive file permissions (600 for .env)
- [ ] Run as non-root user
- [ ] Keep Go and dependencies updated
- [ ] Regular security updates: `sudo apt update && sudo apt upgrade`
- [ ] Set up automatic backups
- [ ] Monitor logs for suspicious activity
- [ ] Use strong Discord bot token

## Performance Tuning

For high-traffic deployments:

```bash
# Increase file descriptor limit
sudo nano /etc/systemd/system/office-tracker.service
```

Add under `[Service]`:

```ini
LimitNOFILE=65536
```

Reload and restart:

```bash
sudo systemctl daemon-reload
sudo systemctl restart office-tracker
```

## Scaling

For large deployments:

1. **Read Replicas**: SQLite does not support replication natively.
  For multi-server setups, consider PostgreSQL.
2. **Load Balancer**: Use nginx or HAProxy to distribute load across multiple instances.
3. **Separate Discord Bot**: Run Discord bot as a separate service
  if it impacts HTTP performance.
4. **Caching**: Add Redis for frequently accessed data (user stats, leaderboards).

## Support

For issues or questions:

- Check logs: `/var/log/office-tracker/`
- Review [DATABASE_MAINTENANCE.md](DATABASE_MAINTENANCE.md)
- Review [DISCORD_SETUP.md](DISCORD_SETUP.md)
- Check GitHub issues
