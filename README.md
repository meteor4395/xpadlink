<p align="center"><img src="app/assets/logo.png" width="120" alt="XPadLink logo"></p>

# XPadLink

[![CI](https://github.com/josuebrunel/xpadlink/actions/workflows/ci.yml/badge.svg)](https://github.com/josuebrunel/xpadlink/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/josuebrunel/xpadlink)](https://github.com/josuebrunel/xpadlink/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Use an Android phone, or a gamepad plugged into it, as a gamepad on your Linux or Windows PC.

```
phone (Flutter app) --UDP, 60 Hz--> xpadlinkd (Go) --> virtual Xbox 360 pad
                                                   Linux: /dev/uinput
                                                   Windows: ViGEmBus
```

## Contents

- [Features](#features)
- [Screenshots](#screenshots)
- [Quick start](#quick-start)
  - [1. Server](#1-server)
    - [Windows setup](#windows-setup)
    - [Build on Windows](#build-on-windows)
  - [2. App (Android)](#2-app-android)
  - [3. Play](#3-play)
- [Bluetooth](#bluetooth)
- [Vibration and motion](#vibration-and-motion)
  - [Vibration](#vibration)
  - [Motion (gyro) in Cemu](#motion-gyro-in-cemu)
- [Calibration](#calibration)
- [Run in the background](#run-in-the-background)
- [Troubleshooting](#troubleshooting)
- [Reference](#reference)
  - [Flags](#flags)
  - [Known gaps](#known-gaps)
- [Development](#development)
  - [Wire format](docs/PROTOCOL.md)
- [License](#license)

## Features

- A virtual Xbox 360 pad (`045e:028e`), so SDL, Steam, Proton, Cemu and browsers work without mapping.
- Touch controls on the phone, or a real controller plugged into it.
- Over Wi-Fi, or over Bluetooth when there is no shared network.
- The app finds your PC on Wi-Fi by itself. No typing an IP.
- Up to 4 phones at once, one virtual pad each.
- A token keeps other people on your network from sending input.
- Vibration: a game's rumble buzzes the phone, or a controller with motors.
- Calibration for analog triggers (Hall effect pads), stick center and deadzone, and the phone's gyro.
- Motion for Cemu (experimental): the phone's tilt, for Wii U GamePad games.
- Runs as a background service on Linux, macOS and Windows.

## Screenshots

<table>
  <tr>
    <td><img src="docs/images/app-connect-tabs.png" alt="Connect screen with Saved, Wi-Fi and Bluetooth tabs"></td>
    <td><img src="docs/images/app-controller-pad.png" alt="Touch controller with the Back button, player ring and link chip"></td>
  </tr>
  <tr>
    <td align="center">Pick a saved server, find your PC on Wi-Fi, or choose a paired one over Bluetooth</td>
    <td align="center">Touch controls, with the link and ping on top. Back leaves the view</td>
  </tr>
  <tr>
    <td><img src="docs/images/app-controller-hidden.png" alt="Hidden controls view with the logo and live stick values"></td>
    <td><img src="docs/images/hardwaretester.png" alt="Browser gamepad tester showing XPadLink"></td>
  </tr>
  <tr>
    <td align="center">The eye hides the touch controls (they hide by themselves when a real pad is used)</td>
    <td align="center">The PC sees a standard gamepad</td>
  </tr>
  <tr>
    <td colspan="2"><img src="docs/images/cemu-input.png" alt="Cemu input settings with the pad mapped"></td>
  </tr>
  <tr>
    <td colspan="2" align="center">Cemu maps it with Auto-assign, no manual setup</td>
  </tr>
</table>

## Quick start

You need three things: the server on your PC, the app on your phone, and both on the same Wi-Fi.

### 1. Server

Download the archive for your PC from the
[Releases](https://github.com/josuebrunel/xpadlink/releases) page. Or build it (Go 1.22+):

**Linux:** choose `xpadlinkd-linux-amd64.tar.gz` (or `arm64`) and unpack it. Set up uinput once so the daemon can create virtual pads without root:

```sh
git clone https://github.com/josuebrunel/xpadlink && cd xpadlink
make build            # produces server/xpadlinkd
make uinput-setup     # loads uinput, installs the udev rule, adds you to the input group
make firewall         # only if you use ufw: opens UDP 9777
```

Start the daemon. It prints the token and the addresses to use:

```sh
server/xpadlinkd
```

**Windows:** use the `xpadlinkd-windows-amd64.zip` release. Install
[ViGEmBus](https://github.com/nefarius/ViGEmBus/releases). The XPadLink GitHub release also
provides `xpadlinkd-windows-amd64.exe` and `ViGEmClient.dll` as separate downloads; place
both in the same folder. The DLL is built from the official ViGEmClient source by CI. The
ViGEmBus installer is not redistributed; download it from the official link above.

Follow [Windows setup](#windows-setup) to install and start it. The daemon prints a token
and LAN address to enter in the Android app. Allow the executable through Windows Defender
Firewall on private networks (UDP port 9777). Windows currently supports Wi-Fi; Bluetooth
connections are Linux-only.

#### Windows setup

**Requirements:** a 64-bit Windows PC, the ViGEmBus driver, and both XPadLink files in one
folder. Use the `windows-amd64` XPadLink executable and x64 `ViGEmClient.dll`. The
ViGEmClient project is retired and its GitHub releases do not provide a prebuilt DLL.

1. Download and install **ViGEmBus** from the [official releases](https://github.com/nefarius/ViGEmBus/releases). Restart Windows if its installer asks.
2. Download `xpadlinkd-windows-amd64.exe` and `ViGEmClient.dll` from the same XPadLink release. Put them together in a folder such as `C:\XPadLink`. The ViGEmClient license is included in the release as `ViGEmClient-LICENSE.txt`.
3. Open PowerShell in that folder and start the server:

   ```powershell
   .\xpadlinkd-windows-amd64.exe -bt=false
   ```

4. If Windows Firewall prompts, allow `xpadlinkd-windows-amd64.exe` on **Private networks**. Otherwise, add an inbound firewall rule for UDP port `9777`.
5. Copy the token and IP address printed in the PowerShell window. Keep the server running.
6. On the Android phone, join the same Wi-Fi, open XPadLink, select the PC on the **Wi-Fi** tab (or enter its printed `IP:9777` address), enter the token, and connect.
7. Test input on Windows with **Win+R**, `joy.cpl`, select the Xbox 360 controller, then **Properties**. Press app buttons and move the sticks; they should register.

The virtual pad appears when the phone connects and disappears when it disconnects. Keep both XPadLink files together whenever starting the server.

#### Build on Windows

To build locally instead of downloading the release, install Go 1.22+, Git, CMake 3.20+, and Visual Studio 2022 or Build Tools 2022 with the **Desktop development with C++** workload. Install ViGEmBus separately as above.

Build the XPadLink server executable from the repository. This writes the ignored
`server\xpadlinkd-windows-amd64.exe`:

```powershell
cd C:\path\to\xpadlink\server
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -trimpath -ldflags="-s -w" -o .\xpadlinkd-windows-amd64.exe .\cmd\xpadlinkd
```

Build the matching client DLL from the pinned official ViGEmClient source. The export flags are required so the DLL provides `vigem_alloc`:

```powershell
cd $env:TEMP
git clone https://github.com/nefarius/ViGEmClient.git
cd .\ViGEmClient
git checkout b66d02d57e32cc8595369c53418b843e958649b4
cmake -S . -B build -G "Visual Studio 17 2022" -A x64 `
  -DViGEmClient_DLL=ON `
  -DCMAKE_CXX_FLAGS="/DVIGEM_DYNAMIC /DVIGEM_EXPORTS"
cmake --build build --config Release --target ViGEmClientShared --clean-first
dumpbin /exports .\build\Release\ViGEmClient.dll | findstr vigem_alloc
```

The last command should list `vigem_alloc`. Copy the DLL beside the XPadLink executable:

```powershell
Copy-Item .\build\Release\ViGEmClient.dll C:\path\to\xpadlink\server\
```

If the Visual Studio generator is unavailable, install the C++ workload. With an x64 Native Tools Command Prompt and NMake, use `-G "NMake Makefiles"` and omit `-A x64`.

The locally built executable, DLL, and optional ViGEmBus installer are ignored by Git. The XPadLink executable and DLL are attached individually to tagged XPadLink releases by CI; the ViGEmBus installer remains an upstream download.

### 2. App (Android)

Download `xpadlink.apk` from the [Releases](https://github.com/josuebrunel/xpadlink/releases) page and install it
(allow "install unknown apps" for your browser or file manager when Android asks).

Or build it yourself. You need the [Flutter SDK](https://docs.flutter.dev/get-started/install) and a phone with USB debugging:

```sh
make app              # generates app/xpadlink_app (code, icon, permissions)
make app-run          # builds and runs on the connected phone
```

### 3. Play

1. Open the app. The connect screen has three tabs: **Saved**, **Wi-Fi** and **Bluetooth**. Servers on your Wi-Fi appear under the **Wi-Fi** tab.
2. Tap yours, enter the token, press **Connect**. **Save** keeps it for next time, and it then shows up under **Saved**. No Wi-Fi? See [Bluetooth](#bluetooth).
3. Use the touch controls, or plug in a controller. The touch controls hide when a controller is used.

Check it on the PC: `evtest` (pick `XPadLink - Xbox Controller 1`) or open
<https://hardwaretester.com/gamepad>.

In the app's top bar: the eye hides the touch controls, the rotation icon turns motion on or off, the tune icon opens [calibration](#calibration), and the flip icon turns the screen upside down if it ended up the wrong way.

## Bluetooth

The phone can reach the PC over Bluetooth instead of Wi-Fi. Wi-Fi and Bluetooth phones can play at the same time. It needs Linux with BlueZ (`bluetoothd`, which every desktop has).

1. Pair the phone with the PC once, in the Bluetooth settings of both.
2. Start `xpadlinkd`. It logs `bluetooth listening` (turn it off with `-bt=false`).
3. In the app, open the **Bluetooth** tab and press **Show paired devices** (Android 12 and later asks for the Bluetooth permission). Tap your PC, enter the token, press **Connect**. **Save** keeps it under **Saved**.

Same controls, vibration and motion as on Wi-Fi. If the connect fails, run `bluetoothctl trust <phone-address>` on the PC, and check that `bluetoothctl show` says `Powered: yes`.
If the PC doesn't show up when you pair from the phone, run `bluetoothctl pairable on` and `bluetoothctl discoverable on` on the PC first, or start the pairing from the PC.

## Vibration and motion

### Vibration

A game's rumble goes to the phone, or to your controller if it has motors. Nothing to set up. It is sent as media vibration, so it works even when "touch vibration" is off in the phone settings.
Check it without a game: connect the phone, then run `go run ./cmd/xpadrumble` in `server/`.

### Motion (gyro) in Cemu

**Experimental.** Some Wii U games need the GamePad's motion sensors, like the maze shrines in Breath of the Wild.
The app sends the phone's gyro and accelerometer, and `xpadlinkd` serves them to Cemu as a DSU (Cemuhook) server on `127.0.0.1:26760`.

1. In Cemu, open **Options, Input settings** and set the emulated controller to **Wii U GamePad**.
2. Add a controller with the API **DSUController**, server `127.0.0.1`, port `26760`, and the controller index matching your player number minus one.
3. Hold the phone landscape, like a GamePad, and press **Calibrate** in Cemu.

No gyroscope on the phone? Many budget phones have none. The app then derives the turn rate from the rotation vector (accelerometer plus magnetometer). It's noisier and drifts near magnets.
On such a phone the BOTW maze shrine is not working yet (see [Known gaps](#known-gaps)).
Use `-dsu ""` to turn the motion server off.

## Calibration

Open it with the tune icon in the controller screen's top bar. It has three tabs, and each one has its own **Save** and **Reset**.
Values are stored on the phone, per controller name, and applied in the app before anything is sent. The server and the wire format don't change.

- **Triggers.** Pull both triggers all the way, then let go. The app learns the lowest and highest value your pad reports and rescales that range to 0..255. This matters for Hall effect triggers, which rarely reach 100%, and for racing games that use them as throttle and brake. **Start over** clears the readings if you pulled badly.
- **Sticks.** With both sticks at rest, press **Set center**, then spin each stick around the edge. The app saves the resting center and how far each axis reaches. The deadzone slider (8% by default) is applied after that.
- **Motion.** Lay the phone flat and still, then press **Measure gyro**. Two seconds of readings become the gyro zero point, so the cursor stops drifting. If tilting the phone left moves the cursor right in Cemu, switch on **Flip roll**. Motion settings belong to the phone, not to a controller.

Triggers and sticks need a physical controller that has sent some input. Press a button first. The touch controls are not calibrated.

A pad that reports an analog trigger axis is read as analog. The L2/R2 key is only used as an on/off fallback for pads with no analog trigger axis.

## Run in the background

`xpadlinkd install` registers a per-user service that starts at login and restarts on failure.
The token is saved once and reused, so the phone never needs a new one.

| | Linux | macOS | Windows |
|---|---|---|---|
| Install | `xpadlinkd install` | `xpadlinkd install` | `xpadlinkd install` |
| Mechanism | systemd user unit | launchd LaunchAgent | scheduled task at logon |
| Logs | `journalctl --user -u xpadlinkd -f` | `~/Library/Logs/xpadlinkd.log` | Task Scheduler history |
| Stop and remove | `xpadlinkd uninstall` | `xpadlinkd uninstall` | `xpadlinkd uninstall` |
| Token file | `~/.config/xpadlink/token` | `~/Library/Application Support/xpadlink/token` | `%AppData%\xpadlink\token` |

- Pass server flags after `install`: `xpadlinkd install -max 2 -log-level debug`.
- Show the token any time: `xpadlinkd token`.
- Linux: to start at boot instead of at login, run `loginctl enable-linger $USER`.
- Keep the binary where it is: the service points at its current path. Run `install` again after moving it.
- macOS can run the daemon with `-dry` but cannot create virtual pads.
- Windows virtual pads require the amd64 build, ViGEmBus, and an x64 `ViGEmClient.dll`. The two XPadLink files are separate release downloads; obtain the driver installer from ViGEmBus upstream. Use `-dry` to run without the virtual pad dependencies.

## Troubleshooting

| Problem | Fix |
|---------|-----|
| App says **Wrong token** | Run `xpadlinkd token` and re-enter it |
| App says **Server full** | Raise `-max`, or disconnect another phone |
| No servers under the Wi-Fi tab | Same Wi-Fi subnet? Guest networks block it. Type `ip:9777` instead. Check the firewall for UDP 9777 |
| `permission denied: /dev/uinput` | Run `make uinput-setup`, then log out and in |
| Windows does not show a controller | Check that ViGEmBus is installed, both XPadLink files are beside each other, and the x64 DLL exports `vigem_alloc`; use the Windows amd64 build |
| Linux game does not see the pad | `evtest` should list `XPadLink - ...`. Cemu may show the generic SDL name "Xbox 360 Controller" |
| Plugged-in controller not detected | A USB-C controller needs the phone's only port. Use wireless adb: `adb tcpip 5555 && adb connect <phone-ip>:5555`, then unplug the cable |
| Triggers act like on/off buttons | Open Calibration, pull each trigger fully on the **Triggers** tab and save. If it still jumps to full, open an issue with the controller model and the app log |
| Stick drifts or never reaches full | Calibrate it on the **Sticks** tab, or raise the deadzone |
| Screen is upside down | Press the flip icon in the app's bottom bar |
| Cemu shows no motion | `ss -ulnp \| grep 26760` should list `xpadlinkd`, and the daemon log should say `dsu: client subscribed` once the DSUController is added. `motion unavailable` in the app log means the phone has no gyroscope and no rotation vector sensor. Check the rotation icon is on |
| Motion is wrong in a game | Press **Calibrate** in Cemu with the phone held steady. In the app, measure the gyro on the **Motion** tab, and try **Flip roll** if left and right are swapped. Open an issue with the game and what happens |
| Bluetooth connect fails | Pair first. Then `bluetoothctl trust <phone-address>`. `xpadlinkd` must log `bluetooth listening`; if it says `bluetooth disabled`, `bluetoothd` is not running |
| Input lags or drops | Watch the app log: `adb logcat -v time xpadlink:V flutter:V '*:S'`. A `link:` line every 5 s shows acks and send failures |

A Bluetooth controller paired to the phone should work the same as a USB one (Android lists it as a gamepad), but it is untested.

## Reference

### Flags

| Flag | Default | Meaning |
|------|---------|---------|
| `-listen` | `:9777` | UDP address |
| `-token` | saved token | Shared secret. Order: flag, `$XPADLINK_TOKEN`, token file (created on first run) |
| `-max` | `4` | Max simultaneous controllers |
| `-timeout` | `3s` | Drop a client after this much silence |
| `-bt` | on | Accept phones over Bluetooth (Linux, needs `bluetoothd`) |
| `-dsu` | `127.0.0.1:26760` | DSU motion server for Cemu, empty to disable |
| `-log-level` | `info` | `debug`, `info`, `warn`, `error` |
| `-dry` | off | Log input instead of creating pads (implies `debug`) |

Test without a phone: `xpadlinkd -token abc &` then `go run ./cmd/xpadsim -token abc` in `server/`.

### Known gaps

- Linux virtual pads use uinput. Windows virtual pads use ViGEmBus and currently require Windows amd64; macOS supports dry-run only.
- Motion is experimental. It reaches Cemu, but in the BOTW maze shrine the maze still flips upside down (tested on a phone without a gyroscope, #32). It is only wired to Cemu (DSU).
- Wi-Fi is LAN only with no encryption: the token only stops other people on the network from injecting input. Bluetooth needs pairing, and the link is encrypted by Bluetooth itself.
- Bluetooth is Linux only, and untested with more than one phone at once.
- Calibration covers physical controllers only, and it has no screenshots in this README yet.
- Controllers that put the right stick on RX/RY instead of Z/RZ need a small change in `app/kotlin/MainActivity.kt`.

## Development

```sh
make help       # all targets
make test       # Go tests
make dry        # run the server without uinput
```

Check vibration without a game: connect the phone, then run `go run ./cmd/xpadrumble` in `server/`.
It plays a rumble effect on the virtual pad, the way a game would.

- CI (`.github/workflows/ci.yml`): tests on Linux, macOS and Windows, builds `xpadlinkd` for
  Linux and macOS (amd64, arm64) and Windows amd64 with [GoReleaser](https://goreleaser.com)
  (`.goreleaser.yaml`). The Windows job builds the ViGEmClient DLL from a pinned source
  revision. Pushing a `v*` tag publishes archives, checksums, and separate Windows executable,
  DLL, and client-license assets. The ViGEmBus installer is linked upstream, not redistributed.
- Wire format: [docs/PROTOCOL.md](docs/PROTOCOL.md).
- Layout: `server/` Go daemon (`proto`, `hub`, `bt`, `vpad`, `dsu`, `cmd/`), `app/` Flutter code (`lib/`: screens in `main.dart` and `onscreen.dart`, colours and shared widgets in `theme.dart`, Wi-Fi and Bluetooth links in `link.dart`) and Kotlin (`kotlin/`: gamepad and sensors in `MainActivity.kt`, the Bluetooth connection in `BtLink.kt`), copied into a generated project by `app/setup.sh`.

## License

[MIT](LICENSE). The bundled fonts (Chakra Petch, Barlow, JetBrains Mono) are under the SIL Open Font License, see `app/assets/fonts/`.
