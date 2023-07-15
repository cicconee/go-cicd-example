# Example GO CICD Application
An example Go application that uses github actions to deploy code commits. 

# Required
- DigitalOcean account
- Github account

# Technologies
- Go 1.19
- DigitalOcean Droplet (Ubuntu 22.04)
- Github Actions

# Steps For Deployment

Go to DigitalOcean and create a new Ubuntu 22.04 (LTS) Droplet. Make sure to pair a SSH key with the droplet. Once it is created the IP address should be displayed near the droplet name. In a terminal try to login to the server with the command `ssh root@<your-droplet-ip>`.

Once

For simplicity create a folder to hold the application binary at `/your-app-name`. Change into the new directory and clone your application.


# Systemd Service File
[Unit]
Description=Your service description

[Service]
Type=simple
Restart=always
RestartSec=3
ExecStart=/counter/go-cicd-example/app
RemainAfterExit=yes
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=gocicd

[Install]
WantedBy=multi-user.target