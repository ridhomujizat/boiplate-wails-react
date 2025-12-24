//go:build darwin || linux

package assets

import _ "embed"

//go:embed icon.png
var IconData []byte
