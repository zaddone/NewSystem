package main
import(
	"fmt"
	_ "github.com/zaddone/NewSystem/server"
)
func main(){
	fmt.Println("start new system")
	var cmd string
	for {
		i,e := fmt.Scanf("%s",&cmd)
		fmt.Println(i,e,cmd)
	}
}
