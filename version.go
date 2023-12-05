package battery

import "fmt"

const (
	version = "0.0.0"
)

var logo = `
########     ###    ######## ######## ######## ########  ##    ## 
##     ##   ## ##      ##       ##    ##       ##     ##  ##  ##  
##     ##  ##   ##     ##       ##    ##       ##     ##   ####   
########  ##     ##    ##       ##    ######   ########     ##    
##     ## #########    ##       ##    ##       ##   ##      ##    
##     ## ##     ##    ##       ##    ##       ##    ##     ##    
########  ##     ##    ##       ##    ######## ##     ##    ##  
game sever framework @v%s
`

func GetLOGO() string {
	return fmt.Sprintf(logo, Version())
}

func Version() string {
	return version
}
