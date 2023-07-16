# Example GO CICD Application
An example Go application that uses github actions to deploy code commits. 

# Required
- DigitalOcean account
- Github account

# Technologies
- Go 1.19
- DigitalOcean Droplet (Ubuntu 22.04)
- Github Actions

# Configuring CI/CD

This section documents the steps needed to configure the CI/CD pipeline using Github Actions and DigitalOcean Droplets. 

## Generating SSH Keys

Configuring the CI/CD pipeline will require the use of two seperate SSH keys. The first key will be for the root user. This key is very important as gaining access to this key will give root access to the server. The second key is for the Github Actions deployer. This key will be for a user with limited access.

Generate the root key. It is highly recommended to password protect your SSH keys.

```sh
ssh-keygen
```

You will be prompted to save and name the key. I like to name my keys by service (and role if it applies).

```
Generating public/private rsa key pair. Enter file in which to save the key (/Users/USER/.ssh/id_rsa): ~/.ssh/digitalocean_root
Enter passphrase (empty for no passphrase):
Enter same passphrase again:
```

This will generate two files `digitalocean_root` and `digitalocean_root.pub`.

Repeat this process for the second key but use `digitalocean_deployer` for the key file name.

## DigitalOcean

Go to DigitalOcean and create a new project. Give it a meaningful name and description. Once the project is created, create a new Ubuntu 22.04 (LTS) Droplet. When prompted to choose your authentication method select `SSH Key` and choose `New SSH Key`. Give this key a name and paste the contents of `cat ~/.ssh/digitalocean_root` for the key value. Add the SSH key and make sure it is selected before creating the droplet. **Do not set up the other key yet.**
Lastly, select the new project you just created.

Once it is created the IP address should be displayed near the droplet name. In a terminal try to login to the server with the command `ssh root@<your-droplet-ip>`. Enter your password if you used one when creating the key.

### Systemd

In the ssh session you need to configure a systemd unit file to run the application as a system service. Use your favorite text editor (I use nano) and create a new systemd file `nano /etc/systemd/system/<service-name>.service` with the following content. 

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
Environment=ENV_ONE=one
Environment=ENV_TWO=two

[Install]
WantedBy=multi-user.target
```

`RemainAfterExit=yes` will ensure that the programs remain running after exiting a ssh session. Besure to change `ExecStart=/path-to-binary` and `SyslogIdentifier=app-logging-identifier` with your own system configurations.

All environment variables required by the service can be set using the syntax `Environment=KEY=value`.

Now you can enable the service.

```shell
# this will enable the service to start at server bootup
systemctl enable <service-name>

# start the service (Executes ExecStart prepending any environment variables)
systemctl start <service-name>
```

If you view the status you will see that it is `inactive`. 

```shell
systemctl status <service-name>
```

This is because the binary does not exist yet. If you continue viewing the service you will see that the time of inactivity keeps moving forward in time by 3 seconds. This is because of the configuration `Restart=always` and `RestartSec=3`. It is saying to always try and restart every 3 seconds. Stop the service so these pointless restart attempts terminate.

```shell
systemctl stop <service-name>
```

# Viewing Service Logs

It is very important to be able to view log files. Typically newer developers are used to viewing them in the console they ran the program, however, now the application is being ran as a system service. By default, systemd services will write to `syslog`. The `SyslogIdentifier=app-logging-identifier` configuration tells `systemd` to tag each log entry with `app-logging-identifier`.

Make sure the service is running before viewing the log journal.

```shell
journalctl -u example
```

_Hint: append a `-f` to the end of the command to get a continous stream of logs_.

If everything is configured correctly you should see the logs. 

If you are like me and got the `No journal files were found` message, something is configured incorrectly. I found that the Ubuntu 22.04 (LTS) droplet comes with a corrupted journal service (or I am overlooking a configuration setting). First verify that the service is at least writing to the `syslog`.

```shell
cat /var/log/syslog
``` 

Check if there are any log entries from your service (look for your `SyslogIdentifier` value). If you do not see any logs, then either your service is not running or you configured something incorrectly in the previous steps.

Now that you verified logs are at least being written to `syslog`, restart the `systemd-journald` service.

```shell
systemctl restart systemd-journald.service
```

View your service logs again, and you should see the service logs. 


## Create group

groupadd devops

## Change group owner of a file

chgrp devops env-file

## Set Env File Permissions

chmod 660 env-file

This will allow read and writes only by the owner and group. All others will have zero access to the file.