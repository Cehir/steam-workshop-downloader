# steam-workshop-downloader
A client that uses steamcmd and configuration files to download mods from the steam workshop.

## Requirements
- [steamcmd](https://developer.valvesoftware.com/wiki/SteamCMD)

## Usage
### Create configuration file
Create a configuration file in your home directory or your current working directory. 
The configuration can be either in JSON ([json example](examples/mac_os.json)) or YAML ([yaml example](examples/mac_os.yaml)) format.

Default is `.steam-workshop-downloader.yaml` in your home directory.

### Run steam-workshop-downloader
Run the steam-workshop-downloader with the path to your configuration file as a named argument.

    $ steam-workshop-downloader download --config /path/to/config.yaml
