# tunesbus
An experimental MPRIS integration for iTunes (running on Wine)

## How to run
First, open iTunes. Then, open this program. This should be all.

This was tested on wine-11.12 (Staging), iTunes 12.12.0.6, 64-bits, for which you can download [here](https://appledb.dev/firmware/iTunes/12.12.0.6.html). The latest version doesn't install on Wine at all, but if it ever does, probably this program *should* work (assuming that nothing was changed on iTunes' COM interface)  

WineHQ report for this version: https://appdb.winehq.org/objectManager.php?sClass=version&iId=31322  
Recommended winetricks verbs: `dxvk windowmanagerdecorated=n fontsmooth=rgb`. Do not install anything related to WShell.Host, this program doesn't need it and Wine's OLE will break if you do so.

## Compiling
Fortunately, you can cross-compile for Windows/Wine (the intended target) without any issues:
```bash
GOOS=windows GOARCH=amd64 go build -o tunescomtest.exe cmd/main.go 
```
The modified godbus/dbus code used for this project is not avaliable (yet), so you might be able to compile it, but the program won't be able to connect to the dbus socket (since this library doesn't compile the relevant code for that on Windows targets and it only supports dbus connections with Unix file descriptors, something we don't have on Wine)  
Note that this program is only supposed to be ran on a Wine version that supports AF_UNIX sockets (the latest wine-staging should do it)

## Known issues
There is no actual release for this yet, but these are the current issues:
* tunesbus will "stop" listening to events if you executed it twice in the same iTunes session (tunesbus isn't properly releasing COM objects)
* Closing iTunes normally (by clicking the X button) will take *little* a bit of time and then it'll ask you if you really want to close it (*something about programs still using the "scripting interface" do you really really want to close me...?*). That happens for the same reason above, programs need to clean up  after receiving the `OnAboutToPromptUserToQuit` event as soon as possible
* iTunes might randomly stop working. Not specifically an issue with this program, but iTunes itself is known for being unstable¹, even on Windows. Using this might worsen stability by some degree, though.
* If there is any other "issue", that "issue" is either because I haven't implemented the relevant stuff to it yet or I just forgot to add it to this list (it's an actual issue, then)

> ¹ Try not to use the "For You" or anything that uses its embedded webview *too much* :)