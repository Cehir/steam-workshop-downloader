package steamcmd

import "regexp"

// extractPathRegex extracts the path from the steamcmd output
// download regex input example: Downloaded item 2169435993 to "C:\steamcmd\steamapps\workshop\content\108600\2169435993" (31729 bytes)
// will return C:\steamcmd\steamapps\workshop\content\108600\2169435993
var extractPathRegex = regexp.MustCompile(`Downloaded item \d+ to "(.+)" \(\d+ bytes\)`)

// appIDRegex extract workshop id from path
// example: C:\steamcmd\steamapps\workshop\content\108600\2169435993
// will return 108600
var appIDRegex = regexp.MustCompile(`content\\(\d+)\\`)
