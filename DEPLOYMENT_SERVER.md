# MailGo — Production Deployment Guide

**Server:** Ubuntu 24 — `/home` for projects  
**SSL:** Cloudflare (no certbot needed)  
**Nginx:** Already running on port 80, serving vendorex.shop and dg.firacode.ir  
**Docker:** Already installed  

---

## Prerequisites

- DNS: `mail.tracix.net` A record pointing to your server IP (Cloudflare proxied)
- Cloudflare SSL mode: **Full** (not Full Strict) — since Nginx has no real cert
- Ports 25, 587, 993 open on your server firewall (for Mailcow mail delivery)

---

## Step 1 — Clone the repo

```bash
cd /home
git clone git@github.com:vhamed02/mailgo.git
cd mailgo
```

---

## Step 2 — Create production env file

```bash
cp .env.production.example .env.production
nano .env.production
```

Fill in every value. Generate secrets inline:

```bash
# Generate strong passwords
openssl rand -hex 32   # use for DB_PASSWORD
openssl rand -hex 32   # use for REDIS_PASSWORD
openssl rand -hex 32   # use for JWT_SECRET
```

Final `.env.production` should look like:

```env
APP_ENV=production
APP_PORT=8080
LOG_LEVEL=info

DB_USER=mailgo
DB_PASSWORD=<generated>

REDIS_PASSWORD=<generated>

JWT_SECRET=<generated>

MAILCOW_API_URL=http://mailcow-nginx:80/api/v1
MAILCOW_API_KEY=                          # fill after Step 6

IMAP_HOST=mailcow-dovecot
IMAP_PORT=993
IMAP_TLS=true

BREVO_API_KEY=xkeysib-REDACTED
BREVO_API_URL=https://api.brevo.com/v3
BREVO_FROM_EMAIL=info@yerevan.digital
BREVO_FROM_NAME=MailGo

NEXT_PUBLIC_API_URL=https://mail.tracix.net/api/v1
```

---

## Step 3 — Add Nginx server block for mail.tracix.net

Your existing Nginx config is at `/etc/nginx/sites-available/`. Add a new block for MailGo:

```bash
nano /etc/nginx/sites-available/mailgo
```

Paste this:

```nginx
server {
    listen 80;
    server_name mail.tracix.net;

    # Frontend (Next.js on port 3001)
    location / {
        proxy_pass         http://127.0.0.1:3001;
        proxy_http_version 1.1;
        proxy_set_header   Upgrade $http_upgrade;
        proxy_set_header   Connection 'upgrade';
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # Backend API
    location /api/ {
        proxy_pass         http://127.0.0.1:8082;
        proxy_http_version 1.1;
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }

    location /health {
        proxy_pass http://127.0.0.1:8082;
    }
}
```

Enable it and reload:

```bash
ln -s /etc/nginx/sites-available/mailgo /etc/nginx/sites-enabled/mailgo
nginx -t
systemctl reload nginx
```

---

## Step 4 — Open firewall ports for mail

```bash
ufw allow 25/tcp
ufw allow 465/tcp
ufw allow 587/tcp
ufw allow 993/tcp
ufw allow 995/tcp
ufw status
```

---

## Step 5 — Build and start MailGo containers

```bash
cd /home/mailgo
docker compose -f docker-compose.prod.yml up -d --build
```

Wait ~2 minutes for builds. Verify:

```bash
docker compose -f docker-compose.prod.yml ps
curl http://localhost:8082/health
```

Expected: `{"service":"mailgo-api","status":"healthy"}`

---

## Step 6 — Set up Mailcow

Mailcow is a full mail server stack (15 containers). It runs separately alongside MailGo.

```bash
cd /home
git clone https://github.com/mailcow/mailcow-dockerized
cd mailcow-dockerized
```

Run the config generator:

```bash
./generate_config.sh
```

When prompted:
- **Hostname:** `mail.tracix.net`
- **Timezone:** your timezone (e.g. `Asia/Tehran`)

After generation, edit `mailcow.conf` to avoid port conflicts:

```bash
nano mailcow.conf
```

Change these lines:

```
HTTP_PORT=8083
HTTP_BIND=127.0.0.1
HTTPS_PORT=8444
HTTPS_BIND=127.0.0.1
```

Start Mailcow:

```bash
docker compose pull
docker compose up -d
```

This takes 3–5 minutes on first run.

---

## Step 7 — Add Nginx block for Mailcow admin

```bash
nano /etc/nginx/sites-available/mailcow-admin
```

```nginx
server {
    listen 80;
    server_name admin-mail.tracix.net;

    location / {
        proxy_pass         http://127.0.0.1:8083;
        proxy_http_version 1.1;
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }
}
```

```bash
ln -s /etc/nginx/sites-available/mailcow-admin /etc/nginx/sites-enabled/mailcow-admin
nginx -t && systemctl reload nginx
```

Add `admin-mail.tracix.net` A record in Cloudflare pointing to same server IP.

---

## Step 8 — Get Mailcow API key

Open `https://admin-mail.tracix.net` in your browser.

Login: `admin` / `moohoo` (change immediately after first login)

Navigate to: **Configuration → Access → API**
- Enable Read-Write API access
- Add your server IP to allowed IPs (or `0.0.0.0/0` for all)
- Copy the API key

Update `.env.production`:

```bash
nano /home/mailgo/.env.production
# Set: MAILCOW_API_KEY=your-key-here
```

Restart MailGo worker and API to pick up the new key:

```bash
cd /home/mailgo
docker compose -f docker-compose.prod.yml restart control-plane worker
```

---

## Step 9 — Connect MailGo to Mailcow Docker network

MailGo's worker needs to reach Mailcow's internal containers (Dovecot for IMAP, Nginx for API). Connect them:

```bash
# Find Mailcow's network name
docker network ls | grep mailcow

# Connect MailGo worker and API to Mailcow network
docker network connect mailcowdockerized_mailcow-network mailgo-api
docker network connect mailcowdockerized_mailcow-network mailgo-worker
```

---

## Step 10 — Add DNS records

At Cloudflare DNS for `tracix.net`:

| Type | Name | Value | Proxy |
|------|------|-------|-------|
| A | `mail` | `your-server-IP` | ✅ Proxied |
| A | `admin-mail` | `your-server-IP` | ✅ Proxied |
| MX | `yerevan.digital` | `mail.tracix.net` | ❌ DNS only |

> **Note:** MX records must be DNS-only (not proxied) — Cloudflare cannot proxy mail traffic.

---

## Step 11 — Verify everything

```bash
# MailGo API
curl https://mail.tracix.net/health

# MailGo frontend
curl -I https://mail.tracix.net

# Worker logs
docker logs mailgo-worker --tail 20

# API logs
docker logs mailgo-api --tail 20
```

---

## Post-deploy checklist

- [ ] `https://mail.tracix.net` loads the MailGo login page
- [ ] Register an account and log in
- [ ] Add `yerevan.digital` domain — status goes to "Setting up" then "DNS Required"
- [ ] Mailcow admin accessible at `https://admin-mail.tracix.net`
- [ ] Worker can provision mailboxes (check `docker logs mailgo-worker`)
- [ ] Send email from webmail — appears in Brevo logs
- [ ] Sent folder shows the message

---

## Updates (future deploys)

```bash
cd /home/mailgo
git pull origin main
docker compose -f docker-compose.prod.yml up -d --build
```
