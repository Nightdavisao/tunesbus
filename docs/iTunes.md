# Getting iTunes to run on Wine, Linux
The latest version doesn't install on Wine at all, so you need to install an older version.  

Current tested versions (**wine-11.13 (Staging)**)  
* [iTunes 12.12.0.6, 64 bits](https://appledb.dev/firmware/iTunes/12.12.0.6.html)
* [iTunes 12.12.6.1, 64 bits](https://appledb.dev/firmware/iTunes/12.12.6.1.html)

WineHQ report for 12.x versions *(some information is outdated)*: https://appdb.winehq.org/objectManager.php?sClass=version&iId=31322

Use the last version listed here if you want to change the language/locale of iTunes for some specific reason (in my special case, it's because I want to avoid localized metadata on my library), the first one doesn't work with any language other than English (United States). Set the `LC_ALL` env to something of your choice, such as `pt_BR.UTF-8` before running iTunes. Only changing the language on the preferences won't make any related web feature use the chosen language.

These are the Winetricks verbs that your prefix needs: `dxvk corefonts cjkfonts windowmanagerdecorated=n fontsmooth=rgb`

`gdiplus` is also needed here, though the latest version of Winetricks pulls a 32-bit version of this DLL, which is obviously incompatible with iTunes. So you need to find it yourself on the wide internet, preferrably pick up a Windows 10 version.

Do not install the `wsh57` verb, this program doesn't need it and Wine's OLE will break if you do so.