# Borgmox

This project takes inspiration from Proxborg.

This is a go software that runs Proxmox backup and sends them to a borgbackup repository.

## Known Issues
- Back up of Unprivileged LXCs will likely fail

## Installing

Download the latest Borgmox release to your PVE server.  
A simple wget is sufficient.
See the [Releases](https://github.com/EssGeeEich/Borgmox/releases/latest) page.

The suggested location of the executable is `/usr/local/bin/borgmox`.

## Setting up a new Backup Job

### Borg repository

If you haven't done so, it's best to set up your borg repository before this step: Borgmox will NOT create your target repository.

Here's an example taken from [Jeff Stafford's blog](https://jstaf.github.io/posts/backups-with-borg-rsync/#setup):

```
# only for Rsync.net users, use borg12 or borg14 according to your local borg version (borg -V)
export BORG_REMOTE_PATH=borg12

# see "borg init --help" for more options like storage quotas, encryption options, etc.
borg init -e repokey-blake2 username@remote.host.address:/destination/folder
```

Note down your encryption key, as you'll have to store it within the Job Configuration file.

#### SSH access to Borg repository

If your Borg repository is accessed through ssh, **you must set up public key authentication** (i.e. `ssh-keygen` locally, then paste your public key in the remote's `.ssh/authorized_keys`).

We will not guide you into these steps.

### VM/LXC Pool
In order to select VMs or LXCs to back up, you will have to set up a PVE "Pool" with all the VMs or LXCs that you want to back up.

First, create your PVE Pool: `Datacenter`->`Pools`->`Create`  

Then, select your newly created pool from the Server View (in the left panel), open the `Members` section, and `Add` your LXCs and VMs.

### Setting up the Backup Job file

Copy and rename the file `scripts/sample.toml` to a folder of your choice.  
The suggested location of this file would be within `/etc/borgmox/conf.d/`.

Open the file with your favourite editor and fill the fields with whatever looks appropriate to you.

You can find out more about these options at the end of this README.

## Running the configured Jobs

### Running both the Backup and Prune Job

From your preferred shell, run the following command (as root):

`borgmox /etc/borgmox/conf.d/*.toml`

### Running ONLY the Backup Job

From your preferred shell, run the following command (as root):

`borgmox --no-prune /etc/borgmox/conf.d/*.toml`

### Running ONLY the Prune Job

From your preferred shell, run the following command (as root):

`borgmox --no-backup /etc/borgmox/conf.d/*.toml`

## Setting up a systemd service

Sample `borgmox.service` and `borgmox.timer` files have been provided in the `scripts/` folder.

These two files are configured to run Borgmox on a daily basis at 3AM, with the paths and filenames suggested before.

You can download them and tweak them (if needed).

To enable the systemd service, put the two files within `/etc/systemd/system/` and run:

```
systemctl daemon-reload
systemctl enable borgmox.timer
```

## Restoring from a Backup

To restore a VM/LXC from a Backup, currently the best way is to use the shell, using borg directly:

```
# Only for Rsync.net users, use borg12 or borg14 according to your local borg version (borg -V=
export BORG_REMOTE_PATH=borg12

# Set the Repository and Passphrase based on what's already in your config
export BORG_REPO=ssh://your.borg.repo
export BORG_PASSPHRASE='your.borg!passphrase'

# List the available backups
borg list | grep your_vmid

# Start the restore process (VM Only)
borg extract ::your-backup-file_date_hour.vma --stdout | qmrestore - (new_vmid)

# Start the restore process (LXC Only)
borg extract ::your-backup-file_date_hour.tar --stdout | pct restore (new_vmid) --rootfs (your_new_rootfs) -
```

TODO: Document what your_new_rootfs should look like...!

# Configuration file

A configuration file, first of all, contains a list of "jobs":

```toml
[BackupJobs]
[BackupJobs.'My Job']
...
[BackupJobs.'My Second Job']
...
```

## Sparse settings
First of all, let's look at the few job settings that don't belong to a sub-group:

```toml
[BackupJobs.'My Job']
ArchivePrefix = ''
VmPool = ['my_proxmox_vm_pool', 'my_proxmox_lxc_pool']
VmMode = 'image'
LxcMode = 'image'
```

### ArchivePrefix
This setting holds the Prefix of the Archive Name.  
If left empty, this will take the node's hostname (kind of how PVE already does when backing up with vzdump).

### VmPool
A list of PVE pools that will be read and deduplicated.  
The backup/prune operations will run on all of the VMs/LXCs listed in these pools.

### VmMode
Backup mode for VMs.  
This is reserved for future use. Can only be `image`.

### LxcMode
Backup mode for LXCs.  
This is reserved for future use. Can only be `image`.

## Notification settings
Borgmox uses [ntfy](https://ntfy.sh/) to send backup/prune job notifications.  
It is not a critical dependency, and you can disable notifications altogether.

The Notification Settings are stored within the "Notification" group.

```toml
[BackupJobs.'My Job'.Notification]
TargetServer = ''
AuthUser = 'my_user_or_empty_for_access_token'
AuthPassword = 'my_user_password_or_token'
Topic = 'MyNotificationTopic'
```

### TargetServer
The target server should point to the ntfy instance of preference.  
Leave empty to disable notifications.  
Should include `https://` or `http://`.

### AuthUser
The ntfy user.  
Should be left empty if the Topic is not protected, or if you want to use the Access Token.

### AuthPassword
The ntfy password, or Access Token.  
If this is an Access Token, you should empty AuthUser.

### Topic
The ntfy topic that we'll publish the notifications to.

## Backup and Prune Job Notification Settings
The Backup and Prune Jobs will have some Notification Settings of their own:

```toml
[BackupJobs.'My Job'.Notification.BackupTargetInfo]
Frequency = 'single job'
SuccessPriority = 'default'
FailurePriority = 'high'
SuccessEmailTarget = ''
FailureEmailTarget = ''

[BackupJobs.'My Job'.Notification.PruneTargetInfo]
Frequency = 'every vm'
SuccessPriority = 'default'
FailurePriority = 'high'
SuccessEmailTarget = ''
FailureEmailTarget = ''
```

### Frequency
Frequency should be one of the following values:
- `never`:  
  Never sends a notification.
- `every vm`:  
  Sends a notification right after a VM/LXC Backup/Prune has finished.
- `single job`:  
  Sends a notification after the entire Backup/Prune Job has finished.

### SuccessPriority and FailurePriority
Notification Priority in case of Success or Failure.
See [ntfy Message priority](https://docs.ntfy.sh/publish/#message-priority) for additional informations.

Should be one of the following:

- max
- urgent
- high
- default
- low
- min
- off

### SuccessEmailTarget and FailureEmailTarget
A single email address that will receive a notification in case of Success or Failure.

See [ntfy E-mail notifications](https://docs.ntfy.sh/publish/#e-mail-notifications) for additional informations.

## Borg Repository Settings
Finally, there are the Borg repository settings:

```toml
[BackupJobs.'My Job'.Borg]
Repository = 'ssh://my_borg_repo'
RemotePath = '/my/remote/borg/path/if/needed/or/empty'
Passphrase = 'my-borg-passphrase'
```

### Repository
Path to the Borg repository.

If your Borg repository is accessed through ssh, **you must set up public key authentication** (i.e. `ssh-keygen` locally, then paste your public key in the remote's `.ssh/authorized_keys`).

We will not guide you into these steps.

### RemotePath
The Remote Path indicates the path for the remote borg executable.  
Only required if the Repository is stored on a system with multiple borg installations.

Example:  
For rsync.net, you should set this to `borg12` or `borg14` depending on which version of borg you have installed.

### Passphrase
The passphrase for the Repository.  
This is the one you typed in the "Setting up a new Backup Job" step.

## Borg Prune Settings
We will rely on the `borg prune` and `borg compact` commands to erase old backups.

```toml
[BackupJobs.'My Job'.Borg.Prune]
Enabled = false
Compact = true
KeepWithin = '15D'
KeepLast = 10
KeepMinutely = 0
KeepHourly = 0
KeepDaily = 0
KeepWeekly = 8
KeepMonthly = 12
KeepYearly = 10
```

See [borg prune](https://borgbackup.readthedocs.io/en/stable/usage/prune.html) for additional informations.
