# Bean Bags App

## Documentation Here: https://docs.google.com/presentation/d/1eIGK4CumX2tlkcIMx23y40eKWpdyXRJGEjptNwQUvqI/edit?slide=id.p#slide=id.p

### Next Steps:
- ~~Error handling~~
- ~~Toasts~~
- ~~Confirm Modal~~
- ~~Game persistence (in case of refresh/crash)~~
- ~~Forgot Password~~
- **Testing**
- Organizer being able to edit completed games
- Codes for teams and leagues for users to search on
- Heatmap analytics

### Things to Remember:
Don't forget to update environment variables for docker container.

- s := os.Getenv("JWT_SECRET")
- port := os.Getenv("SMTP_PORT")
- user := os.Getenv("SMTP_USER")
- pass := os.Getenv("SMTP_PASS")
- from := os.Getenv("SMTP_FROM")

## AWS Lightsail Deploy

###Infrastructure (Lightsail)
One Linux instance (Ubuntu, ~$7-12/month) can run everything
Static IP — assign one in Lightsail so your DNS doesn't break on restarts
Domain — buy from Route 53 or any registrar, point DNS to your static IP
SSL — free via Let's Encrypt + Certbot, or Lightsail has a built-in CDN/certificate option
App Architecture on the Server

Internet → Nginx (port 80/443)
              ├── /api/* → Go backend (port 8080)
              └── /* → Nuxt (port 3000)
Nginx handles SSL termination and routes requests to the right service.

Code Changes Needed
Environment variables — you'll need to set these in production:

JWT_SECRET=<strong random value>
APP_URL=https://yourdomain.com
SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM  (if you want email)
SQLite — works fine on Lightsail since the disk is persistent. Just make sure the DB path is somewhere stable like /var/app/data/beanbags.db. No migration to Postgres needed for a small app.

CORS — your Go backend likely has localhost:3000 hardcoded as the allowed origin. That needs to become your real domain.

###Deployment Steps
Create Ubuntu instance + assign static IP + open ports 80/443 in Lightsail firewall
Install Nginx, Go, Node.js on the instance
Run nuxt build locally → upload .output/ to server, run with node .output/server/index.mjs
Cross-compile Go binary for Linux: GOOS=linux GOARCH=amd64 go build -o beanbags → upload and run
Set up systemd services for both processes so they restart on crash/reboot
Configure Nginx as reverse proxy with the routing above
Run Certbot for SSL