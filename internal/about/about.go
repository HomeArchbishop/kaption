package about

import (
	"fmt"
)

func PrintAbout(version string) {
	fmt.Printf(`Kaption %s
Kaption is licensed under the MIT License.
Kaption uses the following open source libraries:
- vosk:      Licensed under the Apache-2.0 license. No changes have been made. (https://github.com/alphacep/vosk-api)
- websocket: Licensed under the BSD 2-Clause "Simplified" License. (https://github.com/gorilla/websocket)
- ffmpeg:    Licensed under the LGPL license. (https://www.ffmpeg.org)
`, version)
}
