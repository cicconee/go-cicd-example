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
Generating public/private rsa key pair. Enter file in which to save the key (/Users/USER/.ssh/id_rsa): /Users/USER/.ssh/digitalocean
Enter passphrase (empty for no passphrase):
Enter same passphrase again:
```

This will generate two files `digitalocean` and `digitalocean.pub`.

Repeat this process for the second key but use `digitalocean_deployer` for the key file name.





## DigitalOcean

Go to DigitalOcean and create a new project. Give it a meaningful name and description. Once the project is created, create a new Ubuntu 22.04 (LTS) Droplet. When prompted to choose your authentication method select `SSH Key` and choose `New SSH Key`. Give this key a name and paste the contents of `cat ~/.ssh/digitalocean.pub` for the key value. Add the SSH key and make sure it is selected before creating the droplet. **Do not set up the other key yet.**
Lastly, select the project you just created.

Once it is created the IP address should be displayed near the droplet name. In a terminal try to login to the server with the command `ssh -i /path-to-digitalocean-key root@<your-droplet-ip>`. Enter your password if you used one when creating the key.






### Systemd

In a SSH session logged in as root, configure a systemd unit file to run the application as a system service. Use your favorite text editor (I use nano) and create a new systemd file `nano /etc/systemd/system/<service-name>.service` with the following content. 

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
EnvironmentFile=/path-to-env-file

[Install]
WantedBy=multi-user.target
```

`RemainAfterExit=yes` will ensure that the programs remain running after exiting a ssh session. Besure to change `ExecStart=/path-to-binary`, `EnvironmentFile=/path-to-env-file`, and `SyslogIdentifier=app-logging-identifier` with your own system configurations.

All environment variables required by the service can be set in the file used by `EnvironmentFile`.

Now you can enable the service to start at boot time and start it.

```shell
systemctl enable <service-name>
systemctl start <service-name>
```

If you view the status you will see that it keeps trying to activate. 

```shell
systemctl status <service-name>
```

This is because the binary does not exist yet. If you continue viewing the service you will see that the time of inactivity keeps moving forward in time by 3 seconds. This is because of the configuration `Restart=always` and `RestartSec=3`. It is saying to always try and restart every 3 seconds. Stop the service so these pointless restart attempts terminate.

```shell
systemctl stop <service-name>
```




## Keep the Systemd Unit File Private

The systemd unit file should restrict who can read and write to it. It contains sensitive data in the form of environment variables and it should only allow privileged users access.

Set the permissions so that the owner and group `admin_devops` have read and write permissions. All other users should have zero access to the file.

```sh
groupadd admin_devops
chgrp admin_devops /etc/systemd/system/<service>.service
chmod 660 /etc/systemd/system/<service>.service
```







### Create Deployer User

The user `deployer` will be the user that Github Actions uses to SSH into the droplet.

Log in to the droplet as `root` and create the user with a home directory.

```sh
useradd -m deployer
```

Set the password of `deployer`. 

```sh
passwd deployer
```

Input the password as prompted.


### Give Systemctl Restart Privilege to Deployer

`deployer` will be used by Github Actions to SSH into the droplet and restart the service after copying the new binary. This is a very specific privilege `deployer` requires. A perfect reason to make use of the `sudoers` file. 

Open the `/etc/sudoers` file as `root` and add the line:

```
deployer    ALL = NOPASSWD: /usr/bin/systemctl restart <service>
```

This will allow `deployer` to run `sudo systemctl restart <service>` without needing to provide a password.




## Adding Deployer SSH Key

On the local machine that generated the key `digitalocean_deployer`, print the public key and copy the output.

```sh
cat ~/.ssh/digitalocean_deployer.pub
```

SSH into the droplet as root and then log in to `deployer`.

```sh
login deployer
```

Create the `~/.ssh` directory if it does not already exist.

```sh
mkdir -p ~/.ssh
```

Add the SSH key to an `authorized_keys` file in this directory.

```sh
nano ~/.ssh/authorized_keys
```

Paste the key to the file and save.

The `~/.ssh` directory and `authorized_keys` file must have specific restricted permissions (`700` for `~/.ssh` and `600` for `authorized_keys`). If they don't, you won't be able to log in.

```sh
chmod -R go=~/.ssh
chown -R $USER:$USER ~/.ssh
```

Now it is possible to SSH into the droplet logging in as `deployer`.

To log in as deployer, use the following command (replacing the key path with the correct path):

```sh
ssh -i /path-to-digitalocean_deployer-key deployer@<ip-address>
```


### Github Workflow


### Github Secrets


### Commit and Push




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






