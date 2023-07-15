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

Go to DigitalOcean and create a new Ubuntu 22.04 (LTS) Droplet. Make sure to pair a SSH key with the droplet. Once it is created the IP address should be displayed near the droplet name. In a terminal try to login to the server with the command `ssh root@<your-droplet-ip>`. Enter your password if you used one when creating the key.

In the ssh session you need to configure a systemd unit file to run the application as a system service. Use your favorite text editor (I use nano) and create a new systemd file `nano /etc/systemd/system/<app-name>.service` with the following content. 

```
[Unit]
Description=Your service description

[Service]
Type=simple
Restart=always
RestartSec=3
ExecStart=/path-to-binary
RemainAfterExit=yes
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=app-logging-identifier

[Install]
WantedBy=multi-user.target
```

`RemainAfterExit=yes` will ensure that the programs remain running after exiting a ssh session. Besure to change `ExecStart=/path-to-binary` and `SyslogIdentifier=app-logging-identifier` with your own system configurations.

Now you can enable the service.

```shell
# this will enable the service to start at server bootup
systemctl enable <app-name>

# start the service (Executes ExecStart prepending any environment variables)
systemctl start <app-name>
```

If you view the status you will see that 