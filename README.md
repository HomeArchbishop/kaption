# Kaption

Kaption is a tool designed to generate captions for videos. It uses local speech recognition services instead of uploading your files to remote servers.

Kaption starts a local WS server, receives video streams in the agreed format, and generates caption information to return

**Requirements**
- Windows 64 bit (Support on Linux/Darwin is under development)
- At least 8GB of RAM (recommended)

## Build

For quick start, download a pre built release and skip this part.

**Step 1**

```sh
git clone https://github.com/HomeArchbishop/kaption.git
cd kaption
```

**Step 2**

- Download a certain voice-recognition model from [vosk - model list](https://alphacephei.com/vosk/models#:~:text=modified%20in%20runtime.-,Model%20list,-This%20is%20the)
- Unzip downloaded model
- Copy the unzipped model into project directory and rename it as "model"

Your project directory should be like
```
kaption/
├── model/
│   ├── conf/
│   ├── am/
│   └── ...
├── ...
└── README.md
```

**Step 3** Download go modules.

```sh
set GO111MODULE=on
go mod vendor
```

**Step 4** Build

```sh
sh scripts/build.sh
```

Check built files in `/dist`

## Usage

> If you download from release, please download a model and add it into the built Kaption app. **See "Build (Step 2)" on this page**.

With a built Kaption app, simply click `start.exe`. A WS server will listen on port 8080 (by default).

Otherwise, run either of the commands below to customize your port

```sh
# same as starting with click, default port 8080
./start.exe

# use port 9090
./start.exe -port=9090
```

After the model is successfully loaded, the interface will give a tip of server startup.

Then connect your client to the local server and send messages to it. The server will response the caption infomations immediately upon recognizing something. A valid message flow can be as below:

```
[↑] Binary Message
[↑] Binary Message
[↑] Binary Message
[↑] Binary Message
[↓] Text {
      "End":{"Index":2,"Time":2.27},
      "Start":{"Index":4,"Time":0.85},
      "Text":"Hello world"
    }
[↑] Binary Message
[↑] Binary Message
... ............
```

For specific message exchange formats, please refer to [Reference: message format](#reference:-message-format)

## Reference: message format

Take .m3u8/.ts (hls) video media as an example.

- **video hash**: The entire video will be cut into a large number of `.ts` files. These `.ts` files of the same video must have a unique `video hash` for Kaption to recognize them as one video. **It is 16-byte-long such as "p077ugzwtafk5sej"**

- **file hash**: Each `.ts` file must have a unique hash to distinguish it from other `.ts` files. **It is 16-byte-long such as "thlh3iqy9az54l49"**

- **file index**: Each `.ts` file has an increasing number starting from 0 for Kaption to handle in order. Note that this index is not related to the order in the `.m3u8` file (although it is usually consistent), you can completely number the 51st to 60th in `.m3u8` as `file index: 0-9`. **It must start from 0 and increase sequentially, otherwise it will cause kaption blocking until its waiting 'next file index' is received. Pad it with '0' on the left to a 16 byte string and add it to the beginning of the binary just like the previous 2 hash**
```
# Message is in binary format

00000000: 7030 3737 7567 7a77 7461 666b 3573 656a  (16 bytes video hash)
00000001: 7468 6c68 3369 7179 3961 7a35 346c 3439  (16 bytes file hash)
00000002: 3030 3030 3030 3030 3030 3030 3030 3030  (16 bytes file index)
00000003: 4740 1110 0042 f025 0001 c100 00ff 01ff  (file binary)
00000004: ...
```

Then each response is a JSON text:
```js
{
  "Start": {"Index": 2, "Time": 2.27}, // start at 2.27s of slice (file index:2)
  "End": {"Index": 4, "Time": 0.85}, // end at 0.85s of slice (file index:4)
  "Text": "Hello world" // caption raw text
}
```

A valid message flow can be as below:

```
[↑] Binary Message
[↑] Binary Message
[↑] Binary Message
[↑] Binary Message
[↓] Text {"End":{"Index":4,"Time":0.85},"Start":{"Index":2,"Time":2.27},"Text":"Hello world"}
[↑] Binary Message
[↑] Binary Message
... ............
```

# Licenses

Kaption is licensed under the [MIT License](LICENSE).

Kaption includes the following third-party libraries:

- [vosk](https://github.com/alphacep/vosk-api): Licensed under the Apache-2.0 license. No changes has been made.
- [websocket](https://github.com/gorilla/websocket): Licensed under the BSD 2-Clause "Simplified" License.
- [ffmpeg](https://www.ffmpeg.org): Licensed under the LGPL license
