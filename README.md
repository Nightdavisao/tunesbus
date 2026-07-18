# tunesbus
An experimental MPRIS integration for iTunes (running on Wine, Linux)

[Main repository](https://codeberg.org/Nightdavisao/tunesbus)  
[Releases](https://codeberg.org/Nightdavisao/tunesbus/releases)  

Pull requests are disabled on the GitHub mirror. I'm lazy to set up something to pull commits from it and this kinda demotivates the average slop machine to "contribute" to this project, anyway. 

## How to run
First, open iTunes. Then, open this program. This should be all.  
You can also just open this program and it'll open iTunes anyway. That's just how Windows' OLE/COM works, I didn't write any code to do this. :)

### How to run iTunes itself
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
There is no actual release for this yet, but these are the current issues:
* tunesbus will "stop" listening to events if you executed it twice in the same iTunes session (unsure why it does that, polling the metadata will have to be implemented at some point if I can't find any solution to this)
* Closing iTunes normally (by clicking the X button) will take *a little* bit of time and then it'll ask you if you really want to close it (*something about programs still using the "scripting interface" do you really really want to close me...?*). That happens for the same reason above, programs need to clean up after receiving the `OnAboutToPromptUserToQuit` event as soon as possible
* iTunes might randomly stop working. Not specifically an issue with this program, but iTunes itself is known for being unstable, even on Windows. Using this might worsen stability by some degree, though.
* If there is any other "issue", that "issue" is either because I haven't implemented the relevant stuff to it yet or I just forgot to add it to this list (it's an actual issue, then)