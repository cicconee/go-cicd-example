# Example GO CICD Application
An example Go application that uses github actions to deploy code commits. 

## Required
- DigitalOcean account
- Github account

## Technologies
- Go 1.19
- DigitalOcean Droplet (Ubuntu 22.04)
- Github Actions

## The Production Server

This section documents how to make the server production ready and configuring CI/CD pipeline. The server will be hosted on a `DigitalOcean Droplet`. The CI/CD pipeline uses `Github actions`.

### Generating SSH Keys

The server will require the use of two seperate SSH keys. 

The first key will be for a user that belongs to the `sudo` group. This account will have elevated privileges.

The second key will be for the CI/CD pipeline user. This account will have limited access.

Generate the `sudo` user key. It is highly recommended to password protect your SSH keys.

```sh
ssh-keygen
```

You will be prompted to name the key. I like to name my keys by service (and role if it applies).

```
Generating public/private rsa key pair. Enter file in which to save the key (/Users/USER/.ssh/id_rsa): /Users/USER/.ssh/digitalocean
Enter passphrase (empty for no passphrase):
Enter same passphrase again:
```

This will generate two files `digitalocean` and `digitalocean.pub`.

Repeat this process for the pipeline key, but use `digitalocean_deployer` for the key file name.

### DigitalOcean

Go to DigitalOcean and create a new project. Give it a meaningful name and description. Once the project is created, create a new Ubuntu 22.04 (LTS) Droplet. When prompted to choose your authentication method select `SSH Key` and choose `New SSH Key`. Give this key a name and paste the contents of `~/.ssh/digitalocean.pub`.

```sh
cat ~/.ssh/digitalocean.pub
``` 

Add the SSH key and make sure it is selected before creating the droplet. **Do not set up the other key yet.**

Lastly, select the project you just created.

Once it is created the IP address should be displayed near the droplet name. In a terminal try to login to the server.

```sh
ssh -i ~/.ssh/digitalocean root@<ip-address>
```

### Create the Sudo User

Log in as `root` and add the user.

```sh
adduser <username>
```

You will be prompted to create and verify a password for the user.

```
New password:
Retype new password:
```

Next, you will be asked to fill in some information about the user. You may leave these blank.

```
Enter the new values, or press ENTER for default
    Full Name []:
    Room Number []:
    Work Phone []:
    Home Phone []:
    Other [];
Is the information correct? [Y/n]
```

Now add the user to the `sudo` group.

```sh
usermod -aG sudo <username>
```

Switch to the new user and test a `sudo` command.

```sh
su - <username>
sudo ls -la /root
```

### Create The Pipeline User

The pipeline user will deploy and run the code. 

Log in as the `sudo` user and create the user with a home directory.

```sh
sudo useradd -m <username>
```

Then set the password.

```sh
sudo passwd <username>
```

You will be prompted to create and verify a password for the user.

```
New password:
Retype new password:
```

Now try switching to the pipeline user.

```sh
su - <username>
```


### Install Postgresql

To install `postgresql` log in as the `sudo` user and enable the `postgresql` official package repository.

```sh
sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
sudo wget -qO- https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo tee /etc/apt/trusted.gpg.d/pgdg.asc &>/dev/null
```

Update the packages.

```sh
sudo apt update
```

Install the `postgresql` package along with a `postgresql-client` package which adds some useful utilities and functionality.

```sh
sudo apt install postgresql postgresql-client -y
```

Ensure the service is running and check the status (you should see a `active` status).

```sh
systemctl start postgresql.service
systemctl status postgresql.service
```

Execute the `psql` command as the `postgres` user to enter the `postgresql` console.

_Note that installing `postgresql` will create a `postgres` user on the server._

```sh
sudo -u postgres psql
```

You successfully installed `postgresql`. Type `\q` to exit the `postgresql` console.



### Configure Postgresql for the Application

Log in as the `sudo` user and create a `postgresql role` for the pipeline user.

```sh
sudo -u postgres createuser --interactive
```

Make the `role` name the same as the pipeline user, and enter `n` for all prompts. 

_Note that `postgresql` by default will allow user accounts to log in to a `role` if the names match. So the `name` of the `role` should match the `name` of the `linux` user account that will use it. It is possible to login as another `role` using the command `sudo -u <role-name> psql`, but for simplicity they will remain the same._

Create the application database and make the name the same as the `role` and pipeline user. 

```sh
sudo -u postgres createdb <database-name>
```

_Just like the linking of user accounts and roles, by default, the `psql` command will log in a user and connect to the database with the same name as the role. This can be overriden with `psql -d DATABASE_NAME`. For simplicity, make the database name the same as the `role` name._

Enter the console as the `postgres` role.

```sh
sudo -u postgres psql
```

Connect to the database, grant the pipeline role privileges, and set the password.

```sql
\c <database-name>

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO <pipeline-role>;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO <pipeline-role>;

GRANT CREATE ON SCHEMA public TO <pipeline-role>;

ALTER ROLE <pipeline-role> WITH PASSWORD 'password';
```

TODO!!!
These privileges allow the pipeline role to create, drop, and alter tables, select, insert, update, and delete rows, and create and use sequences. This limitation adds some security to the configuration as the pipeline `role` does not have `superuser` privileges.

### Verify the Application Role Privileges

Log in as the `sudo` user and enter the `postgresql` console as the pipeline user. 

```sh 
sudo -u <pipeline-username> psql
```

Remember, by default, `postgresql` will look for a `role` that matches the user account. Since the `sudo` user is executing `psql` on behalf of the pipeline user, `postgresql` will look for a role identical to the pipeline username. Since they match, it connects successfully.

Also note, by default `postgresql` will attempt to connect to a database that matches the `role` name if no database is specified. For example, if the `role` name is `app`, it will look for the database `app`.

To override this, simply specify the database name.

```sh
sudo -u <username> psql -d <database-name>
```

Create a table, insert some data, select the data, and then drop the table.

```sql
CREATE TABLE test_table(
    id INTEGER,
    name VARCHAR(255)
);

INSERT INTO test_table(id, name) VALUES(10, "foo");

SELECT * FROM test_table;

DROP TABLE test_table;
```

Everything should succeed.

Now connect to the `postgres` database.

```sh
sudo -u <pipeline-username> psql -d postgres
```

Or if you are still in the `postgresql` console use: `\c postgres`

Try running the commands again. You should get `ERROR: permission denied for schema public`. 


### Installing Migrate

The database migration utility will be [golang-migrate](https://github.com/golang-migrate/migrate/tree/master).

Log in as the `sudo` user and install the repository.

```sh
sudo curl -s https://packagecloud.io/install/repositories/golang-migrate/migrate/script.deb.sh | sudo bash
```

Then install the package.

```sh
sudo apt-get install migrate=4.16.2
```

This will install the `migrate` command. 

### Running Migrations

The main options to use are `-path` and `-database`.

`-path` is the directory path that contains the migration files.

`-database` is the database driver and url, formatted as `driver://url`.

The commonly used commands are `up` on `down`. 

`up` will apply all or N up migrations.

`down` will apply all or N down migrations. This should almost never be used on the production server.

#### Example Up Migration

`migrate` will look for a directory named `migrations` located in the directory the command was executed from. 

Then apply all the up migrations to the database `table_name`.

```sh
migrate -path migrations -database postgres://user:password@localhost:5432/table_name up
```

### Systemd

Log in as the `sudo` user and configure a systemd unit file to run the application as a system service. Use your favorite text editor (I use nano) and create a new systemd file. 

```sh
sudo nano /etc/systemd/system/<service-name>.service
``` 

Paste the following content.

```
[Unit]
Description=Your service description

[Service]
Type=simple
Restart=always
RestartSec=3
ExecStart=<PATH-TO-BINARY>
RemainAfterExit=yes
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=<APP-IDENTIFIER>
EnvironmentFile=<PATH-TO-ENV>

[Install]
WantedBy=multi-user.target
```

Besure to change `ExecStart`, `EnvironmentFile`, and `SyslogIdentifier` with your own system configurations.

All environment variables required by the service will be set in the file used by `EnvironmentFile`. This file will be generated by the pipeline via Github Actions.

Now you can enable the service to start at boot time.

```shell
sudo systemctl enable <service-name>
```


### Give Systemctl Restart Privilege to the Pipeline User

The pipeline user will be used by Github Actions to SSH into the droplet and restart the service after copying the new binary to the server. 

This is a very specific privilege required by the pipeline user. A perfect reason to make use of the `sudoers` file. 

Log in as the `sudo` user and open the `/etc/sudoers` file.

```sh
sudo nano /etc/sudoers
```

Add the line in the `User privilege specification` section below the `root` definition:

```
<pipeline-username>    ALL = NOPASSWD: /usr/bin/systemctl restart <service-name>
```

This will allow the pipeline user to run `sudo systemctl restart <service-name>` without needing to provide a password.

### Adding Deployer SSH Key

On the local machine that generated the key `digitalocean_deployer`, print the public key and copy the output.

```sh
cat ~/.ssh/digitalocean_deployer.pub
```

Log in to the droplet as the `sudo` user and create the `.ssh` directory in the home directory of the pipeline user.

```sh
sudo mkdir -p /home/<pipeline-username>/.ssh
```

Add the SSH key to an `authorized_keys` file in this directory.

```sh
sudo echo <digitalocean_deployer_key> >> /home/<pipeline-username>/.ssh/authorized_keys
```

The `/home/<pipeline-username>/.ssh` directory and `/home/<pipeline-username>/.ssh/authorized_keys` file must have specific restricted permissions (`700` for `.ssh` and `600` for `authorized_keys`). If they don't, you won't be able to log in.

```sh
chmod -R go= /home/<pipeline-username>/.ssh
chown -R <pipeline-username>:<pipeline-username> /home/<pipeline-username>/.ssh
```

Now it is possible to SSH into the droplet logging in as `app`.

```sh
ssh -i ~/.ssh/digitalocean_deployer <pipeline-username>@<ip-address>
```









## CI/CD Pipeline

The CI/CD pipeline uses `Github Actions` to build, test, and deploy the application to the server. 

The following secrets are needed to deploy this application.

| Secret | Description |
| ------ | ----------- |
| HOST | The server IP address  adf asdf asdf adf adsf adf asdf asdf asdf asdf asdf asdf asdf asdf asfd as fasdfasf asdf |
| KEY | The SSH key |
| PASSPHRASE | The password for the SSH key |
| USERNAME | The user that will SSH into the server |

Below are the secrets that are application specific. These are the secrets that will be written to the `.env` file and then used by the application.

| Secret | Description |
| ------ | ----------- |
| MY_NAME | A line to be printed |

TODO: DB_NAME, DB_USER, DB_PASSWORD, DB_PORT

### Using Environment Variables in Production

Environment variables will be set by using `Github Secrets`. The `deploy` workflow will grab the required secrets, copy them to a `.env` file, and then eventually, copy the `.env` file to the server.

The `Create .env file` step of the `deploy` workflow will create a `.env` file and then append secrets to the file.

```yaml
- name: Create .env file
  run: |
    touch .env
    echo ENV_VAR=${{ secrets.ENV_VAR }} >> .env
```

**As your application requires more environment variables, you will need to create a secret and then write the secret to the file.**

### Adding Environment Variable for Production

Navigate to the projects `settings` page and choose `Actions` under `Secrets and variables`. Select `New repository secret` and enter the name and value.

In the `Create .env file` step of the `deploy` workflow, add a new line in the `run` section that will write the secret to the `.env` file.

```sh
echo NEW_ENV_VAR=${{ secrets.NEW_ENV_VAR }} >> .env
```

### Copying Resources to Production Server

The `Publish to server` step will do a secure copy of the compiled application and `.env` file.

```yaml
- name: Publish to server
  uses: appleboy/scp-action@v0.1.4
  with:
    host: ${{ secrets.HOST }}
    username: ${{ secrets.USERNAME }}
    key: ${{ secrets.KEY }}
    passphrase: ${{ secrets.PASSPHRASE }}
    source: "app,.env"
    target: /home/app
```

The target path of `app` and `.env` should match the values of the systemd `ExecStart` and `EnvironmentFile`. 

For example, `app` is being copied to `/home/app` so its target path is `/home/app/app`.

**As the application requires more resources, you will need to append the files/directories to `source`**

### Running the Service

The `Start the service` step will SSH into the server. It will change the permissions of `.env` to allow read and writes by the owner and group. The `USERNAME` secret is the owner and group. 

Then the systemd service `go-cicd` will be restarted.

```yaml
- name: Start the service
  uses: appleboy/ssh-action@v0.1.10
  with:
    host: ${{ secrets.HOST }}
    username: ${{ secrets.USERNAME }}
    key: ${{ secrets.KEY }}
    passphrase: ${{ secrets.PASSPHRASE }}
    script: |
    chmod 660 /home/app/.env
    sudo systemctl restart go-cicd
```







## Local Development

### Postgres Docker Container

In `docker` log in as the `superuser` role.

```sh
psql -d weather_app_db -U weather_app
```

Create the application `role` to mock the production environment and set the appropriate privileges.

```sql
CREATE ROLE app WITH PASSWORD 'password';

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO app;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO app;

GRANT CREATE ON SCHEMA public TO app;

ALTER ROLE app WITH LOGIN;
```

If any tables were already created before creating the new role, execute the following.

```sql
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO app;

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO app;
```