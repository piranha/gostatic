package hotreload

import _ "embed"

//go:embed assets/morphdom.js
var Morphdom []byte

//go:embed assets/hotreload.js
var Script []byte
