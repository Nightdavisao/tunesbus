# tunesbus
An experimental MPRIS integration for iTunes (running on Wine, Linux)

* [Nightly builds](https://nightly.link/Nightdavisao/tunesbus/workflows/build/main/tunesbus)
* [Releases](https://codeberg.org/Nightdavisao/tunesbus/releases)  
* [Codeberg repository](https://codeberg.org/Nightdavisao/tunesbus)  
* [GitHub mirror](https://github.com/Nightdavisao/tunesbus)  

## How to run
First, open iTunes. Then, open this program. This should be all. You can also just open this program and it'll open iTunes anyway. 

tunesbus will create a configuration file in the executable path if there's no existing one. Currently, the options avaliable for configuration are only related to how the MRPIS server will identify itself to clients. You might want to change the `BusNameSuffix` value to `cider`, if you want it to be picked up by Music Presence, for example.

### How to run iTunes itself?
Refer to [this page](./docs/iTunes.md).

## Cloning and compiling
Note that this repository has **submodules**. You can clone it this way:
```bash
git clone --recurse-submodules https://codeberg.org/Nightdavisao/tunesbus.git
```

Fortunately, you can cross-compile for Windows/Wine (the intended target) without any issues:
```bash
GOOS=windows GOARCH=amd64 go build -o tunesbus.exe ./cmd/
```
Note that this program is only supposed to be ran on a Wine version that supports **AF_UNIX sockets** (the latest wine-staging should do it)

## Known issues
There is no actual stable release for this yet, but these are the current issues:
* Closing iTunes normally (by clicking the X button) will take *a little* bit of time and then it'll ask you if you really want to close it (*something about programs still using the "scripting interface" do you really really want to close me...?*). Note to self: Reference count for the last OLE releaser is always 3 for some reason.
* iTunes might randomly stop working. Not specifically an issue with this program, but iTunes itself is known for being unstable, even on Windows. Using this might worsen stability by some degree, though.