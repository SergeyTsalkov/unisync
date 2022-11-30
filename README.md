Unisync -- The Programmer's Sync Tool
====
Unisync is a command-line tool for **continuous 2-way sync between a folder on your computer, and a folder on a remote server**.
It's meant for using a local text editor to do development on your remote VPS. Unisync generally runs over SSH, but it doesn't have to.

For the initial sync, it will prefer whichever side has the "newer" version of each file (by date modified).
This behavior can be changed in the config file.

After the initial sync, it will keep a local cache file that will let it correctly sync all changes even
if one or both sides have a wrong clock.

While connected, it will continuously sync changes in both directions. If you disconnect and then later reconnect, 
it will "catch up" all of the changes that happened, and then go back to watching for new changes.

If your client is running Windows and doesn't have SSH installed, unisync will run its own built-in SSH client. 
It will even check for Pageant SSH keys, since many Windows users also use PuTTY/Pageant.

If your remote server doesn't run SSH, you should use the secure direct connect mode instead. 
It runs over TLS and is safe and encrypted. However, we recommend you run over SSH if you can.



## Getting Started:

Let's assume you just created a new Ubuntu VPS at [ip] and you want to sync a local project folder to it.

### Generate a local SSH key, create a remote user, and set up that user for SSH connections
Skip these steps if you already have a remote ssh user set up, and your local ssh key is installed for the user.

1. Create a local SSH key.
```
mkdir -p ~/.ssh
ssh-keygen -t rsa -f "$HOME/.ssh/id_rsa" -N ""
```

2. Create your remote user account on the VPS, then remotely install the SSH key.
```
ssh root@[ip] 'adduser --disabled-password --gecos "" [username]'
scp 
```


### Install unisync remotely and locally

1. Install unisync on the remote server you'll be connecting to.

```
ssh [username]@[ip] 'mkdir -p ~/.unisync && cd ~/.unisync && curl unisync.sh/download'
```

2. Install unisync locally, and put it into a directory in your PATH.

```
# add ~/bin to your PATH
cd ~/bin && curl unisync.sh/download
```

3. Create a unisync config file. Your config file will only exist locally-- the remote server doesn't need one.

```
mkdir -p ~/.unisync && cd ~/.unisync
nano remoteserver.conf
```

The format of the config file is a simple "key = value" structure, one per line. Put this into the file:
```
local = /path/to/local/folder
remote = /path/to/remote/folder
user = [usernmae]
host = [ip]
```

The local and remote folders are the ones that will be synced. One of them should be empty the first time you sync, so its contents can be filled in from the other one. The user and host are your SSH connection information for the remote server.

### Run unisync!
4. Run unisync locally with the name of the config file you created in the last step (remoteserver in this example). It will use SSH to connect to the remote copy of unisync, and they'll sync the specified folders.
```
unisync remoteserver
```

Assuming that either local or remote was empty to begin with, unisync will use the non-empty side to fill it. It will then watch both sides for changes and sync those changes.
