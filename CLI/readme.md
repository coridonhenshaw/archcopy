# Archcopy

Archcopy is a clean-slate file copy utility for Linux. It is designed to deliver predictable results and reasonable performance under most use cases without manual tuning. Archcopy supports both local filesystems and network-based transfers over TCP or through an SSH tunnel. Network transfers are cryptographically signed and optional mutual TLS is supported for greater security.

Archcopy is experimental software and should not be used for anything important. Also see the license terms, below, which prohibit commercial use.

## Performance:

In testing with an 11 GB virtual machine disk image, Archcopy performance is comparable or better to that of `rsync`, without the need to tune options to avoid pathological cases. Archcopy internal readback verification is also considerably faster than manually checking the integrity of an rsync copied file by using shasum against both the source and destination files.

### 100 Mbps LAN

| Duration | Tool + options |
| - | - |
8m18s   | Archcopy over SSH | 
8m20s   | Archcopy over TCP | 
11m53   | rsync over SSH, with `-z -zc=zstd` compression | 
16m43  | rsync over SSH | 
19m27s  | rsync over NFS | 

### 1000 Mbps LAN

| Duration | Tool + options |
| - | - |
1m6s | Archcopy TCP  |
1m39s | Rsync SSH
1m45s | Archcopy SSH
1m50s | Rsync NFS
1m51s | Archcopy NFS
2m30s | Archcopy TCP with verification  |
10m23 | Rsync SSH, `-z -zc=zstd` |

### Local filesystem, 7200 RPM HDD

| Duration | Tool + options |
| - | - |
3m14s |  Archcopy |
3m54s  | Rsync  |
4m47s |  Archcopy with verification |
7m14s | Rsync, with manual verification using shasum to hash both the source and destination files (3m20s consumed by shasum).

## Sample Output

    >archcopy -if ~/qemu/test.qcow2 -od /mnt/WD2003FZEX/Workarea -ro ssh://root@nas-srv -f
    Archcopy

    R: Archcopy
    R: JSON: {"PSK":"hVJ2SE0FVIneJQfz19X1wQf3","CACert":null}
    R: Slave mode. Waiting for connection on unix:///tmp/archcopy-pV1UTM6CcBkRIYxG-iJFZOw35iZxahmL. TLS is disabled.
    Transferring 1 file comprising 11 GB.
    
    R: Client-id kSK_2W49YCv076HioL6g4dxfI_vmVjuQ9yvivuCUpFY authorized for session QnB4UrB47ZPWCkkRB0Z7qnyIUQShI-VF2d3vsQwD1IU
    R: WriteFile (kSK_2W49YCv076HioL6g4dxfI_vmVjuQ9yvivuCUpFY): /mnt/WD2003FZEX/Workarea/test.qcow2
    R: Finalize: 8m17.742s 11 MB/s 23 MB/s 0.493  /mnt/WD2003FZEX/Workarea/test.qcow2
    Done 11 GB (8m18s 23 MB/s) /home/chenshaw/qemu/test.qcow2 => /mnt/WD2003FZEX/Workarea/test.qcow2
    
    Success.
    
    R: Ended session QnB4UrB47ZPWCkkRB0Z7qnyIUQShI-VF2d3vsQwD1IU for client-id kSK_2W49YCv076HioL6g4dxfI_vmVjuQ9yvivuCUpFY

## Usage

| Short Parameter | Long Parameter | Description |
| :---: | :---: | - |
-if | --inputfile           |  Source file. Must be used with `-od` if repeated.
-ic  | --inputconventional  |  Treat all command line parameters after -- as input filenames. An output directory must be specified with -od.
-id | --inputdirectory      | Source directory. Contents will be recursively copied. Must be used with `-od`.
-of | --outputfile          | Destination filename. Only valid if `-if` is used once and only once.
-od  | --outputdirectory    | Destination directory.
-p | --preserverelativepath | Create source relative path in the destination directory. If this parameter is not specified, any path in the source will be removed when the destination filename is built. E.g. *without* `-p`, `-if /home/foo/srcfile -od /home/bar` will result in the creation of `/home/bar/srcfile`. If `-p` is specified, `/home/bar/home/foo/srcfile` will be created.
-f | --force                | Overwrite existing destination files.
-r | --resume               | Resume an interrupted transfer.
-s | --sparse               | Write 4K blocks containing only zero bytes as sparse extents.
-v | --verify               | Verify that the Blake2b hashes of the source and destination files are identical after copy. This option will increase copy times as the destination file must be read back from disk to confirm its hash. The hash of the source file is calculated during the copy operation.
-c | --continue             | Continue with other files in the event of errors. Default behavior is to exit on any error.
-d | --dryrun               | Dry run only. List files to be transferred, but do not transfer them.
-ro | --remoteoutputurl     | RPC: URL for remote output. Output files will be written to the paths provided (by `-of` or `-od`) on the system at the specified URL. See URL syntax, below.
-ri  |--remoteinputurl      | RPC: URL for remote input. Input files will be read from the paths provided (by `-if` or `-id`, etc) on the system at the specified URL. See URL syntax, below.
-rp  |--rpcpsk              | RPC: Pre-shared key. Required when interfacing with remote systems.
-rk  | --rpcsshkey          | RPC: Override SSH key (default: ~/.ssh/id_rsa)
| |                         | SSH key passphrases can be specified with the `archcopysshpassphrase` environment variable. SSH password-based authentication is not supported.
-ca  | --rpccacert          | RPC: CA certificate.
-cc | --rpccert             | RPC: Client certificate.
-ck | --rpccertkey          | RPC: Client certificate key.
||| All three of `-ca`, `-cc` and `-ck` must be provided to enable TLS. Use the GenerateCerts subcommand to create these files.

### GenerateCerts subcommand

This subcommand accepts no parameters. A certificate authority public key, and client/server certificate/key pairs will be written to the current directory. Pre-existing files will be overwritten.

Note: To use externally-generated TLS credentials with Archcopy, the server name field in the certificate must be set to 'Archcopy'.

### Slave subcommand

| Short Parameter | Long Parameter | Description |
| :---: | :---: | - |
| -sl | --listenurl      | URL to listen for connections. See URL syntax, below.
| -sa | --auto           | Generate a PSK automatically.
| -st | --singlesession  | Terminate after the client disconnects.
| -sc | --certificate    | TLS certificate file. Use the GenerateCerts subcommand to create this file.
| -sk | --certificatekey | TLS certificate private key file. Use the GenerateCerts subcommand to create this file.
| | | Note: to enable TLS a CA certificate must also be provided using `--rpccacert`.

### URL Syntax

| Structure | Example | Description
| --- | --- | --- |
| tcp://\<address>:\<port> | tcp://:19683               | Listen on port 19683/tcp on all interfaces.  Both IPv4 and IPv6 are supported.
|                          | tcp://192.168.0.128:19683  | Listen on port 19683/tcp on 192.168.0.128, or connect to port 19683/tcp on 192.168.0.128.
|                          | tcp://hostsystem:19683     | Connect to `hostsystem` port 19683.
| ssh://\<username>@\<address> | ssh://root@hostsystem  | Connect to `hostsystem` as `root`. Public keys are the only supported SSH authentication method.
| unix:///\<absolutepath> | unix:///tmp/archcopy-tunnel | Listen or connect to a Unix domain socket at `/tmp/archcopy-tunnel`. Unix socket support is present to implement SSH tunneling (e.g. with ssh:// URLs) and may not have other practical uses cases.

## Notes

Archcopy must be installed on the remote system (under the name `archcopy`, in the system path and with +x set) for ssh:// URLs to work.

Do not expose a slave (listening) instance of Archcopy to the Internet. Archcopy has not been sufficiently tested to be suitable for exposure to hostile actors. To transit Archcopy traffic over the Internet, use a well-known VPN (e.g. Wireguard) or use a well-known gRPC proxy (e.g. Nginx) to implement authentication and encryption.

Portability to operating systems other than Linux is unknown. Permission and ownership transfer will not work on Windows due to underlying differences between Windows and the POSIX security model. Sparse file support on Windows will not work without additional code to access Windows sparse file APIs.

## License

Copyright 2022 Coridon Henshaw

Permission is granted to all natural persons to execute, distribute, and/or modify this software (including its documentation) subject to the following terms:

1. Subject to point \#2, below, **all commercial use and distribution is prohibited.** This software has been released for personal and academic use for the betterment of society through any purpose that does not create income or revenue. *It has not been made available for businesses to profit from unpaid labor.*

2. Re-distribution of this software on for-profit, public use, repository hosting sites (for example: Github) is permitted provided no fees are charged specifically to access this software.

3. **This software is provided on an as-is basis and may only be used at your own risk.** This software is the product of a single individual's recreational project. The author does not have the resources to perform the degree of code review, testing, or other verification required to extend any assurances that this software is suitable for any purpose, or to offer any assurances that it is safe to execute without causing data loss or other damage.

4. **This software is intended for experimental use in situations where data loss (or any other undesired behavior) will not cause unacceptable harm.** Users with critical data safety needs must not use this software and, instead, should use equivalent tools that have a proven track record.

5. If this software is redistributed, this copyright notice and license text must be included without modification.

6. Distribution of modified copies of this software is discouraged but is not prohibited. It is strongly encouraged that fixes, modifications, and additions be submitted for inclusion into the main release rather than distributed independently.

7. This software reverts to the public domain once 10 years elapses after its most recent update or immediately upon the death of its author, whichever happens first.