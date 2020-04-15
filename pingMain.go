package Main

import (
	"flag"
	"os"
)
func main() {
	message := string("Allows the user to select the maximum amount of the pings")
	flag.Int("l",0,message)
	flag.Parse() //Get host name/link

	if flag.NArg()==0 {
		flag.Usage()
		return
	}

	hostname := flag.Arg(0)


}

