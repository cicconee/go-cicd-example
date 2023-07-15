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

Go to DigitalOcean and create a new Ubuntu 22.04 (LTS) Droplet. Make sure to configure a SSH key with the droplet. Once it is created the IP address should be displayed near the droplet name. In a terminal try to login to the server with the command `ssh root@<your-droplet-ip>`. Enter your password if you used one when creating the key.

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

Any environment variables required by the service can be set using the syntax `Environment=ENV_VAR=value`.

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