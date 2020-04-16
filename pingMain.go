package Main

import (
	"flag"
	"fmt"
)
func main() {
	message := string("Allows the user to select the maximum amount of the pings")
	limit := flag.Int("l",0,message)
	flag.Parse() //Get host name/link

	if flag.NArg()==0 {
		flag.Usage()
		return
	}

	hostname := flag.Arg(0)
	pingHandler,err := ping(hostname,*limit)
	if err != nil{
		fmt.Print("Encountered Error: ", err.Error())
	}
	pingHandler.start()





}

