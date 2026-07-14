# Mailbox - Production Deployment Guide

**Server:** Ubuntu 24 - `/home` for projects
**Public website/API:** `mailbox.yerevan.digital` through Cloudflare proxy
**Mail server hostname:** `mail.tracix.net` DNS-only, used for MX and mail protocols
**SSL:** Cloudflare for public web traffic; no certbot needed for Mailbox
**Nginx:** Already running on port 80, serving yerevan.digital and dg.firacode.ir
**Docker:** Already installed

No new server packages are required by this plan. New runtime services run under Docker containers. The only host-level changes are DNS, firewall rules, and reverse-proxy config in the already installed Nginx.

---

## Prerequisites

- DNS: `mailbox.yerevan.digital` A record pointing to your server IP, Cloudflare proxied
- DNS: `mail.tracix.net` A record pointing to your server IP, DNS-only
- Cloudflare SSL mode for `mailbox.yerevan.digital`: **Full** (not Full Strict) since Nginx has no real cert
- Ports 25, 465, 587, 993, and 995 open on the server firewall for Mailcow mail traffic
- MX records for hosted mail domains point to `mail.tracix.net`, DNS-only

---

## Step 1 - Clone the repo

```bash
cd /home
git clone git@github.com:vhamed02/mailgo.git
cd mailgo
```

---

## Step 2 - Create production env file

```bash
cp .env.production.example .env.production
nano .env.production
```

Fill in every value. Generate secrets inline:

```bash
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

MAILCOW_API_URL=http://nginx:8083/api/v1
MAILCOW_API_KEY=                          # fill after Mailcow API setup

IMAP_HOST=dovecot
IMAP_PORT=993
IMAP_TLS=true

BREVO_API_KEY=xkeysib-REDACTED
BREVO_API_URL=https://api.brevo.com/v3
BREVO_FROM_EMAIL=info@yerevan.digital
BREVO_FROM_NAME=Mailbox

NEXT_PUBLIC_API_URL=https://mailbox.yerevan.digital/api/v1
```

`MAILCOW_API_URL` and `IMAP_HOST` use Mailcow's Docker network aliases. Mailbox joins the Mailcow network through `docker-compose.prod.yml`; no manual `docker network connect` is needed.

---

## Step 3 - Set up Mailcow

Mailcow is a full mail server stack and runs separately alongside Mailbox.

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
- **Timezone:** your timezone, for example `Asia/Tehran`

After generation, edit `mailcow.conf` to avoid web port conflicts with the existing host Nginx:

```bash
nano mailcow.conf
```

Set these lines:

```env
HTTP_PORT=8083
HTTP_BIND=127.0.0.1
HTTPS_PORT=8444
HTTPS_BIND=127.0.0.1
SKIP_CLAMD=y
```

Start Mailcow:

```bash
docker compose pull
docker compose up -d
```

This creates the `mailcowdockerized_mailcow-network` Docker network that Mailbox uses.

---

## Step 4 - Add Nginx block for Mailbox website

Your existing Nginx config is at `/etc/nginx/sites-available/`. Add a new block for Mailbox:

```bash
nano /etc/nginx/sites-available/mailgo
```

Paste this:

```nginx
server {
    listen 80;
    server_name mailbox.yerevan.digital;

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

## Step 5 - Add Nginx block for Mailcow admin

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
        proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }
}
```

```bash
ln -s /etc/nginx/sites-available/mailcow-admin /etc/nginx/sites-enabled/mailcow-admin
nginx -t && systemctl reload nginx
```

Add `admin-mail.tracix.net` A record in Cloudflare pointing to the same server IP. This admin hostname can be proxied because it is web traffic only.

---

## Step 6 - Open firewall ports for mail

```bash
ufw allow 25/tcp
ufw allow 465/tcp
ufw allow 587/tcp
ufw allow 993/tcp
ufw allow 995/tcp
ufw status
```

---

## Step 7 - Get Mailcow API key

Open `https://admin-mail.tracix.net` in your browser.

Login: `admin` / `moohoo` and change the password immediately after first login.

Navigate to: **Configuration -> Access -> API**

- Enable Read-Write API access
- Add your server IP to allowed IPs, or `0.0.0.0/0` for all
- Copy the API key

Update `.env.production`:

```bash
nano /home/mailgo/.env.production
# Set: MAILCOW_API_KEY=your-key-here
```

---

## Step 8 - Build and start Mailbox containers

Mailbox starts after Mailcow because `docker-compose.prod.yml` uses Mailcow's external Docker network.

```bash
cd /home/mailgo
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build
```

Wait about 2 minutes for builds. Verify:

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml ps
curl http://localhost:8082/health
```

Expected:

```json
{"service":"mailgo-api","status":"healthy"}
```

---

## Step 9 - Add DNS records

At Cloudflare DNS for `tracix.net`:

| Type | Name | Value | Proxy |
|------|------|-------|-------|
| A | `mailbox.yerevan.digital` | `your-server-IP` | Proxied |
| A | `admin-mail` | `your-server-IP` | Proxied |
| A | `mail` | `your-server-IP` | DNS only |

At the DNS zone for each hosted email domain, for example `yerevan.digital`:

| Type | Name | Value | Proxy |
|------|------|-------|-------|
| MX | `@` | `mail.tracix.net` | DNS only |

Mail protocol hostnames and MX targets must be DNS-only because Cloudflare cannot proxy SMTP, IMAP, or POP3 traffic on normal plans.

---

## Step 10 - Verify everything

```bash
# Mailbox API
curl https://mailbox.yerevan.digital/health

# Mailbox frontend
curl -I https://mailbox.yerevan.digital

# Worker logs
docker logs mailgo-worker --tail 20

# API logs
docker logs mailgo-api --tail 20
```

---

## Post-deploy checklist

- [ ] `https://mailbox.yerevan.digital` loads the Mailbox login page
- [ ] Register an account and log in
- [ ] Add `yerevan.digital` domain - status goes to "Setting up" then "DNS Required"
- [ ] Mailcow admin accessible at `https://admin-mail.tracix.net`
- [ ] Worker can provision mailboxes (check `docker logs mailgo-worker`)
- [ ] Send email from webmail - appears in Brevo logs
- [ ] Sent folder shows the message

---

## Updates

```bash
cd /home/mailgo
git pull origin main
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build
```
