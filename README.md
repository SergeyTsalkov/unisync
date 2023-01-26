Unisync -- The Programmer's Sync Tool
====
Unisync is a command-line tool for **continuous 2-way sync between a folder on your computer, and a folder on a remote server**.
It's meant for using a local text editor to do development on your remote VPS. Unisync generally runs over SSH, but it doesn't have to.

**To get started, see the documentation on our website: https://unisync.sh**

### How to Use
1. Download unisync to local and remote side. The local copy should go somewhere in the system PATH (such as `/usr/local/bin`). The remote copy should go to a directory in the system PATH, or else in the `~/.unisync` folder (if you don't have root access).

1. Create a unisync config file on the local side.

    ```
    mkdir -p ~/.unisync && cd ~/.unisync
    nano remoteserver.conf
    ```

    The format of the config file is a simple "key = value" structure, one per line. Put this into the file:

    ```
    local = /path/to/local/folder
    remote = /path/to/remote/folder
    user = [username]
    host = [ip]
    ```

    The local and remote folders are the ones that will be synced. One of them should be empty the first time you sync, so its contents can be filled in from the other one. The user and host are your SSH connection information for the remote server.

1. Run unisync locally with the name of the config file you created in the last step (remoteserver in this example). It will use SSH to connect to the remote copy of unisync, and they'll sync the specified folders.
    ```
    unisync remoteserver
    ```

### Lots more features!
You should read our documentation at https://unisync.sh to learn more about configuring unisync.
